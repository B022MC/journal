#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import io
import re
import sys
from collections import Counter, defaultdict
from pathlib import Path

BASELINE_NAME = "2026-03-13-db-merge-legacy-baseline.csv"
SCAN_ROOTS = (
    "api",
    "admin-api",
    "rpc",
    "cmd",
    "common",
    "model",
    "deploy",
    "docker-compose.yaml",
)
TEXT_SUFFIXES = {".go", ".sql", ".yaml", ".yml", ".json", ".md", ".sh", ".txt"}
IGNORE_NAMES = {"journal.json"}
IGNORE_SUFFIXES = {".pb.go"}
PATTERNS = (
    ("legacy_schema", re.compile(r"\b(journal_biz|journal_admin)\b")),
    ("legacy_config_key", re.compile(r"\b(BizDB|AdminDB)\b")),
    ("legacy_table_sql", re.compile(r"`(user_achievement|keyword_rule|cold_paper|rating|paper|news|flag|user)`")),
    (
        "legacy_table_sql",
        re.compile(
            r"(?i)\b(?:from|join|into|update|table|use)\s+`?(user_achievement|keyword_rule|cold_paper|rating|paper|news|flag|user)`?(?=[\s`(.;,]|$)"
        ),
    ),
)


def parse_args() -> argparse.Namespace:
    backend_root = Path(__file__).resolve().parents[1]
    parser = argparse.ArgumentParser(
        description="Guard the DB merge freeze window by rejecting new legacy schema or table references."
    )
    parser.add_argument(
        "--baseline",
        type=Path,
        default=backend_root / "docs" / "release" / BASELINE_NAME,
        help="Path to the audited legacy baseline CSV.",
    )
    parser.add_argument(
        "--dump-current",
        action="store_true",
        help="Print the current scan result as baseline CSV and exit.",
    )
    return parser.parse_args()


def should_scan(path: Path) -> bool:
    if path.suffix not in TEXT_SUFFIXES:
        return False
    if any(path.name.endswith(suffix) for suffix in IGNORE_SUFFIXES):
        return False
    if path.name in IGNORE_NAMES or path.name.endswith("_test.go"):
        return False
    if "docs" in path.parts:
        return False
    return True


def scan_current(backend_root: Path) -> tuple[Counter, dict[tuple[str, str, str], list[str]]]:
    counts: Counter = Counter()
    samples: dict[tuple[str, str, str], list[str]] = defaultdict(list)

    for root_name in SCAN_ROOTS:
        root = backend_root / root_name
        paths = [root] if root.is_file() else sorted(p for p in root.rglob("*") if p.is_file())
        for path in paths:
            if not should_scan(path):
                continue
            rel_path = path.relative_to(backend_root).as_posix()
            text = path.read_text(encoding="utf-8", errors="ignore")
            for line in text.splitlines():
                for kind, pattern in PATTERNS:
                    for match in pattern.finditer(line):
                        token = match.group(1)
                        key = (rel_path, kind, token)
                        counts[key] += 1
                        excerpt = line.strip()
                        if excerpt not in samples[key]:
                            samples[key].append(excerpt)
    return counts, samples


def load_baseline(path: Path) -> Counter:
    if not path.exists():
        raise FileNotFoundError(f"baseline not found: {path}")

    counts: Counter = Counter()
    with path.open("r", encoding="utf-8", newline="") as csv_file:
        reader = csv.DictReader(csv_file)
        expected = {"path", "kind", "token", "count"}
        if set(reader.fieldnames or ()) != expected:
            raise ValueError(f"invalid baseline headers: expected {sorted(expected)}, got {reader.fieldnames}")
        for row in reader:
            key = (row["path"], row["kind"], row["token"])
            counts[key] = int(row["count"])
    return counts


def render_csv(counts: Counter) -> str:
    output = io.StringIO()
    writer = csv.writer(output)
    writer.writerow(["path", "kind", "token", "count"])
    for rel_path, kind, token in sorted(counts):
        writer.writerow([rel_path, kind, token, counts[(rel_path, kind, token)]])
    return output.getvalue()


def main() -> int:
    args = parse_args()
    backend_root = Path(__file__).resolve().parents[1]
    current, samples = scan_current(backend_root)

    if args.dump_current:
        sys.stdout.write(render_csv(current))
        return 0

    baseline = load_baseline(args.baseline)
    violations = []
    for key, current_count in sorted(current.items()):
        baseline_count = baseline.get(key, 0)
        if current_count > baseline_count:
            violations.append((key, baseline_count, current_count))

    if violations:
        print("db naming freeze violation(s) detected:")
        for (rel_path, kind, token), baseline_count, current_count in violations:
            print(f"- {rel_path} [{kind}:{token}] baseline={baseline_count} current={current_count}")
            for excerpt in samples[(rel_path, kind, token)][:3]:
                print(f"    {excerpt}")
        print(f"baseline: {args.baseline}")
        return 1

    total_hits = sum(current.values())
    print(
        f"db naming freeze OK: {len(current)} tracked tuples, {total_hits} legacy hits remain within baseline {args.baseline.name}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
