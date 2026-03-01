package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gmb/internal/config"
	"gmb/internal/instagram"
	"gmb/internal/state"
	"gmb/internal/telegram"
)

type fakeStore struct {
	sent map[string]struct{}
}

func (f *fakeStore) ListSentIDs(context.Context) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	for k := range f.sent {
		out[k] = struct{}{}
	}
	return out, nil
}
func (f *fakeStore) AppendSentID(_ context.Context, id string) error {
	if f.sent == nil {
		f.sent = map[string]struct{}{}
	}
	f.sent[id] = struct{}{}
	return nil
}
func (f *fakeStore) ResetSent(context.Context) error {
	f.sent = map[string]struct{}{}
	return nil
}

type fakeInstagram struct{ posts []instagram.Post }

func (f *fakeInstagram) ListProfilePosts(context.Context, string, int) ([]instagram.Post, error) {
	return f.posts, nil
}

type fakeNotifier struct {
	called      int
	lastCaption string
}

func (f *fakeNotifier) SendVideo(_ context.Context, req telegram.SendRequest, _ io.Reader, _ string) (telegram.SendResult, error) {
	f.called++
	f.lastCaption = req.Caption
	return telegram.SendResult{MessageID: 1}, nil
}

type fakeRoundTripper struct{}

func (f fakeRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("mp4-bytes")),
		Header:     make(http.Header),
	}, nil
}

func TestRunProductionResetsWhenAllSent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gmbadass_20260301-abc.mp4")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := &fakeStore{sent: map[string]struct{}{"gmbadass_20260301-abc": {}}}
	n := &fakeNotifier{}
	s := &Service{
		Cfg: config.Config{
			OneShot:              false,
			InstagramProfile:     "gmbadass",
			DescriptionRegex:     "(?i)good",
			InstagramScanLimit:   10,
			DownloadDir:          dir,
			DownloadMaxPerRun:    1,
			DownloadDelaySeconds: 0,
			TelegramChatID:       "123",
		},
		Store:     st,
		Instagram: &fakeInstagram{},
		Notifier:  n,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Sleep:     func(time.Duration) {},
	}
	if err := s.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if n.called != 1 {
		t.Fatalf("expected send called once, got %d", n.called)
	}
	if n.lastCaption != "gmbadass - 2026-03-01" {
		t.Fatalf("unexpected caption: %q", n.lastCaption)
	}
}

func TestBuildCaptionFromPath(t *testing.T) {
	got := buildCaptionFromPath("gmbadass", "/data/video/gmbadass_20260301-abc.mp4")
	if got != "gmbadass - 2026-03-01" {
		t.Fatalf("unexpected caption: %q", got)
	}
}

var _ state.Store = (*fakeStore)(nil)
var _ instagram.Client = (*fakeInstagram)(nil)
var _ telegram.Client = (*fakeNotifier)(nil)
