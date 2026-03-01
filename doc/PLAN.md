# Revised Plan: Single Container Instagram Downloader + Telegram Sender

## Summary
Build one containerized Go application (no docker-compose) that uses Instagram data via Python Instaloader (invoked from Go), stores videos on disk, and sends one random unsent video to Telegram.

The same image supports:
- Production mode: daily scheduled run (default 06:00), downloads next batch, then sends one random unsent video.
- Dev mode: one-shot run that downloads the first matching video and immediately sends it.

## Core Requirements (Locked)
- Single container for both downloader and sender.
- Multi-platform image build: `linux/amd64` (NAS) and `linux/arm64` (macOS).
- No Playwright / Chrome / Chromium / browser automation.
- Use Python Instaloader from Go via command invocation.
- Storage path format:
  - `/data/video/<instagram_profile>_<YYYYMMDD>-<shortcode>.mp4`
- Polite downloading:
  - 30-second delay between each downloaded video.
- Image name/tag to publish:
  - `rrrrdockerrrr/gmb-bot:latest`

## Runtime Modes

### 1) Production Mode
- Schedule: once per day, default at 6:00 AM.
- Per run behavior:
  1. List Instagram profile posts.
  2. Filter for video posts whose description/caption matches configured regex.
  3. Download next batch of missing videos (with 30s delay between downloads).
  4. Select one random unsent local video.
  5. Send to Telegram.
  6. Track sent videos; if all known videos were sent, reset sent-history and continue cycle.

### 2) Dev Mode (One-Shot)
- Run once and exit.
- Behavior:
  1. Find first video matching filter regex.
  2. Download it if missing.
  3. Send it to Telegram immediately.

## Downloader Design

### Source
- Instagram profile (default: `gmbadass`).
- Data access via Instaloader CLI/API called from Go (`exec.Command`).

### Filtering
- Configurable regex against caption/description.
- Default regex: `(?i)(good ?morning|wake)`.

### Deduplication
- Before download, infer uniqueness from shortcode and target filename.
- Skip if target file already exists.

### Filename Convention
- Final file path:
  - `/data/video/<profile>_<YYYYMMDD>-<shortcode>.mp4`
- Date source: post date (UTC normalized in filename generation).

### Download Pacing
- Sleep 30 seconds after each successful download before processing next candidate.

## Sender Design

### Selection
- Build candidate list from `/data/video/*.mp4`.
- Maintain sent-history file (e.g. `/data/state/sent.log`) with stable media IDs derived from filename.
- Choose random from unsent set.
- If unsent set is empty, clear/reset history and select again from full set.

### Delivery
- Send video file via Telegram Bot API (`sendVideo`).
- Append to sent-history only on successful send.

## Configuration (Environment Variables)
- `IG_PROFILE` default `gmbadass`
- `DESCRIPTION_REGEX` default `(?i)(good ?morning|wake)`
- `DOWNLOAD_DIR` default `/data/video`
- `SENT_LOG_PATH` default `/data/state/sent.log`
- `DOWNLOAD_DELAY_SECONDS` default `30`
- `APP_CRON` default `0 6 * * *`
- `ONE_SHOT` default `false`
- `TELEGRAM_BOT_TOKEN` required
- `TELEGRAM_CHAT_ID` required (preferred explicit target)
- `LOG_LEVEL` default `info`

## Make Targets (Required)
- `make run`
  - Run local container in production mode via `docker run --rm ...`.
- `make dev`
  - One-shot mode: download first regex-matching video and send it.
- `make test`
  - Run Go tests.
- `make build`
  - Build multi-arch image (`linux/amd64,linux/arm64`).
- `make push`
  - Build + push multi-arch image to Docker Hub as `rrrrdockerrrr/gmb-bot:latest`.

Notes:
- No docker-compose usage.
- Local execution should mount `./gmb_data:/data`.

## Docker/Image Plan
- Single Dockerfile including:
  - Go binary build
  - Python + Instaloader runtime dependency
- Multi-arch buildx pipeline for `amd64` + `arm64`.
- Published image:
  - `rrrrdockerrrr/gmb-bot:latest`

## Operational Model

### Local Manual Run
- `make dev` for one-shot verification.
- `make run` for production-like scheduled behavior.

### NAS Deployment
- Pull image `rrrrdockerrrr/gmb-bot:latest`.
- Configure env vars in NAS container manager.
- Mount persistent volume to `/data`.
- Run container continuously (scheduler inside app handles daily 6 AM execution).

## Testing and Acceptance Criteria

### Tests
- Regex filter matching logic.
- Filename generation format and parsing.
- Dedup skip logic for existing files.
- Sent-history cycle behavior (no repeat until exhaustion, then reset).
- Dev one-shot path selects first matching video and sends it.

### Acceptance Criteria
- One container handles download + send.
- No browser automation dependency in code or image.
- Videos saved as `/data/video/<profile>_<YYYYMMDD>-<shortcode>.mp4`.
- Production mode runs daily at 6 AM by default.
- Dev mode does one-shot download+send of first regex match.
- Multi-arch image is buildable and pushable to `rrrrdockerrrr/gmb-bot:latest`.

## Short README Requirements
README must be concise and include only:
- Required env vars.
- Manual local runs (`make dev`, `make run`).
- Build and push commands (`make build`, `make push`).
- NAS setup steps:
  1. pull image,
  2. mount `/data`,
  3. set env,
  4. run container.
