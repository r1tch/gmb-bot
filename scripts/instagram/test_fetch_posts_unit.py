import pathlib
import subprocess
import sys
import unittest


SCRIPT = pathlib.Path(__file__).with_name("fetch_posts.py")


class FetchPostsUnitTests(unittest.TestCase):
    def test_script_help_sanity(self):
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), "--help"],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(proc.returncode, 0, msg=proc.stderr)
        self.assertIn("--profile", proc.stdout)


if __name__ == "__main__":
    unittest.main()
