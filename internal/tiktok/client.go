package tiktok

import (
	"context"
	"io"
	"time"
)

type Video struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	DownloadURL string    `json:"download_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type Client interface {
	ListLatestVideos(ctx context.Context, profile string, limit int) ([]Video, error)
	DownloadVideo(ctx context.Context, video Video) (io.ReadCloser, string, error)
}
