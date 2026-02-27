package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"gmb/internal/config"
	errs "gmb/internal/errors"
	"gmb/internal/state"
	"gmb/internal/telegram"
	"gmb/internal/tiktok"
)

type Service struct {
	Cfg      config.Config
	Store    state.Store
	TikTok   tiktok.Client
	Notifier telegram.Client
	Logger   *slog.Logger
}

func (s *Service) RunOnce(ctx context.Context) error {
	start := time.Now()
	runID := fmt.Sprintf("run-%d", start.UnixNano())
	logger := s.Logger.With("run_id", runID)

	lastSentID, err := s.Store.GetLastSentID(ctx)
	if err != nil {
		return err
	}
	logger.Info("loaded state", "last_sent_video_id", lastSentID)

	videos, err := s.TikTok.ListLatestVideos(ctx, s.Cfg.TikTokProfile, s.Cfg.TikTokLookback)
	if err != nil {
		return err
	}
	logger.Info("fetched videos", "count", len(videos))
	for i, v := range videos {
		if i >= 3 {
			break
		}
		logger.Debug("video sample", "index", i, "video_id", v.ID, "url", v.URL, "description", v.Description)
	}

	video, ok := selectNext(videos, s.Cfg.MatchSubstring, lastSentID)
	if !ok {
		logger.Info("no matching unsent video found")
		return nil
	}

	caption := fmt.Sprintf("%s\n%s", video.Description, video.URL)
	request := telegram.SendRequest{
		ChatID:    s.Cfg.TelegramChatID,
		Username:  s.Cfg.TelegramUsername,
		VideoURL:  video.URL,
		Caption:   caption,
		ParseMode: s.Cfg.TelegramParseMode,
	}

	var body io.ReadCloser
	contentType := ""
	if s.Cfg.SendMode != "link_only" {
		body, contentType, err = s.TikTok.DownloadVideo(ctx, video)
		if err != nil {
			logger.Warn("video download failed; notifier will fallback to link", "error", err)
		}
	}
	if body != nil {
		defer body.Close()
	}

	result, err := s.Notifier.SendVideoOrFallback(ctx, request, body, contentType)
	if err != nil {
		return errs.Wrap(errs.KindTransport, "notify", err)
	}

	if err := s.Store.SetLastSentID(ctx, video.ID); err != nil {
		return err
	}

	logger.Info("video sent", "video_id", video.ID, "mode", result.Mode, "message_id", result.MessageID, "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func selectNext(videos []tiktok.Video, matchSubstring, lastSentID string) (tiktok.Video, bool) {
	needle := strings.ToLower(strings.TrimSpace(matchSubstring))
	for _, v := range videos {
		if lastSentID != "" && v.ID == lastSentID {
			continue
		}
		desc := strings.ToLower(v.Description)
		if strings.Contains(desc, needle) {
			return v, true
		}
	}
	return tiktok.Video{}, false
}
