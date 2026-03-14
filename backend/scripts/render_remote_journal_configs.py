#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
from dataclasses import dataclass
from pathlib import Path

from remote_validation_env import DEFAULT_ENV_FILE, load_values


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

DIRECT_RPC_ENDPOINTS = {
    "UserRpc": "127.0.0.1:9001",
    "PaperRpc": "127.0.0.1:9002",
    "RatingRpc": "127.0.0.1:9003",
    "NewsRpc": "127.0.0.1:9004",
    "AdminRpc": "127.0.0.1:9005",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Render remote single-db configs for all journal services."
    )
    parser.add_argument(
        "--dsn",
        default="",
        help="Remote journal DSN. Defaults to REMOTE_JOURNAL_DSN.",
    )
    parser.add_argument(
        "--redis-pass",
        default="",
        help="Redis password for the remote validation profile.",
    )
    parser.add_argument(
        "--etcd-key-suffix",
        default="",
        help="Suffix appended to every rendered Etcd key to avoid live collisions.",
    )
    parser.add_argument(
        "--rpc-discovery-mode",
        choices=("direct", "etcd"),
        default="",
        help="Use direct localhost RPC endpoints or rendered Etcd keys.",
    )
    parser.add_argument(
        "--output-dir",
        default="/tmp/journal-remote-validation",
        help="Directory for rendered remote configs.",
    )
    parser.add_argument(
        "--env-file",
        default=str(DEFAULT_ENV_FILE),
        help="Path to the local env file used when --dsn is omitted.",
    )
    args = parser.parse_args()
    env_values = load_values(
        (
            "REMOTE_JOURNAL_DSN",
            "REMOTE_REDIS_PASS",
            "REMOTE_ETCD_KEY_SUFFIX",
            "REMOTE_RPC_DISCOVERY_MODE",
            "REMOTE_VALIDATION_OWNER",
            "REMOTE_VALIDATION_DATE",
        ),
        Path(args.env_file),
    )
    if not args.dsn:
        args.dsn = env_values["REMOTE_JOURNAL_DSN"]
    if not args.redis_pass:
        args.redis_pass = env_values["REMOTE_REDIS_PASS"]
    if not args.etcd_key_suffix:
        args.etcd_key_suffix = env_values["REMOTE_ETCD_KEY_SUFFIX"] or default_etcd_key_suffix(
            env_values["REMOTE_VALIDATION_OWNER"],
            env_values["REMOTE_VALIDATION_DATE"],
        )
    if not args.rpc_discovery_mode:
        args.rpc_discovery_mode = env_values["REMOTE_RPC_DISCOVERY_MODE"] or "direct"
    return args


def sanitize_suffix_token(value: str) -> str:
    cleaned = []
    for char in value.lower():
        if char.isalnum():
            cleaned.append(char)
        else:
            cleaned.append("-")
    token = "".join(cleaned).strip("-")
    return token or "remote-validation"


def default_etcd_key_suffix(owner: str, validation_date: str) -> str:
    owner_token = sanitize_suffix_token(owner) if owner else "remote-validation"
    date_token = sanitize_suffix_token(validation_date.replace("-", "")) if validation_date else ""
    if date_token:
        return f".{owner_token}.{date_token}"
    return f".{owner_token}"


def generated_targets(output_root: Path) -> list[Path]:
    return [output_root / service.output for service in SERVICE_CONFIGS]


def purge_generated_outputs(output_root: Path) -> None:
    for path in generated_targets(output_root):
        if path.exists():
            path.unlink()


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


def rewrite_redis_pass(text: str, redis_pass: str) -> str:
    if not redis_pass:
        return text

    lines = text.splitlines()
    output: list[str] = []
    section_stack: list[tuple[int, str]] = []

    for line in lines:
        stripped = line.strip()
        indent = len(line) - len(line.lstrip(" "))
        while section_stack and indent <= section_stack[-1][0]:
            section_stack.pop()

        if stripped.endswith(":") and not stripped.startswith("- "):
            section_stack.append((indent, stripped[:-1]))
            output.append(line)
            continue

        path = [name for _, name in section_stack]
        if path in (["Redis"], ["CacheRedis"]) and stripped.startswith("Pass:"):
            output.append(f'{" " * indent}Pass: "{redis_pass}"')
            continue
        output.append(line)

    return "\n".join(output) + "\n"


def suffixed_etcd_key(key: str, suffix: str) -> str:
    if not suffix or key.endswith(suffix):
        return key
    return f"{key}{suffix}"


def rewrite_etcd_keys(text: str, suffix: str) -> str:
    if not suffix:
        return text

    lines = text.splitlines()
    output: list[str] = []
    section_stack: list[tuple[int, str]] = []

    for line in lines:
        stripped = line.strip()
        indent = len(line) - len(line.lstrip(" "))
        while section_stack and indent <= section_stack[-1][0]:
            section_stack.pop()

        if stripped.endswith(":") and not stripped.startswith("- "):
            section_stack.append((indent, stripped[:-1]))
            output.append(line)
            continue

        path = [name for _, name in section_stack]
        if path and path[-1] == "Etcd" and stripped.startswith("Key:"):
            key = stripped.split(":", 1)[1].strip().strip('"').strip("'")
            output.append(f'{" " * indent}Key: {suffixed_etcd_key(key, suffix)}')
            continue
        output.append(line)

    return "\n".join(output) + "\n"


def rewrite_rpc_server_to_direct(text: str) -> str:
    lines = text.splitlines()
    output: list[str] = []
    index = 0

    while index < len(lines):
        line = lines[index]
        if line == "Etcd:":
            index += 1
            while index < len(lines):
                child = lines[index]
                if child and not child.startswith("  "):
                    break
                index += 1
            continue

        output.append(line)
        index += 1

    return "\n".join(output) + "\n"


def rewrite_rpc_clients_to_direct(text: str) -> str:
    lines = text.splitlines()
    output: list[str] = []
    index = 0

    while index < len(lines):
        line = lines[index]
        stripped = line.strip()
        if stripped.endswith(":") and stripped[:-1] in DIRECT_RPC_ENDPOINTS and not line.startswith(" "):
            block_name = stripped[:-1]
            endpoint = DIRECT_RPC_ENDPOINTS[block_name]
            output.append(line)
            index += 1
            inserted_endpoints = False

            while index < len(lines):
                child = lines[index]
                if child and not child.startswith("  "):
                    break

                child_stripped = child.strip()
                if child.startswith("  Etcd:"):
                    if not inserted_endpoints:
                        output.append("  Endpoints:")
                        output.append(f"    - {endpoint}")
                        inserted_endpoints = True
                    index += 1
                    while index < len(lines):
                        grandchild = lines[index]
                        if grandchild and not grandchild.startswith("    "):
                            break
                        index += 1
                    continue

                if not inserted_endpoints and child_stripped.startswith("Timeout:"):
                    output.append("  Endpoints:")
                    output.append(f"    - {endpoint}")
                    inserted_endpoints = True

                output.append(child)
                index += 1

            if not inserted_endpoints:
                output.append("  Endpoints:")
                output.append(f"    - {endpoint}")
            continue

        output.append(line)
        index += 1

    return "\n".join(output) + "\n"


def validate_rendered(
    text: str,
    service: ServiceConfig,
    dsn: str,
    redis_pass: str,
    etcd_key_suffix: str,
    rpc_discovery_mode: str,
) -> None:
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
    if redis_pass and "Redis:" in text and f'Pass: "{redis_pass}"' not in text:
        raise ValueError("missing rendered redis password")
    if rpc_discovery_mode == "direct":
        if "\nEtcd:\n" in text:
            raise ValueError("unexpected Etcd block in direct RPC mode")
        for endpoint in DIRECT_RPC_ENDPOINTS.values():
            if service.source.startswith(("api/etc/", "admin-api/etc/")) and endpoint in text:
                break
        else:
            if service.source.startswith(("api/etc/", "admin-api/etc/")):
                raise ValueError("missing direct RPC endpoints")
    elif etcd_key_suffix:
        for line in text.splitlines():
            stripped = line.strip()
            if stripped.startswith("Key:"):
                key = stripped.split(":", 1)[1].strip().strip('"').strip("'")
                if not key.endswith(etcd_key_suffix):
                    raise ValueError(f"missing suffix on Etcd key: {key}")


def render_config(
    backend_root: Path,
    service: ServiceConfig,
    output_root: Path,
    dsn: str,
    redis_pass: str,
    etcd_key_suffix: str,
    rpc_discovery_mode: str,
) -> Path:
    source = backend_root / service.source
    target = output_root / service.output
    text = source.read_text(encoding="utf-8")
    rendered = rewrite_db_block(text, dsn)
    rendered = rewrite_redis_pass(rendered, redis_pass)
    if rpc_discovery_mode == "direct":
        if service.source.startswith("rpc/"):
            rendered = rewrite_rpc_server_to_direct(rendered)
        else:
            rendered = rewrite_rpc_clients_to_direct(rendered)
    else:
        rendered = rewrite_etcd_keys(rendered, etcd_key_suffix)
    validate_rendered(rendered, service, dsn, redis_pass, etcd_key_suffix, rpc_discovery_mode)
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(rendered, encoding="utf-8")
    return target


def main() -> int:
    args = parse_args()
    dsn = args.dsn
    if not dsn:
        dsn = load_values(("REMOTE_JOURNAL_DSN",), Path(args.env_file))["REMOTE_JOURNAL_DSN"]
    if not dsn or "<" in dsn or ">" in dsn:
        raise SystemExit("missing --dsn and REMOTE_JOURNAL_DSN")

    backend_root = Path(__file__).resolve().parents[1]
    output_root = Path(args.output_dir)
    purge_generated_outputs(output_root)
    rendered_paths = [
        render_config(
            backend_root,
            service,
            output_root,
            dsn,
            args.redis_pass,
            args.etcd_key_suffix,
            args.rpc_discovery_mode,
        )
        for service in SERVICE_CONFIGS
    ]

    for path in rendered_paths:
        print(path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
