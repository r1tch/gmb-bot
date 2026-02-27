package state

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	errs "gmb/internal/errors"
)

type Store interface {
	GetLastSentID(ctx context.Context) (string, error)
	SetLastSentID(ctx context.Context, id string) error
}

type FileStore struct {
	Path string
}

type filePayload struct {
	LastSentVideoID string `json:"last_sent_video_id"`
}

func NewFileStore(path string) *FileStore {
	return &FileStore{Path: path}
}

func (s *FileStore) GetLastSentID(_ context.Context) (string, error) {
	b, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errs.Wrap(errs.KindState, "read-state", err)
	}

	var p filePayload
	if err := json.Unmarshal(b, &p); err != nil {
		return "", errs.Wrap(errs.KindState, "parse-state", err)
	}
	return p.LastSentVideoID, nil
}

func (s *FileStore) SetLastSentID(_ context.Context, id string) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return errs.Wrap(errs.KindState, "mkdir-state-dir", err)
	}
	b, err := json.MarshalIndent(filePayload{LastSentVideoID: id}, "", "  ")
	if err != nil {
		return errs.Wrap(errs.KindState, "marshal-state", err)
	}
	if err := os.WriteFile(s.Path, b, 0o600); err != nil {
		return errs.Wrap(errs.KindState, "write-state", err)
	}
	return nil
}
