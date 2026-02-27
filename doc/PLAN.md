# Go Container App Plan: TikTok "Good Morning" to Telegram Bot

## Summary
Build a containerized Go service that runs hourly, scans the latest configurable N videos from `@gmbadass` (default 20), finds the newest unsent video whose description contains `good morning` (case-insensitive), and sends it through your Telegram bot (`gmb_sender_bot`) to a configured recipient.

Delivery behavior:
1. Plan A: send video via Telegram `sendVideo`.
2. Plan B: automatically fallback to sending the TikTok link via `sendMessage` if video delivery fails.

State behavior:
- Persist `last_sent_video_id` in a mounted JSON file to avoid duplicate sends across restarts.

## Architecture and Implementation Details

### 1) Repository Layout
- `cmd/gmb-bot/main.go`
- `internal/config/`
- `internal/logging/`
- `internal/tiktok/`
  - `client.go`
  - `playwright_client.go`
- `internal/telegram/`
  - `client.go`
  - `botapi_client.go`
- `internal/state/`
  - `store.go`
  - `file_store.go`
- `internal/scheduler/cron.go`
- `internal/app/service.go`
- `internal/errors/`
- `Dockerfile`, `docker-compose.yml`, `Makefile`, `README.md`

### 2) Public Interfaces/Types (important)
- `type TikTokClient interface { ListLatestVideos(ctx context.Context, profile string, limit int) ([]Video, error); DownloadVideo(ctx context.Context, v Video) (io.ReadCloser, string, error) }`
- `type Notifier interface { SendVideoOrFallback(ctx context.Context, req SendRequest) (SendResult, error) }`
- `type StateStore interface { GetLastSentID(ctx context.Context) (string, error); SetLastSentID(ctx context.Context, id string) error }`
- `type Video struct { ID, URL, Description, DownloadURL string; CreatedAt time.Time }`
- `type SendRequest struct { ChatID string; Username string; Video Video; FallbackText string }`
- `type SendResult struct { Mode string /* video|link */; MessageID int }`

### 3) Core Flow
1. Cron tick (`0 * * * *` by default).
2. Load config and read `last_sent_video_id`.
3. Fetch latest videos from TikTok profile using Playwright-based extraction.
4. Filter by case-insensitive substring `good morning`.
5. Choose newest matching video not equal to last sent ID.
6. Resolve target recipient:
   - prefer `TELEGRAM_CHAT_ID`
   - fallback to `TELEGRAM_USERNAME` if chat ID absent
7. Try `sendVideo` with downloaded bytes.
8. On send failure/unsupported constraints, send TikTok URL text via `sendMessage`.
9. Update state only after successful video or link delivery.
10. Log structured run result to stdout.

### 4) Configuration (env vars)
- `APP_CRON` default `0 * * * *`
- `TIKTOK_PROFILE` default `gmbadass`
- `TIKTOK_LOOKBACK_LIMIT` default `20`
- `MATCH_SUBSTRING` default `good morning`
- `STATE_FILE_PATH` default `/data/state.json`
- `TELEGRAM_BOT_TOKEN` required
- `TELEGRAM_CHAT_ID` optional (preferred)
- `TELEGRAM_USERNAME` optional (fallback if no chat ID)
- `TELEGRAM_PARSE_MODE` optional (`MarkdownV2`/`HTML`/empty)
- `LOG_LEVEL` default `info`
- `PLAYWRIGHT_HEADLESS` default `true`
- `SEND_MODE` default `video_with_link_fallback`

Validation:
- Require at least one of `TELEGRAM_CHAT_ID` or `TELEGRAM_USERNAME`.
- If both provided, use chat ID.

### 5) Docker and Runtime
- Multi-stage build:
  - compile Go binary
  - runtime with Chromium/Playwright dependencies
- Mount volume for `/data` to persist state file.
- Long-running cron process in container.
- Optional one-shot mode flag for CI/manual execution (without changing default cron behavior).

### 6) README Content
- Bot setup prerequisites (already created bot, token handling).
- How to obtain `chat_id` from Telegram updates (recommended stable target method).
- Env var reference and example `.env`.
- Docker run / compose examples with mounted `/data`.
- Behavior docs for Plan A vs Plan B fallback.
- Troubleshooting:
  - bot not allowed in target chat
  - invalid chat ID/username
  - TikTok extraction changes
  - large video or delivery failures causing link fallback

### 7) Engineering Practices
- TDD-first for core selection/fallback/state rules.
- SOLID boundaries with interfaces for TikTok and Telegram adapters.
- Structured JSON logs to stdout with run ID, selected video ID, send mode, latency, and error class.
- Typed errors: extraction, transport, config, transient/permanent.
- Idempotent state updates (never mark sent on failed delivery).
- Secret-safe logging (never print token).

### 8) Test Plan and Acceptance Criteria

#### Unit tests
- config validation and recipient precedence.
- description matcher (case-insensitive).
- selection logic (newest unsent match).
- fallback logic (`sendVideo` failure => `sendMessage` success).
- state write semantics (only after successful send).

#### Integration-style tests (mocked APIs)
- TikTok list + description mix in latest 20.
- Telegram video success path.
- Telegram video failure then link success path.
- target resolution with chat ID and username fallback.

#### Container smoke test
- test mode runs a single cycle, emits expected logs, exits 0 on success path.

#### Acceptance criteria
- Hourly run checks latest N videos.
- Sends first unsent "Good Morning" video to Telegram.
- Falls back to link automatically when video send fails.
- Persists `last_sent_video_id` across restart with mounted volume.
- README is sufficient for another engineer to configure and run.

## Assumptions and Defaults Chosen
- TikTok retrieval uses browser automation (no login).
- Default schedule remains hourly.
- Matching uses case-insensitive substring.
- Telegram bot token is provided via env var.
- Recipient precedence: `TELEGRAM_CHAT_ID` first, `TELEGRAM_USERNAME` second.
- Default delivery mode is video with automatic link fallback.

## Current TikTok Method (Step-by-step)
- Current step: open `https://www.tiktok.com/@gmbadass` and list first ~20 video links in the form `https://www.tiktok.com/@gmbadass/video/<video_id>`.
- Next step (recorded method): click videos one-by-one, inspect each video description, and find the first description containing `Good Morning`.
- Download approach to implement next: on matching video page, right-click the playing video and choose `Download video`; this triggers browser download for the media file.
