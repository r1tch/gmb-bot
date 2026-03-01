#!/usr/bin/env python3
import argparse
import json
import sys
import time
from datetime import timezone

import instaloader


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--profile", required=True)
    parser.add_argument("--scan-limit", type=int, default=120)
    parser.add_argument("--sleep-seconds", type=float, default=0.0)
    args = parser.parse_args()

    loader = instaloader.Instaloader(
        download_pictures=False,
        download_videos=False,
        download_video_thumbnails=False,
        download_comments=False,
        save_metadata=False,
        compress_json=False,
        quiet=True,
    )

    profile = instaloader.Profile.from_username(loader.context, args.profile)
    posts = []
    for i, post in enumerate(profile.get_posts()):
        if i >= args.scan_limit:
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
        if args.sleep_seconds > 0 and (i + 1) % 5 == 0:
            time.sleep(args.sleep_seconds)

    print(json.dumps(posts, separators=(",", ":")))


if __name__ == "__main__":
    try:
        main()
    except BrokenPipeError:
        # keep script quiet when stdout consumer closes early
        sys.exit(0)
