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
	"gmb/internal/logging"
	"gmb/internal/scheduler"
	"gmb/internal/state"
	"gmb/internal/telegram"
	"gmb/internal/tiktok"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New("error")
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.LogLevel)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	service := &app.Service{
		Cfg:      cfg,
		Store:    state.NewFileStore(cfg.StateFilePath),
		TikTok:   tiktok.NewPlaywrightClient("scripts/tiktok/fetch_videos.mjs", cfg.PlaywrightHeadless, cfg.TikTokStorageState, cfg.TikTokSessionID, logger),
		Notifier: telegram.NewBotAPIClient(cfg.TelegramBotToken),
		Logger:   logger,
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
	if err := scheduler.Run(ctx, cfg.AppCron, run); err != nil && !errors.Is(err, context.Canceled) {
		kind := errs.KindTransient
		var appErr *errs.AppError
		if errors.As(err, &appErr) {
			kind = appErr.Kind
		}
		logger.Error("scheduler exited", slog.String("kind", string(kind)), "error", err)
		os.Exit(1)
	}
}
