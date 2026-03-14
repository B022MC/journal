import importlib.util
import sys
import unittest
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().with_name("check_remote_journal_validation_env.py")
SPEC = importlib.util.spec_from_file_location("check_remote_journal_validation_env", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = MODULE
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


def valid_values() -> dict[str, str]:
    return {
        "REMOTE_JOURNAL_DSN": "journal:redacted@tcp(remote-host:3306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
        "REMOTE_VALIDATION_OWNER": "codex",
        "REMOTE_VALIDATION_DATE": "2026-03-14",
        "REMOTE_TEST_USER_EMAIL": "codex+remote-validation-20260314@example.invalid",
        "REMOTE_TEST_USER_NAME": "remote-validation-20260314",
        "REMOTE_TEST_USER_PASSWORD": "frontend-secret",
        "REMOTE_TEST_ADMIN_LOGIN": "rv-admin-20260314",
        "REMOTE_TEST_ADMIN_PASSWORD": "admin-secret",
        "REMOTE_TEST_ROLE_CODE": "rv_tmp_role_20260314",
        "REMOTE_TEST_ROLE_NAME": "RV Temporary Role 20260314",
        "REMOTE_TEST_KEYWORD_PATTERN": "rv_tmp_keyword_20260314",
        "REMOTE_TEST_NEWS_TITLE": "RV temporary news 20260314",
        "REMOTE_TEST_PAPER_ID": "paper-20260314-sample",
        "REMOTE_CLEANUP_OWNER": "codex",
    }


class CheckRemoteJournalValidationEnvTest(unittest.TestCase):
    def test_validate_values_accepts_complete_remote_fixture(self) -> None:
        result = MODULE.validate_values(valid_values())

        self.assertEqual(result.errors, [])
        self.assertEqual(result.warnings, [])

    def test_validate_values_rejects_missing_placeholder_and_local_dsn(self) -> None:
        values = valid_values()
        values["REMOTE_JOURNAL_DSN"] = "journal:redacted@tcp(127.0.0.1:13306)/journal"
        values["REMOTE_TEST_ADMIN_PASSWORD"] = "<redacted-via-local-env>"
        values["REMOTE_TEST_PAPER_ID"] = ""

        result = MODULE.validate_values(values)

        self.assertIn("local_dsn:REMOTE_JOURNAL_DSN contains 127.0.0.1", result.errors)
        self.assertIn("placeholder:REMOTE_TEST_ADMIN_PASSWORD", result.errors)
        self.assertIn("missing:REMOTE_TEST_PAPER_ID", result.errors)

    def test_validate_values_requires_expected_naming_rules(self) -> None:
        values = valid_values()
        values["REMOTE_TEST_USER_NAME"] = "demo-user"
        values["REMOTE_TEST_ROLE_CODE"] = "tmp_role_20260314"
        values["REMOTE_TEST_KEYWORD_PATTERN"] = "tmp_keyword_20260314"
        values["REMOTE_TEST_NEWS_TITLE"] = "RV temporary news"

        result = MODULE.validate_values(values)

        self.assertIn("naming:REMOTE_TEST_USER_NAME must start with remote-validation-", result.errors)
        self.assertIn("naming:REMOTE_TEST_ROLE_CODE must start with rv_tmp_", result.errors)
        self.assertIn("naming:REMOTE_TEST_KEYWORD_PATTERN must start with rv_tmp_", result.errors)
        self.assertIn("date_token:REMOTE_TEST_NEWS_TITLE must include 20260314", result.errors)

    def test_validate_values_warns_when_cleanup_owner_differs(self) -> None:
        values = valid_values()
        values["REMOTE_CLEANUP_OWNER"] = "release-owner"

        result = MODULE.validate_values(values)

        self.assertEqual(result.errors, [])
        self.assertEqual(
            result.warnings,
            ["REMOTE_CLEANUP_OWNER differs from REMOTE_VALIDATION_OWNER"],
        )


if __name__ == "__main__":
    unittest.main()
