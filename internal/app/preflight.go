package app

import (
	"os"
	"path/filepath"

	"gmb/internal/config"
	errs "gmb/internal/errors"
)

func ValidateWritablePaths(cfg config.Config) error {
	if err := ensureDirWritable(cfg.DownloadDir); err != nil {
		return errs.Wrap(errs.KindConfig, "download-dir-writable", err)
	}
	if err := ensureFileWritable(cfg.SentLogPath); err != nil {
		return errs.Wrap(errs.KindConfig, "sent-log-writable", err)
	}
	return nil
}

func ensureDirWritable(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".gmb-write-check-*")
	if err != nil {
		return err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
}

func ensureFileWritable(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}
