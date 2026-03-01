package state

import (
	"context"
	"testing"
)

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir + "/sent.log")

	ids, err := store.ListSentIDs(context.Background())
	if err != nil {
		t.Fatalf("ListSentIDs: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty initial ids, got %d", len(ids))
	}

	if err := store.AppendSentID(context.Background(), "abc"); err != nil {
		t.Fatalf("AppendSentID: %v", err)
	}
	if err := store.AppendSentID(context.Background(), "def"); err != nil {
		t.Fatalf("AppendSentID 2: %v", err)
	}

	ids, err = store.ListSentIDs(context.Background())
	if err != nil {
		t.Fatalf("ListSentIDs 2: %v", err)
	}
	if _, ok := ids["abc"]; !ok {
		t.Fatalf("expected abc")
	}
	if _, ok := ids["def"]; !ok {
		t.Fatalf("expected def")
	}

	if err := store.ResetSent(context.Background()); err != nil {
		t.Fatalf("ResetSent: %v", err)
	}
	ids, err = store.ListSentIDs(context.Background())
	if err != nil {
		t.Fatalf("ListSentIDs 3: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty after reset, got %d", len(ids))
	}
}
