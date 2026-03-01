package state

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"

	errs "gmb/internal/errors"
)

type Store interface {
	ListSentIDs(ctx context.Context) (map[string]struct{}, error)
	AppendSentID(ctx context.Context, id string) error
	ResetSent(ctx context.Context) error
}

type FileStore struct {
	Path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{Path: path}
}

func (s *FileStore) ListSentIDs(_ context.Context) (map[string]struct{}, error) {
	ids := map[string]struct{}{}
	f, err := os.Open(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return ids, nil
		}
		return nil, errs.Wrap(errs.KindState, "read-sent-log", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id == "" {
			continue
		}
		ids[id] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, errs.Wrap(errs.KindState, "scan-sent-log", err)
	}
	return ids, nil
}

func (s *FileStore) AppendSentID(_ context.Context, id string) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return errs.Wrap(errs.KindState, "mkdir-sent-dir", err)
	}
	f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return errs.Wrap(errs.KindState, "open-sent-log", err)
	}
	defer f.Close()
	if _, err := f.WriteString(strings.TrimSpace(id) + "\n"); err != nil {
		return errs.Wrap(errs.KindState, "append-sent-log", err)
	}
	return nil
}

func (s *FileStore) ResetSent(_ context.Context) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return errs.Wrap(errs.KindState, "mkdir-sent-dir", err)
	}
	if err := os.WriteFile(s.Path, []byte(""), 0o600); err != nil {
		return errs.Wrap(errs.KindState, "reset-sent-log", err)
	}
	return nil
}
