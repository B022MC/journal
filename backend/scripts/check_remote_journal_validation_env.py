#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import sys
from dataclasses import dataclass


REQUIRED_KEYS = (
    "REMOTE_JOURNAL_DSN",
    "REMOTE_VALIDATION_OWNER",
    "REMOTE_VALIDATION_DATE",
    "REMOTE_TEST_USER_EMAIL",
    "REMOTE_TEST_USER_NAME",
    "REMOTE_TEST_USER_PASSWORD",
    "REMOTE_TEST_ADMIN_LOGIN",
    "REMOTE_TEST_ADMIN_PASSWORD",
    "REMOTE_TEST_ROLE_CODE",
    "REMOTE_TEST_ROLE_NAME",
    "REMOTE_TEST_KEYWORD_PATTERN",
    "REMOTE_TEST_NEWS_TITLE",
    "REMOTE_TEST_PAPER_ID",
    "REMOTE_CLEANUP_OWNER",
)

LOCAL_DSN_MARKERS = (
    "127.0.0.1",
    "localhost",
    ":13306",
    ":13307",
    ":13308",
)

PLACEHOLDER_MARKERS = ("<", ">")


@dataclass(frozen=True)
class ValidationResult:
    errors: list[str]
    warnings: list[str]


def load_values() -> dict[str, str]:
    return {key: os.environ.get(key, "").strip() for key in REQUIRED_KEYS}


def has_placeholder(value: str) -> bool:
    return any(marker in value for marker in PLACEHOLDER_MARKERS)


def validate_values(values: dict[str, str]) -> ValidationResult:
    errors: list[str] = []
    warnings: list[str] = []

    for key, value in values.items():
        if not value:
            errors.append(f"missing:{key}")
            continue
        if has_placeholder(value):
            errors.append(f"placeholder:{key}")

    date_token = ""
    validation_date = values.get("REMOTE_VALIDATION_DATE", "")
    if validation_date and not has_placeholder(validation_date):
        if not re.fullmatch(r"\d{4}-\d{2}-\d{2}", validation_date):
            errors.append("format:REMOTE_VALIDATION_DATE must use YYYY-MM-DD")
        else:
            date_token = validation_date.replace("-", "")

    dsn = values.get("REMOTE_JOURNAL_DSN", "")
    if dsn and not has_placeholder(dsn):
        for marker in LOCAL_DSN_MARKERS:
            if marker in dsn:
                errors.append(f"local_dsn:REMOTE_JOURNAL_DSN contains {marker}")
                break

    if values.get("REMOTE_TEST_USER_NAME") and not values["REMOTE_TEST_USER_NAME"].startswith("remote-validation-"):
        errors.append("naming:REMOTE_TEST_USER_NAME must start with remote-validation-")
    if values.get("REMOTE_TEST_ROLE_CODE") and not values["REMOTE_TEST_ROLE_CODE"].startswith("rv_tmp_"):
        errors.append("naming:REMOTE_TEST_ROLE_CODE must start with rv_tmp_")
    if values.get("REMOTE_TEST_KEYWORD_PATTERN") and not values["REMOTE_TEST_KEYWORD_PATTERN"].startswith("rv_tmp_"):
        errors.append("naming:REMOTE_TEST_KEYWORD_PATTERN must start with rv_tmp_")

    if date_token:
        for key in (
            "REMOTE_TEST_USER_EMAIL",
            "REMOTE_TEST_USER_NAME",
            "REMOTE_TEST_ROLE_CODE",
            "REMOTE_TEST_KEYWORD_PATTERN",
            "REMOTE_TEST_NEWS_TITLE",
        ):
            value = values.get(key, "")
            if value and not has_placeholder(value) and date_token not in value:
                errors.append(f"date_token:{key} must include {date_token}")

    cleanup_owner = values.get("REMOTE_CLEANUP_OWNER", "")
    validation_owner = values.get("REMOTE_VALIDATION_OWNER", "")
    if cleanup_owner and validation_owner and cleanup_owner != validation_owner:
        warnings.append("REMOTE_CLEANUP_OWNER differs from REMOTE_VALIDATION_OWNER")

    return ValidationResult(errors=errors, warnings=warnings)


def print_report(result: ValidationResult) -> None:
    if result.errors:
        print("Remote journal validation env is incomplete:")
        for error in result.errors:
            print(f"- {error}")
    else:
        print("Remote journal validation env is ready.")

    for warning in result.warnings:
        print(f"- warning:{warning}")


def main() -> int:
    result = validate_values(load_values())
    print_report(result)
    return 1 if result.errors else 0


if __name__ == "__main__":
    sys.exit(main())
