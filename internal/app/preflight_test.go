package app

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gmb/internal/config"
)

func TestValidateWritablePathsOK(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DownloadDir: filepath.Join(dir, "video"),
		SentLogPath: filepath.Join(dir, "state", "sent.log"),
	}
	if err := ValidateWritablePaths(cfg); err != nil {
		t.Fatalf("expected writable paths, got error: %v", err)
	}
}

func TestValidateWritablePathsFailsForDownloadDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	root := t.TempDir()
	readOnly := filepath.Join(root, "ro")
	if err := os.MkdirAll(readOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(readOnly, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(readOnly, 0o755) }()

	cfg := config.Config{
		DownloadDir: filepath.Join(readOnly, "video"),
		SentLogPath: filepath.Join(root, "state", "sent.log"),
	}
	if err := ValidateWritablePaths(cfg); err == nil {
		t.Fatal("expected error for non-writable download dir")
	}
}

func TestValidateWritablePathsFailsForSentLog(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits behave differently on windows")
	}

	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sent := filepath.Join(stateDir, "sent.log")
	if err := os.WriteFile(sent, []byte("x\n"), 0o400); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(sent, 0o600) }()

	cfg := config.Config{
		DownloadDir: filepath.Join(root, "video"),
		SentLogPath: sent,
	}
	if err := ValidateWritablePaths(cfg); err == nil {
		t.Fatal("expected error for non-writable sent log")
	}
}
