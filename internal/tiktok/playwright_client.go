package tiktok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"time"

	errs "gmb/internal/errors"
)

type PlaywrightClient struct {
	ScriptPath       string
	Headless         bool
	StorageStatePath string
	SessionID        string
	HTTPClient       *http.Client
	Logger           *slog.Logger
}

func NewPlaywrightClient(scriptPath string, headless bool, storageStatePath string, sessionID string, logger *slog.Logger) *PlaywrightClient {
	return &PlaywrightClient{
		ScriptPath:       scriptPath,
		Headless:         headless,
		StorageStatePath: strings.TrimSpace(storageStatePath),
		SessionID:        strings.TrimSpace(sessionID),
		HTTPClient:       &http.Client{Timeout: 45 * time.Second},
		Logger:           logger,
	}
}

func (c *PlaywrightClient) ListLatestVideos(ctx context.Context, profile string, limit int) ([]Video, error) {
	if limit <= 0 {
		return nil, errs.Wrap(errs.KindExtraction, "list-latest", fmt.Errorf("invalid limit %d", limit))
	}

	headlessArg := "true"
	if !c.Headless {
		headlessArg = "false"
	}
	profileURL := fmt.Sprintf("https://www.tiktok.com/@%s", profile)
	if c.Logger != nil {
		c.Logger.Debug(
			"running playwright fetch",
			"profile", profile,
			"url", profileURL,
			"limit", limit,
			"headless", c.Headless,
			"has_storage_state", c.StorageStatePath != "",
			"has_sessionid", c.SessionID != "",
		)
	}

	args := []string{
		c.ScriptPath,
		"--profile", profile,
		"--limit", fmt.Sprintf("%d", limit),
		"--headless", headlessArg,
	}
	if c.StorageStatePath != "" {
		args = append(args, "--storage-state", c.StorageStatePath)
	}
	if c.SessionID != "" {
		args = append(args, "--sessionid", c.SessionID)
	}

	cmd := exec.CommandContext(ctx, "node", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, errs.Wrap(errs.KindExtraction, "playwright-script", fmt.Errorf("%w; stderr=%s", err, strings.TrimSpace(stderr.String())))
	}
	if c.Logger != nil {
		c.Logger.Debug("playwright script completed", "stderr", strings.TrimSpace(stderr.String()), "stdout_bytes", stdout.Len())
	}

	var videos []Video
	if err := json.Unmarshal(stdout.Bytes(), &videos); err != nil {
		return nil, errs.Wrap(errs.KindExtraction, "parse-playwright-output", err)
	}
	if c.Logger != nil {
		c.Logger.Debug("playwright parsed videos", "count", len(videos))
	}
	return videos, nil
}

func (c *PlaywrightClient) DownloadVideo(ctx context.Context, video Video) (io.ReadCloser, string, error) {
	dl := strings.TrimSpace(video.DownloadURL)
	if dl == "" {
		return nil, "", errs.Wrap(errs.KindExtraction, "download-video", fmt.Errorf("empty download URL for video %s", video.ID))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dl, nil)
	if err != nil {
		return nil, "", errs.Wrap(errs.KindExtraction, "build-download-request", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", errs.Wrap(errs.KindTransport, "download-video", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, "", errs.Wrap(errs.KindTransport, "download-video", fmt.Errorf("status %d", resp.StatusCode))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	return resp.Body, contentType, nil
}
