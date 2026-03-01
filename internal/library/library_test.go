package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildVideoPathAndID(t *testing.T) {
	p := BuildVideoPath("/data/video", "gmbadass", "20260301", "ABC123")
	expected := filepath.Join("/data/video", "gmbadass_20260301-ABC123.mp4")
	if p != expected {
		t.Fatalf("expected %q, got %q", expected, p)
	}
	if id := VideoIDFromPath(p); id != "gmbadass_20260301-ABC123" {
		t.Fatalf("unexpected id %q", id)
	}
}

func TestListVideoFilesSortAndFilter(t *testing.T) {
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, "b.mp4"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(d, "a.mp4"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(d, "ignore.txt"), []byte("x"), 0o644)

	files, err := ListVideoFiles(d)
	if err != nil {
		t.Fatalf("ListVideoFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].ID != "a" || files[1].ID != "b" {
		t.Fatalf("expected sorted ids a,b got %v,%v", files[0].ID, files[1].ID)
	}
}

func TestPickRandomUnsent(t *testing.T) {
	files := []VideoFile{{ID: "a"}, {ID: "b"}}
	picked, ok := PickRandomUnsent(files, map[string]struct{}{"a": {}})
	if !ok {
		t.Fatal("expected a pick")
	}
	if picked.ID != "b" {
		t.Fatalf("expected b, got %s", picked.ID)
	}
	_, ok = PickRandomUnsent(files, map[string]struct{}{"a": {}, "b": {}})
	if ok {
		t.Fatal("expected no pick when all sent")
	}
}
