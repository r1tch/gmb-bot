package telegram

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSendVideoValidations(t *testing.T) {
	c := NewBotAPIClient("token")
	if _, err := c.SendVideo(context.Background(), SendRequest{}, strings.NewReader("x"), "video/mp4"); err == nil {
		t.Fatal("expected error when chat id missing")
	}
	if _, err := c.SendVideo(context.Background(), SendRequest{ChatID: "123"}, nil, "video/mp4"); err == nil {
		t.Fatal("expected error when video reader missing")
	}
}

func TestSendVideoSuccess(t *testing.T) {
	c := NewBotAPIClient("token")
	c.HTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"ok":true,"result":{"message_id":42}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}

	res, err := c.SendVideo(context.Background(), SendRequest{ChatID: "123", Caption: "x"}, strings.NewReader("video"), "video/mp4")
	if err != nil {
		t.Fatalf("SendVideo: %v", err)
	}
	if res.MessageID != 42 {
		t.Fatalf("expected message id 42, got %d", res.MessageID)
	}
}
