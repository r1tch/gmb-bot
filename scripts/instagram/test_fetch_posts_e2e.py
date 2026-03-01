import json
import pathlib
import subprocess
import sys
import unittest


SCRIPT = pathlib.Path(__file__).with_name("fetch_posts.py")


class FetchPostsE2ETests(unittest.TestCase):
    def test_lists_10_posts_from_gmbadass(self):
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), "--profile", "gmbadass", "--scan-limit", "10"],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(proc.returncode, 0, msg=proc.stderr)
        payload = json.loads(proc.stdout)
        self.assertIsInstance(payload, list)
        self.assertEqual(len(payload), 10)


if __name__ == "__main__":
    unittest.main()
