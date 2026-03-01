# Implementation Status (2026-03-01)

## Summary
The project has been migrated from the TikTok/Playwright approach to an Instagram + Instaloader pipeline with Telegram delivery, packaged as a single Go-based container app.

Current baseline:
- One container handles both downloading and sending.
- Instagram metadata is fetched via Python (`instaloader`), invoked from Go.
- Videos are stored locally under `/data/video` and one random unsent local video is sent to Telegram per run.
- Production mode runs on schedule; one-shot mode runs once and exits.

## Completed Architecture Changes
- Removed TikTok/Playwright/browser automation path from active flow.
- Implemented Instagram client integration:
  - Go calls `scripts/instagram/fetch_posts.py`.
  - JSON output is parsed into internal post structs.
- Unified app flow in `internal/app/service.go`:
  - Fetch posts.
  - Log scan samples.
  - Filter by caption regex and video flag.
  - Download missing videos.
  - Select random unsent local video.
  - Send to Telegram.
  - Persist sent history.
- Sent-history cycling implemented:
  - If all videos were sent, sent log is reset and selection continues.

## Current Runtime Behavior
- `ONE_SHOT=true`:
  - Performs one full run (download + send) then exits.
- `ONE_SHOT=false`:
  - Starts scheduler and runs by `APP_CRON`.
- Default cron:
  - `0 6 * * *` (daily 6am).

## Configuration and Logging Improvements
- Added startup config dump at debug level (safe/operational fields).
- Added detailed debug logging for:
  - external fetch command arguments,
  - first scanned post samples,
  - download decisions and outcomes,
  - Telegram send requests/results,
  - scheduler next-run information.
- Added configurable Instagram fetch pacing:
  - Env: `IG_FETCH_SLEEP_SECONDS` (float, default `0.0`).
  - Propagated to Python as `--sleep-seconds`.
- Existing download pacing remains:
  - `DOWNLOAD_DELAY_SECONDS` between successful downloads.

## Python Script Status
- `scripts/instagram/fetch_posts.py` supports:
  - `--profile`
  - `--scan-limit`
  - `--sleep-seconds` (float)
- Sleep is applied periodically during listing to reduce aggressive request cadence.

## Test Status
- Go tests: passing via `go test ./...`.
- Python unit test: passing (`scripts.instagram.test_fetch_posts_unit`).
- E2E-like Python test exists (`scripts.instagram.test_fetch_posts_e2e`) and is network dependent.

Make targets:
- `make test`: Go tests + Python unit test.
- `make e2e-test`: Python Instagram network test.

## Container and Build Status
- `docker-compose` removed from active workflow.
- Local execution uses `docker run` via Makefile targets.
- Make targets currently available:
  - `run`, `dev`, `test`, `e2e-test`, `build`, `push`.
- Multi-platform image build configured for:
  - `linux/amd64`
  - `linux/arm64`
- Published image target:
  - `rrrrdockerrrr/gmb-bot:latest`.

## Operational Guidance Captured in README
- NAS setup now explicitly requires copying local `gmb_data/` contents to server `/data` volume.
- Instagram reliability note documented:
  - frequent `403 graphql/query` bursts (often recoverable with retries),
  - occasional `401`/temporary block behavior.
- Recommended bootstrap pattern documented:
  - one-shot bulk run on healthy IP,
  - then reduce `IG_SCAN_LIMIT` to 12 for daily steady-state.

## Known Constraints / Risks
- Instagram anti-abuse behavior remains external/unpredictable.
- E2E fetch/download behavior depends on source availability, network conditions, and IP reputation.
- `make run` currently uses `LOG_LEVEL=debug` explicitly; adjust per operational preference.

## Key Files Updated
- `cmd/gmb-bot/main.go`
- `internal/app/service.go`
- `internal/config/config.go`
- `internal/instagram/instaloader_client.go`
- `scripts/instagram/fetch_posts.py`
- `scripts/instagram/test_fetch_posts_unit.py`
- `scripts/instagram/test_fetch_posts_e2e.py`
- `Makefile`
- `README.md`
- `doc/PLAN.md`

## Next Practical Focus
1. Add retry/backoff policy around Instagram listing/download to reduce hard failures on transient 403/401 bursts.
2. Add a dry-run mode for scan/filter decision verification without network download/send.
3. Add compact run summary metrics (matched/downloaded/skipped/sent) for easier NAS monitoring.
