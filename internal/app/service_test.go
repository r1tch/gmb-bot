package app

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"gmb/internal/config"
	"gmb/internal/state"
	"gmb/internal/telegram"
	"gmb/internal/tiktok"
)

type fakeStore struct {
	lastSent string
	setCalls int
}

func (f *fakeStore) GetLastSentID(context.Context) (string, error) { return f.lastSent, nil }
func (f *fakeStore) SetLastSentID(_ context.Context, id string) error {
	f.lastSent = id
	f.setCalls++
	return nil
}

type fakeTikTok struct {
	videos       []tiktok.Video
	downloadErr  error
	downloadBody io.ReadCloser
}

func (f *fakeTikTok) ListLatestVideos(context.Context, string, int) ([]tiktok.Video, error) {
	return f.videos, nil
}
func (f *fakeTikTok) DownloadVideo(context.Context, tiktok.Video) (io.ReadCloser, string, error) {
	if f.downloadErr != nil {
		return nil, "", f.downloadErr
	}
	if f.downloadBody != nil {
		return f.downloadBody, "video/mp4", nil
	}
	return io.NopCloser(strings.NewReader("video-bytes")), "video/mp4", nil
}

type fakeNotifier struct {
	result telegram.SendResult
	err    error
	req    telegram.SendRequest
	gotNil bool
}

func (f *fakeNotifier) SendVideoOrFallback(_ context.Context, req telegram.SendRequest, video io.Reader, _ string) (telegram.SendResult, error) {
	f.req = req
	f.gotNil = video == nil
	return f.result, f.err
}

func TestSelectNext(t *testing.T) {
	videos := []tiktok.Video{
		{ID: "1", Description: "not this"},
		{ID: "2", Description: "GOOD MORNING people"},
		{ID: "3", Description: "good morning again"},
	}

	got, ok := selectNext(videos, "good morning", "")
	if !ok || got.ID != "2" {
		t.Fatalf("expected first match ID 2, got %+v ok=%v", got, ok)
	}

	got, ok = selectNext(videos, "good morning", "2")
	if !ok || got.ID != "3" {
		t.Fatalf("expected next match ID 3, got %+v ok=%v", got, ok)
	}
}

func TestRunOnceUpdatesStateAfterSuccess(t *testing.T) {
	store := &fakeStore{}
	tt := &fakeTikTok{videos: []tiktok.Video{{ID: "100", Description: "good morning", URL: "http://x"}}}
	n := &fakeNotifier{result: telegram.SendResult{Mode: "video", MessageID: 1}}

	s := &Service{
		Cfg: config.Config{
			TikTokProfile:     "gmbadass",
			TikTokLookback:    20,
			MatchSubstring:    "good morning",
			TelegramChatID:    "-100",
			SendMode:          "video_with_link_fallback",
			TelegramParseMode: "",
		},
		Store:    store,
		TikTok:   tt,
		Notifier: n,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	if err := s.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if store.lastSent != "100" || store.setCalls != 1 {
		t.Fatalf("expected state update to 100 once, got %q calls=%d", store.lastSent, store.setCalls)
	}
	if n.gotNil {
		t.Fatalf("expected video stream to be provided")
	}
}

func TestRunOnceFallsBackWhenDownloadFails(t *testing.T) {
	store := &fakeStore{}
	tt := &fakeTikTok{videos: []tiktok.Video{{ID: "100", Description: "good morning", URL: "http://x"}}, downloadErr: io.EOF}
	n := &fakeNotifier{result: telegram.SendResult{Mode: "link", MessageID: 2}}

	s := &Service{
		Cfg:      config.Config{TikTokProfile: "gmbadass", TikTokLookback: 20, MatchSubstring: "good morning", TelegramChatID: "-100", SendMode: "video_with_link_fallback"},
		Store:    store,
		TikTok:   tt,
		Notifier: n,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	if err := s.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !n.gotNil {
		t.Fatalf("expected nil video reader after download failure")
	}
	if store.lastSent != "100" {
		t.Fatalf("expected state update on successful fallback send")
	}
}

var _ state.Store = (*fakeStore)(nil)
var _ tiktok.Client = (*fakeTikTok)(nil)
var _ telegram.Client = (*fakeNotifier)(nil)
