package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Minimal scheduler: supports daily 6AM cron ("0 6 * * *") and "@every <duration>".
func Run(ctx context.Context, logger *slog.Logger, expr string, fn func(context.Context) error) error {
	if logger == nil {
		logger = slog.Default()
	}

	if strings.HasPrefix(expr, "@every ") {
		d, err := time.ParseDuration(strings.TrimPrefix(expr, "@every "))
		if err != nil {
			return fmt.Errorf("parse APP_CRON duration: %w", err)
		}
		return runEvery(ctx, logger, d, fn)
	}

	if expr == "0 6 * * *" {
		return runAt(ctx, logger, 6, 0, fn)
	}
	
	return fmt.Errorf("Scheduler: unsupported APP_CRON expression %q; use '0 6 * * *' or '@every <duration>'", expr)
}

func runEvery(ctx context.Context, logger *slog.Logger, interval time.Duration, fn func(context.Context) error) error {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		now := time.Now()
		next := now.Add(interval)
		logger.Info("next scheduled run", "next_run_at", next.Format(time.RFC3339), "wait_seconds", int(interval.Seconds()), "mode", "@every")
		if err := fn(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func runAt(ctx context.Context, logger *slog.Logger, hour int, minute int, fn func(context.Context) error) error {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		wait := time.Until(next)
		logger.Info("next scheduled run", "next_run_at", next.Format(time.RFC3339), "wait_seconds", int(wait.Seconds()), "mode", "daily-6am")

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		if err := fn(ctx); err != nil {
			return err
		}
	}
}
