package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestRunEvery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	count := 0
	err := Run(ctx, "@every 10ms", func(context.Context) error {
		count++
		if count >= 2 {
			cancel()
		}
		return nil
	})
	if err == nil {
		t.Fatalf("expected cancellation error")
	}
	if count < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", count)
	}
}
