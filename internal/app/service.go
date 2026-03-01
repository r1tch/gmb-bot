package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gmb/internal/config"
	errs "gmb/internal/errors"
	"gmb/internal/instagram"
	"gmb/internal/library"
	"gmb/internal/state"
	"gmb/internal/telegram"
)

type Service struct {
	Cfg       config.Config
	Store     state.Store
	Instagram instagram.Client
	Notifier  telegram.Client
	Logger    *slog.Logger
	HTTP      *http.Client
	Sleep     func(time.Duration)
}

func (s *Service) RunOnce(ctx context.Context) error {
	if s.HTTP == nil {
		s.HTTP = &http.Client{Timeout: 90 * time.Second}
	}
	if s.Sleep == nil {
		s.Sleep = time.Sleep
	}

	start := time.Now()
	runID := fmt.Sprintf("run-%d", start.UnixNano())
	logger := s.Logger.With("run_id", runID)

	rx, err := regexp.Compile(s.Cfg.DescriptionRegex)
	if err != nil {
		return errs.Wrap(errs.KindConfig, "compile-regex", err)
	}

	posts, err := s.Instagram.ListProfilePosts(ctx, s.Cfg.InstagramProfile, s.Cfg.InstagramScanLimit)
	if err != nil {
		return err
	}
	logger.Info("fetched instagram posts", "count", len(posts))
	logScannedSamples(logger, posts, 10)

	downloadLimit := s.Cfg.DownloadMaxPerRun
	logger.Debug("download policy", "limit", downloadLimit, "delay_seconds", s.Cfg.DownloadDelaySeconds)

	downloaded := 0
	skippedExisting := 0
	skippedUnmatched := 0
	skippedNoURL := 0

	for i, p := range posts {
		if downloaded >= downloadLimit {
			break
		}
		logger.Debug("evaluating post", "index", i, "shortcode", p.Shortcode, "is_video", p.IsVideo, "has_video_url", strings.TrimSpace(p.VideoURL) != "")
		if !p.IsVideo || !rx.MatchString(p.Caption) {
			skippedUnmatched++
			continue
		}
		if strings.TrimSpace(p.VideoURL) == "" {
			skippedNoURL++
			continue
		}
		path, didDownload, err := s.ensureDownloaded(ctx, logger, p)
		if err != nil {
			return err
		}
		if !didDownload {
			skippedExisting++
			logger.Debug("video already exists", "shortcode", p.Shortcode, "path", path)
			continue
		}
		downloaded++
		logger.Info("downloaded video", "shortcode", p.Shortcode, "path", path, "downloaded_count", downloaded)
		if downloaded < downloadLimit && s.Cfg.DownloadDelaySeconds > 0 {
			d := time.Duration(s.Cfg.DownloadDelaySeconds) * time.Second
			logger.Debug("sleeping between downloads", "duration", d.String())
			s.Sleep(d)
		}
	}

	logger.Info("download summary", "downloaded", downloaded, "skipped_existing", skippedExisting, "skipped_unmatched", skippedUnmatched, "skipped_no_video_url", skippedNoURL)

	files, err := library.ListVideoFiles(s.Cfg.DownloadDir)
	if err != nil {
		return errs.Wrap(errs.KindState, "list-videos", err)
	}
	logger.Debug("local library scanned", "count", len(files))
	if len(files) == 0 {
		logger.Info("no local videos available for sending")
		return nil
	}

	sent, err := s.Store.ListSentIDs(ctx)
	if err != nil {
		return err
	}
	logger.Debug("loaded sent history", "count", len(sent))

	selected, ok := library.PickRandomUnsent(files, sent)
	if !ok {
		logger.Debug("all local videos already sent; resetting history")
		if err := s.Store.ResetSent(ctx); err != nil {
			return err
		}
		selected = files[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(files))]
		logger.Info("sent history exhausted; reset performed")
	}
	logger.Debug("selected video for send", "video_id", selected.ID, "path", selected.Path)

	if err := s.sendFile(ctx, logger, selected.Path); err != nil {
		return err
	}
	if err := s.Store.AppendSentID(ctx, selected.ID); err != nil {
		return err
	}

	logger.Info("video sent", "video_id", selected.ID, "path", selected.Path, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func logScannedSamples(logger *slog.Logger, posts []instagram.Post, max int) {
	if max <= 0 {
		return
	}
	if len(posts) < max {
		max = len(posts)
	}
	for i := 0; i < max; i++ {
		p := posts[i]
		logger.Debug(
			"scanned post sample",
			"index", i,
			"shortcode", p.Shortcode,
			"is_video", p.IsVideo,
			"has_video_url", strings.TrimSpace(p.VideoURL) != "",
			"date_utc", p.DateUTC,
			"caption_preview", trimPreview(p.Caption, 120),
		)
	}
}

func trimPreview(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func (s *Service) ensureDownloaded(ctx context.Context, logger *slog.Logger, p instagram.Post) (string, bool, error) {
	date, err := postDateYYYYMMDD(p.DateUTC)
	if err != nil {
		return "", false, errs.Wrap(errs.KindExtraction, "parse-post-date", err)
	}
	path := library.BuildVideoPath(s.Cfg.DownloadDir, s.Cfg.InstagramProfile, date, p.Shortcode)
	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	} else if !os.IsNotExist(err) {
		return "", false, errs.Wrap(errs.KindState, "stat-video", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", false, errs.Wrap(errs.KindState, "mkdir-download-dir", err)
	}

	logger.Debug("starting video download", "shortcode", p.Shortcode, "url", p.VideoURL, "path", path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.VideoURL, nil)
	if err != nil {
		return "", false, errs.Wrap(errs.KindExtraction, "build-download-request", err)
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return "", false, errs.Wrap(errs.KindTransport, "download-video", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", false, errs.Wrap(errs.KindTransport, "download-video", fmt.Errorf("status %d", resp.StatusCode))
	}

	tmp := path + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", false, errs.Wrap(errs.KindState, "open-download-file", err)
	}
	n, err := io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		return "", false, errs.Wrap(errs.KindTransport, "write-download-file", err)
	}
	if err := f.Close(); err != nil {
		return "", false, errs.Wrap(errs.KindState, "close-download-file", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", false, errs.Wrap(errs.KindState, "rename-download-file", err)
	}
	logger.Debug("completed video download", "shortcode", p.Shortcode, "bytes", n, "path", path)
	return path, true, nil
}

func (s *Service) sendFile(ctx context.Context, logger *slog.Logger, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errs.Wrap(errs.KindState, "open-video", err)
	}
	defer f.Close()

	caption := buildCaptionFromPath(s.Cfg.InstagramProfile, path)
	logger.Debug("sending telegram video", "chat_id", s.Cfg.TelegramChatID, "path", path, "caption", caption)
	result, err := s.Notifier.SendVideo(ctx, telegram.SendRequest{
		ChatID:    s.Cfg.TelegramChatID,
		Caption:   caption,
		ParseMode: s.Cfg.TelegramParseMode,
	}, f, "video/mp4")
	if err != nil {
		return errs.Wrap(errs.KindTransport, "notify", err)
	}
	logger.Debug("telegram send success", "message_id", result.MessageID)
	return nil
}

func buildCaptionFromPath(profile, path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	prefix := strings.TrimSpace(profile) + "_"
	if strings.HasPrefix(base, prefix) {
		rest := strings.TrimPrefix(base, prefix)
		if len(rest) >= 8 {
			datePart := rest[:8]
			if _, err := time.Parse("20060102", datePart); err == nil {
				return fmt.Sprintf(
					"%s - %s-%s-%s",
					strings.TrimSpace(profile),
					datePart[0:4],
					datePart[4:6],
					datePart[6:8],
				)
			}
		}
	}
	return strings.TrimSpace(profile)
}

func postDateYYYYMMDD(raw string) (string, error) {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("20060102"), nil
}

func extractShortcodeFromID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) == 0 {
		return id
	}
	return parts[len(parts)-1]
}
