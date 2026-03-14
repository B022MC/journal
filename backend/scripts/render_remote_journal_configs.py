#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class ServiceConfig:
    source: str
    output: str


SERVICE_CONFIGS = (
    ServiceConfig("api/etc/journal-api.yaml", "api/etc/journal-api.remote.yaml"),
    ServiceConfig("admin-api/etc/admin-api.yaml", "admin-api/etc/admin-api.remote.yaml"),
    ServiceConfig("rpc/user/etc/user.yaml", "rpc/user/etc/user.remote.yaml"),
    ServiceConfig("rpc/paper/etc/paper.yaml", "rpc/paper/etc/paper.remote.yaml"),
    ServiceConfig("rpc/rating/etc/rating.yaml", "rpc/rating/etc/rating.remote.yaml"),
    ServiceConfig("rpc/news/etc/news.yaml", "rpc/news/etc/news.remote.yaml"),
    ServiceConfig("rpc/admin/etc/admin.yaml", "rpc/admin/etc/admin.remote.yaml"),
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render remote single-db configs for all journal services."
    )
    parser.add_argument(
        "--dsn",
        default=os.environ.get("REMOTE_JOURNAL_DSN", ""),
        help="Remote journal DSN. Defaults to REMOTE_JOURNAL_DSN.",
    )
    parser.add_argument(
        "--output-dir",
        default="/tmp/journal-remote-validation",
        help="Directory for rendered remote configs.",
    )
    return parser.parse_args()


def rewrite_db_block(text: str, dsn: str) -> str:
    lines = text.splitlines()
    output: list[str] = []
    index = 0
    found_db = False

    while index < len(lines):
        line = lines[index]
        if line == "DB:":
            found_db = True
            output.append("DB:")
            output.append("  ReadWriteSplit: false")
            output.append(f'  DataSource: "{dsn}"')
            index += 1
            skipping_replicas = False
            while index < len(lines):
                child = lines[index]
                if child and not child.startswith("  "):
                    break
                stripped = child.strip()
                if not stripped:
                    index += 1
                    continue
                if skipping_replicas:
                    if child.startswith("    -"):
                        index += 1
                        continue
                    skipping_replicas = False
                if stripped.startswith("ReadWriteSplit:") or stripped.startswith("DataSource:") or stripped.startswith("Policy:"):
                    index += 1
                    continue
                if stripped.startswith("Replicas:"):
                    skipping_replicas = True
                    index += 1
                    continue
                output.append(child)
                index += 1
            continue

        output.append(line)
        index += 1

    if not found_db:
        raise ValueError("DB block not found")

    return "\n".join(output) + "\n"


def validate_rendered(text: str, dsn: str) -> None:
    required = (
        "ReadWriteSplit: false",
        f'DataSource: "{dsn}"',
    )
    forbidden = (
        "ReadWriteSplit: true",
        "127.0.0.1:13306",
        "127.0.0.1:13307",
        "127.0.0.1:13308",
        "Replicas:",
        "Policy:",
    )
    for marker in required:
        if marker not in text:
            raise ValueError(f"missing marker: {marker}")
    for marker in forbidden:
        if marker in text:
            raise ValueError(f"forbidden marker still present: {marker}")


def render_config(backend_root: Path, service: ServiceConfig, output_root: Path, dsn: str) -> Path:
    source = backend_root / service.source
    target = output_root / service.output
    text = source.read_text(encoding="utf-8")
    rendered = rewrite_db_block(text, dsn)
    validate_rendered(rendered, dsn)
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(rendered, encoding="utf-8")
    return target


def main() -> int:
    args = parse_args()
    if not args.dsn:
        raise SystemExit("missing --dsn and REMOTE_JOURNAL_DSN")

    backend_root = Path(__file__).resolve().parents[1]
    output_root = Path(args.output_dir)
    rendered_paths = [
        render_config(backend_root, service, output_root, args.dsn)
        for service in SERVICE_CONFIGS
    ]

    for path in rendered_paths:
        print(path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
