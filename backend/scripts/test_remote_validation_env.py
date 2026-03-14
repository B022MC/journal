import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().with_name("remote_validation_env.py")
SPEC = importlib.util.spec_from_file_location("remote_validation_env", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = MODULE
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


class RemoteValidationEnvTest(unittest.TestCase):
    def test_parse_env_file_reads_simple_key_values(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            env_file = Path(temp_dir) / ".env.remote-validation.local"
            env_file.write_text(
                '\n'.join(
                    [
                        "# comment",
                        'REMOTE_JOURNAL_DSN="journal:redacted@tcp(remote-host:3306)/journal"',
                        "REMOTE_VALIDATION_OWNER=b022mc",
                    ]
                ),
                encoding="utf-8",
            )

            values = MODULE.parse_env_file(env_file)

            self.assertEqual(
                values["REMOTE_JOURNAL_DSN"],
                "journal:redacted@tcp(remote-host:3306)/journal",
            )
            self.assertEqual(values["REMOTE_VALIDATION_OWNER"], "b022mc")

    def test_load_values_prefers_process_env_over_file(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            env_file = Path(temp_dir) / ".env.remote-validation.local"
            env_file.write_text(
                "REMOTE_VALIDATION_OWNER=file-owner\nREMOTE_CLEANUP_OWNER=file-owner\n",
                encoding="utf-8",
            )

            old_owner = MODULE.os.environ.get("REMOTE_VALIDATION_OWNER")
            MODULE.os.environ["REMOTE_VALIDATION_OWNER"] = "shell-owner"
            try:
                values = MODULE.load_values(
                    ("REMOTE_VALIDATION_OWNER", "REMOTE_CLEANUP_OWNER"),
                    env_file,
                )
            finally:
                if old_owner is None:
                    MODULE.os.environ.pop("REMOTE_VALIDATION_OWNER", None)
                else:
                    MODULE.os.environ["REMOTE_VALIDATION_OWNER"] = old_owner

            self.assertEqual(values["REMOTE_VALIDATION_OWNER"], "shell-owner")
            self.assertEqual(values["REMOTE_CLEANUP_OWNER"], "file-owner")


if __name__ == "__main__":
    unittest.main()
