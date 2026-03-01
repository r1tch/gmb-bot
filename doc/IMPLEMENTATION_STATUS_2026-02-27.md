# TikTok-to-Telegram Bot: Current Status (2026-02-27)

codex resume 019c9f96-fd62-7a00-b092-91255321eabd

## Current State
- Containerized Go app is running and orchestrating the workflow.
- TikTok fetch is now implemented with plain Playwright (not Crawlee), and this is working better with TikTok.
- Debug artifacts are enabled and saved to `/data` (host bind mount `./gmb_data`):
  - rendered HTML (`tiktok-profile.html`)
  - screenshot (`tiktok-profile.png`)
- We successfully reached the phase of listing profile videos from `@gmbadass`.
- Video links are visible in logs/screenshots and extraction path is functioning.

## Important Operational Constraint
- Manual bootstrapping is currently required.
- Bootstrapping must be done inside the container runtime via VNC to preserve the same browser/runtime fingerprint.
- Current working flow:
  1. Run `make bootstrap-tiktok-container`
  2. Connect VNC client to `localhost:5900`
  3. Solve TikTok captcha/challenge manually
  4. Press Enter in terminal to save state to `./gmb_data/tiktok-storage-state.json`
  5. Run normal flow with `TIKTOK_STORAGE_STATE_PATH=/data/tiktok-storage-state.json`

## Next Steps
1. Improve description matching logic.
- Move from simple lowercase substring matching to regex-based matching.
- Default regex should be:
  - `(good ?morning|wake)`
- Keep regex configurable via environment variable.

2. Implement browser interaction for video download.
- Open candidate video page.
- Trigger right-click/context-menu based download flow for the video.
- Ensure Playwright waits for download completion.

3. Normalize downloaded artifact path.
- Download currently lands in a non-deterministic path under `/tmp/playwright-artifacts-.../<filename>`.
- Move the single downloaded file to deterministic target:
  - `/data/video.mp4`
- Assumption for now: only one downloaded file exists in that artifact directory.

4. Continue with Telegram delivery of the downloaded file.
- Use `/data/video.mp4` as source for Telegram `sendVideo` upload.
- Keep fallback to sending link if upload fails.

## Notes
- This status reflects that the major blocker (TikTok page access + listing videos) is now resolved under the bootstrap + state approach.
- Download and upload stages are the primary remaining implementation work.
