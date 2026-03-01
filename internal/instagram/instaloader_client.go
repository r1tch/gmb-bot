package instagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	errs "gmb/internal/errors"
)

type InstaloaderClient struct {
	ScriptPath  string
	FetchSleepS float64
	Logger      *slog.Logger
}

func NewInstaloaderClient(scriptPath string, fetchSleepSeconds float64, logger *slog.Logger) *InstaloaderClient {
	return &InstaloaderClient{
		ScriptPath:  scriptPath,
		FetchSleepS: fetchSleepSeconds,
		Logger:      logger,
	}
}

func (c *InstaloaderClient) ListProfilePosts(ctx context.Context, profile string, scanLimit int) ([]Post, error) {
	if scanLimit <= 0 {
		return nil, errs.Wrap(errs.KindConfig, "scan-limit", fmt.Errorf("scan limit must be > 0"))
	}

	args := []string{
		c.ScriptPath,
		"--profile", profile,
		"--scan-limit", fmt.Sprintf("%d", scanLimit),
		"--sleep-seconds", fmt.Sprintf("%g", c.FetchSleepS),
	}
	if c.Logger != nil {
		c.Logger.Debug(
			"running instagram fetch script",
			"command", "python3",
			"args", args,
			"profile", profile,
			"scan_limit", scanLimit,
			"fetch_sleep_seconds", c.FetchSleepS,
		)
	}
	cmd := exec.CommandContext(ctx, "python3", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errs.Wrap(errs.KindExtraction, "instaloader-list", fmt.Errorf("%w; stderr=%s", err, truncate(strings.TrimSpace(stderr.String()), 3000)))
	}

	var posts []Post
	if err := json.Unmarshal(stdout.Bytes(), &posts); err != nil {
		return nil, errs.Wrap(errs.KindExtraction, "instaloader-parse", err)
	}
	return posts, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...<truncated>"
}
