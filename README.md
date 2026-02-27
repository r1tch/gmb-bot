# GMB Sender Bot

Containerized Go app that polls TikTok profile `@gmbadass`, finds the newest unsent video with description containing `good morning`, and delivers it through Telegram bot API.

## Behavior
- Looks at latest `TIKTOK_LOOKBACK_LIMIT` videos (default `20`).
- Matches description by case-insensitive substring `MATCH_SUBSTRING` (default `good morning`).
- Attempts to send video bytes to Telegram (`sendVideo`).
- Automatically falls back to sending link (`sendMessage`) on video failure.
- Persists last sent video ID to `STATE_FILE_PATH` (default `/data/state.json`).

## Prerequisites
- Telegram bot token for `gmb_sender_bot`.
- Bot must be allowed to message the destination chat.
- Node + Playwright runtime in container (provided by Dockerfile).

## Telegram Setup
1. Create bot via `@BotFather` (already done in your case).
2. Keep bot token private and set `TELEGRAM_BOT_TOKEN`.
3. Open a chat with your bot and send `/start` once (needed for DM delivery).
4. Get your chat ID:
   - Visit `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Send a test message to the bot first.
   - Find `message.chat.id` from response and use it as `TELEGRAM_CHAT_ID`.

Notes:
- `TELEGRAM_CHAT_ID` is preferred and stable.
- `TELEGRAM_USERNAME` can be used as fallback, but direct-user delivery is less reliable than explicit chat ID.

## Environment Variables
- `TELEGRAM_BOT_TOKEN` (required)
- `TELEGRAM_CHAT_ID` (optional, preferred)
- `TELEGRAM_USERNAME` (optional fallback if no chat ID)
- `APP_CRON` default: `0 * * * *`
- `ONE_SHOT` default: `false`
- `TIKTOK_PROFILE` default: `gmbadass`
- `TIKTOK_LOOKBACK_LIMIT` default: `20`
- `TIKTOK_STORAGE_STATE_PATH` optional path to Playwright storage state JSON (recommended for anti-bot bypass)
- `TIKTOK_SESSIONID` optional TikTok `sessionid` cookie value
- `MATCH_SUBSTRING` default: `good morning`
- `STATE_FILE_PATH` default: `/data/state.json`
- `SEND_MODE` one of: `video_with_link_fallback` (default), `link_only`, `video_only`
- `PLAYWRIGHT_HEADLESS` default: `true`
- `LOG_LEVEL` default: `info`
- `TELEGRAM_PARSE_MODE` optional

## Run with Docker Compose
Create `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:abc...
TELEGRAM_CHAT_ID=123456789
APP_CRON=0 * * * *
TIKTOK_STORAGE_STATE_PATH=/data/tiktok-storage-state.json
# TIKTOK_SESSIONID=...
```

Start:

```bash
docker compose up --build -d
```

Logs:

```bash
docker compose logs -f gmb-bot
```

## One-shot run
Useful for testing one cycle:

```bash
ONE_SHOT=true TELEGRAM_BOT_TOKEN=... TELEGRAM_CHAT_ID=... go run ./cmd/gmb-bot
```

## Development

```bash
make test
make build
```

## Troubleshooting
- `sendMessage/sendVideo 403`: bot is not allowed to talk to the target chat; start bot chat first.
- No videos found: TikTok page structure changed; inspect `scripts/tiktok/fetch_videos.mjs` selectors.
- No videos found with `blocked_likely=true`: TikTok anti-bot page was served. Use `TIKTOK_STORAGE_STATE_PATH` (mounted file) and/or `TIKTOK_SESSIONID`.
- Video send fails but link works: expected fallback behavior.
- Duplicate sends after restart: ensure `/data` volume is mounted and writable.
