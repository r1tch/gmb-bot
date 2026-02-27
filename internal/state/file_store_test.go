package state

import (
	"context"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir + "/state.json")

	id, err := store.GetLastSentID(context.Background())
	if err != nil {
		t.Fatalf("GetLastSentID: %v", err)
	}
	if id != "" {
		t.Fatalf("expected empty initial id, got %q", id)
	}

	if err := store.SetLastSentID(context.Background(), "12345"); err != nil {
		t.Fatalf("SetLastSentID: %v", err)
	}

	id, err = store.GetLastSentID(context.Background())
	if err != nil {
		t.Fatalf("GetLastSentID 2: %v", err)
	}
	if id != "12345" {
		t.Fatalf("expected 12345, got %q", id)
	}
}
