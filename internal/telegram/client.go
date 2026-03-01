package telegram

import (
	"context"
	"io"
)

type SendRequest struct {
	ChatID    string
	Caption   string
	ParseMode string
}

type SendResult struct {
	MessageID int
}

type Client interface {
	SendVideo(ctx context.Context, req SendRequest, video io.Reader, contentType string) (SendResult, error)
}
