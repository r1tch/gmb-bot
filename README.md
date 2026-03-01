# gmb-bot

Single-container app that:
- downloads matching Instagram videos from a profile using Instaloader,
- stores them under `/data/video`,
- sends one random unsent video to Telegram.

## Required env vars
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_CHAT_ID`

## Optional env vars (defaults)
- `APP_CRON=0 6 * * *`
- `IG_PROFILE=gmbadass`
- `DESCRIPTION_REGEX=(?i)(good ?morning|wake)`
- `IG_SCAN_LIMIT=10`
- `IG_FETCH_SLEEP_SECONDS=1.5`
- `DOWNLOAD_DIR=/data/video`
- `DOWNLOAD_MAX_PER_RUN=100`
- `DOWNLOAD_DELAY_SECONDS=2`
- `SENT_LOG_PATH=/data/state/sent.log`
- `ONE_SHOT=false`
- `LOG_LEVEL=info`

## Local manual runs
- One-shot dev run (download one matching video then random-send from local library) with debug logs:
```bash
make dev
```

- Production mode (scheduled, default daily at 6am):
```bash
make run
```

## Build and push
- Build multi-arch image (`amd64` + `arm64`):
```bash
make build
```

- Build and push:
```bash
make push
```

Image tag:
- `rrrrdockerrrr/gmb-bot:latest`

## NAS setup
1. Pull `rrrrdockerrrr/gmb-bot:latest`.
2. Mount persistent volume to `/data`.
3. Copy the contents of local `gmb_data/` into that server-side `/data` volume (preserve existing `video/` and `state/` data).
4. Set required env vars (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`) and optional overrides.
5. Run container continuously; scheduler inside app handles daily runs.

## Instagram reliability note
In practice, Instaloader often hits `403 Forbidden when accessing https://www.instagram.com/graphql/query` after about every 12 post downloads. It usually retries and continues. In some cases, this escalates to `401` and looks like an IP-level block.

Recommended approach:
1. Do a one-shot bootstrap on a non-banned IP:
   - set `IG_SCAN_LIMIT=500`
   - set `DOWNLOAD_DELAY_SECONDS=2`
   - set `DOWNLOAD_MAX_PER_RUN=1000`
   - set `IG_FETCH_SLEEP_SECONDS=1.5`
   - run `make dev`
2. After library bootstrap, switch to daily mode with `IG_SCAN_LIMIT=12`.

For daily runs, `IG_SCAN_LIMIT=12` has been sufficient.
