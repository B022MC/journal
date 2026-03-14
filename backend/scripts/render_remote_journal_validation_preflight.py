#!/usr/bin/env python3
from __future__ import annotations

import argparse
from dataclasses import dataclass


@dataclass(frozen=True)
class ExpectedObject:
    name: str
    object_type: str


BUSINESS_VIEWS = (
    ExpectedObject("user", "VIEW"),
    ExpectedObject("paper", "VIEW"),
    ExpectedObject("rating", "VIEW"),
    ExpectedObject("news", "VIEW"),
    ExpectedObject("flag", "VIEW"),
    ExpectedObject("keyword_rule", "VIEW"),
)

ADMIN_TABLES = (
    ExpectedObject("adm_role", "BASE TABLE"),
    ExpectedObject("adm_permission", "BASE TABLE"),
    ExpectedObject("adm_user_role", "BASE TABLE"),
    ExpectedObject("adm_audit_log", "BASE TABLE"),
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render read-only preflight SQL for remote journal validation."
    )
    parser.add_argument(
        "--schema",
        default="journal",
        help="Schema name to validate. Defaults to journal.",
    )
    parser.add_argument(
        "--label",
        default="rv_tmp_20260314",
        help="Disposable label prefix used by cleanup query templates.",
    )
    return parser.parse_args()


def quote_csv(objects: tuple[ExpectedObject, ...]) -> str:
    return ", ".join(f"'{obj.name}'" for obj in objects)


def render_header(schema: str, label: str) -> list[str]:
    return [
        "-- Remote journal validation preflight",
        "-- This workbook is read-only except for cleanup templates that are commented out.",
        f"-- Schema: {schema}",
        f"-- Disposable label prefix: {label}",
        "SELECT DATABASE() AS current_schema, CURRENT_USER() AS current_user;",
        "",
    ]


def render_inventory(schema: str) -> list[str]:
    view_names = quote_csv(BUSINESS_VIEWS)
    table_names = quote_csv(ADMIN_TABLES)
    return [
        "-- Inventory: compatibility views must exist and remain updatable.",
        "SELECT table_name, table_type, is_updatable",
        "FROM information_schema.views",
        f"WHERE table_schema = '{schema}'",
        f"  AND table_name IN ({view_names})",
        "ORDER BY table_name;",
        "",
        "-- Inventory: admin base tables must exist as real tables.",
        "SELECT table_name, table_type",
        "FROM information_schema.tables",
        f"WHERE table_schema = '{schema}'",
        f"  AND table_name IN ({table_names})",
        "ORDER BY table_name;",
        "",
    ]


def render_read_probes(schema: str) -> list[str]:
    lines = ["-- Read probes: each statement should return zero or one row without object-not-found errors.", ""]
    for obj in BUSINESS_VIEWS + ADMIN_TABLES:
        lines.extend(
            [
                f"SELECT '{obj.name}' AS object_name, 1 AS readable",
                f"FROM `{schema}`.`{obj.name}`",
                "LIMIT 1;",
                "",
            ]
        )
    return lines


def render_cleanup_templates(schema: str, label: str) -> list[str]:
    like = f"{label}%"
    return [
        "-- Cleanup templates for disposable fixtures. Review ids first, then uncomment only the rows created by the current validation window.",
        f"-- SELECT id, email, username FROM `{schema}`.`user` WHERE email LIKE '{like}' OR username LIKE '{like}';",
        f"-- SELECT id, code, name FROM `{schema}`.`adm_role` WHERE code LIKE '{like}' OR name LIKE '{like}';",
        f"-- SELECT id, pattern FROM `{schema}`.`keyword_rule` WHERE pattern LIKE '{like}';",
        f"-- SELECT id, title FROM `{schema}`.`news` WHERE title LIKE '{like}';",
        f"-- SELECT id, target_type, target_id FROM `{schema}`.`flag` WHERE detail LIKE '{like}';",
        f"-- DELETE FROM `{schema}`.`adm_user_role` WHERE role_id IN (SELECT id FROM `{schema}`.`adm_role` WHERE code LIKE '{like}');",
        f"-- DELETE FROM `{schema}`.`adm_role` WHERE code LIKE '{like}';",
        f"-- DELETE FROM `{schema}`.`keyword_rule` WHERE pattern LIKE '{like}';",
        f"-- DELETE FROM `{schema}`.`news` WHERE title LIKE '{like}';",
        f"-- DELETE FROM `{schema}`.`user` WHERE email LIKE '{like}' OR username LIKE '{like}';",
        "-- Keep `adm_audit_log` as evidence; query it by time range instead of deleting historical rows.",
        "",
    ]


def main() -> int:
    args = parse_args()
    lines: list[str] = []
    lines.extend(render_header(args.schema, args.label))
    lines.extend(render_inventory(args.schema))
    lines.extend(render_read_probes(args.schema))
    lines.extend(render_cleanup_templates(args.schema, args.label))
    print("\n".join(lines))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
