package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gmb/internal/app"
	"gmb/internal/config"
	errs "gmb/internal/errors"
	"gmb/internal/instagram"
	"gmb/internal/logging"
	"gmb/internal/scheduler"
	"gmb/internal/state"
	"gmb/internal/telegram"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New("error")
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.LogLevel)
	logger.Debug("loaded configuration",
		"app_cron", cfg.AppCron,
		"one_shot", cfg.OneShot,
		"ig_profile", cfg.InstagramProfile,
		"description_regex", cfg.DescriptionRegex,
		"ig_scan_limit", cfg.InstagramScanLimit,
		"ig_fetch_sleep_seconds", cfg.InstagramFetchSleepSeconds,
		"download_dir", cfg.DownloadDir,
		"download_max_per_run", cfg.DownloadMaxPerRun,
		"download_delay_seconds", cfg.DownloadDelaySeconds,
		"sent_log_path", cfg.SentLogPath,
		"telegram_chat_id", cfg.TelegramChatID,
		"telegram_parse_mode", cfg.TelegramParseMode,
		"telegram_token_set", cfg.TelegramBotToken != "",
		"log_level", cfg.LogLevel,
	)
	if err := app.ValidateWritablePaths(cfg); err != nil {
		logger.Error("startup preflight failed", "error", err)
		os.Exit(1)
	}
	logger.Info("startup preflight passed", "download_dir", cfg.DownloadDir, "sent_log_path", cfg.SentLogPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	service := &app.Service{
		Cfg:       cfg,
		Store:     state.NewFileStore(cfg.SentLogPath),
		Instagram: instagram.NewInstaloaderClient("scripts/instagram/fetch_posts.py", cfg.InstagramFetchSleepSeconds, logger),
		Notifier:  telegram.NewBotAPIClient(cfg.TelegramBotToken),
		Logger:    logger,
	}

	run := func(ctx context.Context) error {
		err := service.RunOnce(ctx)
		if err != nil {
			logger.Error("run failed", "error", err)
			return err
		}
		return nil
	}

	if cfg.OneShot {
		if err := run(ctx); err != nil {
			os.Exit(1)
		}
		return
	}

	logger.Info("starting scheduler", "app_cron", cfg.AppCron)
	if err := scheduler.Run(ctx, logger, cfg.AppCron, run); err != nil && !errors.Is(err, context.Canceled) {
		kind := errs.KindTransient
		var appErr *errs.AppError
		if errors.As(err, &appErr) {
			kind = appErr.Kind
		}
		logger.Error("scheduler exited", slog.String("kind", string(kind)), "error", err)
		os.Exit(1)
	}
}
