package telegram

import (
	"context"
	"io"
)

type SendRequest struct {
	ChatID    string
	Username  string
	VideoURL  string
	Caption   string
	ParseMode string
}

type SendResult struct {
	Mode      string
	MessageID int
}

type Client interface {
	SendVideoOrFallback(ctx context.Context, req SendRequest, video io.Reader, contentType string) (SendResult, error)
}
