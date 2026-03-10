#!/usr/bin/env python3
import argparse
import json
import re
import sys
import time
from datetime import timezone

import instaloader

RETRY_BACKOFF_SECONDS = (10, 30, 60, 300, 1800)


def _is_rate_limited(error: Exception) -> bool:
    response = getattr(error, "response", None)
    status_code = getattr(response, "status_code", None)
    if status_code == 401:
        return True
    # Instaloader commonly embeds the status code in exception text.
    return re.search(r"(^|\\D)401(\\D|$)", str(error)) is not None


def _fetch_posts(profile_name: str, scan_limit: int, sleep_seconds: float) -> list[dict]:
    loader = instaloader.Instaloader(
        download_pictures=False,
        download_videos=False,
        download_video_thumbnails=False,
        download_comments=False,
        save_metadata=False,
        compress_json=False,
        quiet=True,
    )

    profile = instaloader.Profile.from_username(loader.context, profile_name)
    posts = []
    for i, post in enumerate(profile.get_posts()):
        if i >= scan_limit:
            break
        dt = post.date_utc
        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)
        posts.append(
            {
                "shortcode": post.shortcode,
                "caption": post.caption or "",
                "video_url": post.video_url if post.is_video else "",
                "date_utc": dt.isoformat(),
                "is_video": bool(post.is_video),
            }
        )
        if sleep_seconds > 0 and (i + 1) % 5 == 0:
            time.sleep(sleep_seconds)
    return posts


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--profile", required=True)
    parser.add_argument("--scan-limit", type=int, default=120)
    parser.add_argument("--sleep-seconds", type=float, default=0.0)
    args = parser.parse_args()

    attempts = len(RETRY_BACKOFF_SECONDS) + 1
    for attempt in range(1, attempts + 1):
        try:
            posts = _fetch_posts(args.profile, args.scan_limit, args.sleep_seconds)
            print(json.dumps(posts, separators=(",", ":")))
            return
        except Exception as exc:
            if not _is_rate_limited(exc):
                raise
            if attempt >= attempts:
                raise
            wait_seconds = RETRY_BACKOFF_SECONDS[attempt - 1]
            print(
                f"[fetch_posts] rate-limited, retrying after {wait_seconds}s "
                f"(attempt {attempt}/{attempts})",
                file=sys.stderr,
            )
            time.sleep(wait_seconds)


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        # keep script quiet when stdout consumer closes early
        sys.exit(0)
