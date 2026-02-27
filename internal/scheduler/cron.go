package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Minimal scheduler: supports default hourly cron ("0 * * * *") and "@every <duration>".
func Run(ctx context.Context, expr string, fn func(context.Context) error) error {
	if strings.HasPrefix(expr, "@every ") {
		d, err := time.ParseDuration(strings.TrimPrefix(expr, "@every "))
		if err != nil {
			return fmt.Errorf("parse APP_CRON duration: %w", err)
		}
		t := time.NewTicker(d)
		defer t.Stop()
		for {
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

	if expr != "0 * * * *" {
		return fmt.Errorf("unsupported APP_CRON expression %q; use '0 * * * *' or '@every <duration>'", expr)
	}

	for {
		next := time.Now().Truncate(time.Hour).Add(time.Hour)
		wait := time.Until(next)
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
