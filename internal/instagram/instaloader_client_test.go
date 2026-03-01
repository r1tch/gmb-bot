package instagram

import (
	"context"
	"strings"
	"testing"
)

func TestListProfilePostsRejectsInvalidScanLimit(t *testing.T) {
	c := NewInstaloaderClient("scripts/instagram/fetch_posts.py", 0, nil)
	_, err := c.ListProfilePosts(context.Background(), "gmbadass", 0)
	if err == nil {
		t.Fatal("expected error for scan limit 0")
	}
	if !strings.Contains(err.Error(), "scan limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abc", 10); got != "abc" {
		t.Fatalf("expected same string, got %q", got)
	}
	got := truncate("abcdefghijklmnopqrstuvwxyz", 5)
	if got != "abcde...<truncated>" {
		t.Fatalf("unexpected truncate output: %q", got)
	}
}
