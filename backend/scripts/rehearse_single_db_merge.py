#!/usr/bin/env python3
from __future__ import annotations

import argparse
from dataclasses import dataclass
from typing import Iterable


@dataclass(frozen=True)
class TableMapping:
    source_schema: str
    source_table: str
    target_table: str


BUSINESS_MAPPINGS = (
    TableMapping("journal_biz", "user", "biz_user"),
    TableMapping("journal_biz", "user_achievement", "biz_user_achievement"),
    TableMapping("journal_biz", "paper", "biz_paper"),
    TableMapping("journal_biz", "cold_paper", "biz_cold_paper"),
    TableMapping("journal_biz", "rating", "biz_rating"),
    TableMapping("journal_biz", "news", "biz_news"),
    TableMapping("journal_biz", "flag", "biz_flag"),
    TableMapping("journal_biz", "keyword_rule", "biz_keyword_rule"),
)

ADMIN_MAPPINGS = (
    TableMapping("journal_admin", "adm_role", "adm_role"),
    TableMapping("journal_admin", "adm_permission", "adm_permission"),
    TableMapping("journal_admin", "adm_role_permission", "adm_role_permission"),
    TableMapping("journal_admin", "adm_user_role", "adm_user_role"),
    TableMapping("journal_admin", "adm_audit_log", "adm_audit_log"),
)

ALL_MAPPINGS = BUSINESS_MAPPINGS + ADMIN_MAPPINGS


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render a reversible single-db merge rehearsal plan for journal -> journal.biz_*/adm_*."
    )
    parser.add_argument(
        "--phase",
        choices=("all", "bootstrap", "backfill", "verify", "rollback"),
        default="all",
        help="Which section of the rehearsal plan to render.",
    )
    return parser.parse_args()


def render_header() -> list[str]:
    return [
        "-- Single-DB merge rehearsal plan",
        "-- Strategy: create target tables in `journal`, backfill by copy, create compatibility views, switch apps, keep old schemas read-only for rollback.",
        "-- Default mode is dry-run: review this output, then execute selected sections with mysql in rehearsal/prod windows.",
        "SET SESSION sql_log_bin = 0;",
        "CREATE DATABASE IF NOT EXISTS `journal` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;",
        "GRANT ALL PRIVILEGES ON `journal`.* TO 'journal'@'%';",
        "FLUSH PRIVILEGES;",
        "",
    ]


def render_bootstrap(mappings: Iterable[TableMapping]) -> list[str]:
    lines = ["-- Phase 1: bootstrap target prefixed tables", ""]
    for mapping in mappings:
        lines.extend(
            [
                f"CREATE TABLE IF NOT EXISTS `journal`.`{mapping.target_table}` LIKE `{mapping.source_schema}`.`{mapping.source_table}`;",
            ]
        )
    lines.append("")
    return lines


def render_backfill(mappings: Iterable[TableMapping]) -> list[str]:
    lines = [
        "-- Phase 2: backfill into target tables without renaming old schemas",
        "-- Re-run during rehearsal after truncating target tables, or during cutover after an application write freeze.",
        "-- If zero-downtime is required, enable temporary dual-write before the final delta copy instead of renaming tables.",
        "",
    ]
    for mapping in mappings:
        lines.extend(
            [
                f"INSERT INTO `journal`.`{mapping.target_table}`",
                f"SELECT * FROM `{mapping.source_schema}`.`{mapping.source_table}`",
                "ON DUPLICATE KEY UPDATE `id` = `id`;",
                "",
            ]
        )
    return lines


def render_verify(mappings: Iterable[TableMapping]) -> list[str]:
    lines = [
        "-- Phase 4: reconcile row counts and primary-key ranges",
        "-- Go/no-go: source and target row_count/min_id/max_id must match for every table before application cutover.",
        "",
    ]
    for mapping in mappings:
        label = mapping.target_table
        lines.extend(
            [
                f"SELECT '{label}' AS table_name, 'source' AS side, COUNT(*) AS row_count, MIN(`id`) AS min_id, MAX(`id`) AS max_id",
                f"FROM `{mapping.source_schema}`.`{mapping.source_table}`",
                "UNION ALL",
                f"SELECT '{label}' AS table_name, 'target' AS side, COUNT(*) AS row_count, MIN(`id`) AS min_id, MAX(`id`) AS max_id",
                f"FROM `journal`.`{mapping.target_table}`;",
                "",
            ]
        )
    return lines


def render_compat_views(mappings: Iterable[TableMapping]) -> list[str]:
    lines = [
        "-- Phase 3: create compatibility views for legacy business table names",
        "-- Keep these views until every runtime SQL path has been moved to the prefixed tables.",
        "",
    ]
    for mapping in mappings:
        if mapping.source_schema != "journal_biz":
            continue
        lines.extend(
            [
                f"CREATE OR REPLACE VIEW `journal`.`{mapping.source_table}` AS",
                f"SELECT * FROM `journal`.`{mapping.target_table}`;",
                "",
            ]
        )
    return lines


def render_rollback(mappings: Iterable[TableMapping]) -> list[str]:
    lines = [
        "-- Phase 5: rollback notes",
        "-- 1. Keep `journal_biz` and `journal_admin` intact and read-only until post-cutover validation passes.",
        "-- 2. If cutover fails, disable read/write split first, then switch DSNs back to the old schemas.",
        "-- 3. Retain the compatibility views only while validating the cutover; remove them after rollback if they hide drift.",
        "-- 4. Do not drop target prefixed tables during rollback; preserve them for diffing and a later retry.",
        "-- 5. After rollback, compare the same verification queries below before re-opening writes.",
        "",
    ]
    for mapping in mappings:
        lines.append(
            f"-- rollback-check {mapping.source_schema}.{mapping.source_table} <-> journal.{mapping.target_table}"
        )
    lines.append("")
    return lines


def main() -> int:
    args = parse_args()
    sections: list[str] = []

    if args.phase == "all":
        sections.extend(render_header())
        sections.extend(render_bootstrap(ALL_MAPPINGS))
        sections.extend(render_backfill(ALL_MAPPINGS))
        sections.extend(render_compat_views(BUSINESS_MAPPINGS))
        sections.extend(render_verify(ALL_MAPPINGS))
        sections.extend(render_rollback(ALL_MAPPINGS))
    elif args.phase == "bootstrap":
        sections.extend(render_header())
        sections.extend(render_bootstrap(ALL_MAPPINGS))
    elif args.phase == "backfill":
        sections.extend(render_backfill(ALL_MAPPINGS))
    elif args.phase == "verify":
        sections.extend(render_compat_views(BUSINESS_MAPPINGS))
        sections.extend(render_verify(ALL_MAPPINGS))
    elif args.phase == "rollback":
        sections.extend(render_rollback(ALL_MAPPINGS))

    print("\n".join(sections))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
