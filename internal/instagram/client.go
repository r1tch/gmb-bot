package instagram

import "context"

type Post struct {
	Shortcode string `json:"shortcode"`
	Caption   string `json:"caption"`
	VideoURL  string `json:"video_url"`
	DateUTC   string `json:"date_utc"`
	IsVideo   bool   `json:"is_video"`
}

type Client interface {
	ListProfilePosts(ctx context.Context, profile string, scanLimit int) ([]Post, error)
}
