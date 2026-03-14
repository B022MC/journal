#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path


DEFAULT_DSN = (
    "journal:redacted@tcp(remote-host:3306)/journal"
    "?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run and validate the remote/local rollback dry-run drill."
    )
    parser.add_argument(
        "--dsn",
        default=DEFAULT_DSN,
        help="Remote DSN placeholder used for the remote dry-run.",
    )
    return parser.parse_args()


def backend_root() -> Path:
    return Path(__file__).resolve().parents[1]


def run_start(profile: str, dsn: str) -> str:
    env = os.environ.copy()
    env["DRY_RUN"] = "1"
    if profile == "remote":
        env["REMOTE_JOURNAL_DSN"] = dsn
    else:
        env.pop("REMOTE_JOURNAL_DSN", None)

    result = subprocess.run(
        ["./start.sh", "dev", profile],
        cwd=backend_root(),
        env=env,
        capture_output=True,
        text=True,
        check=False,
    )
    output = result.stdout + result.stderr
    if result.returncode != 0:
        raise RuntimeError(f"{profile} dry-run failed with exit code {result.returncode}\n{output}")
    return output


def collect_marker_errors(output: str, *, required: tuple[str, ...], forbidden: tuple[str, ...]) -> list[str]:
    errors: list[str] = []
    for marker in required:
        if marker not in output:
            errors.append(f"missing:{marker}")
    for marker in forbidden:
        if marker in output:
            errors.append(f"forbidden:{marker}")
    return errors


def validate_remote_output(output: str) -> list[str]:
    return collect_marker_errors(
        output,
        required=(
            "[remote] rendering configs into /tmp/journal-remote-validation",
            "[infra] starting docker-compose services: redis etcd jaeger",
            "/tmp/journal-remote-validation/rpc/user/etc/user.remote.yaml",
            "/tmp/journal-remote-validation/admin-api/etc/admin-api.remote.yaml",
            "config profile:remote",
            "  - redis:          16379",
            "  - jaeger ui:      16686",
        ),
        forbidden=(
            "mysql-master",
            "mysql-replica1",
            "mysql-replica2",
            "prometheus",
            "grafana",
            "cmd/cron/main.go",
            "config profile:local",
        ),
    )


def validate_local_output(output: str) -> list[str]:
    return collect_marker_errors(
        output,
        required=(
            "[infra] starting docker-compose services: mysql-master mysql-replica1 mysql-replica2 redis etcd jaeger prometheus grafana",
            "go run rpc/user/user.go -f rpc/user/etc/user.yaml",
            "go run admin-api/admin.go -f admin-api/etc/admin-api.yaml",
            "go run cmd/cron/main.go",
            "config profile:local",
            "  - mysql-master:   13306",
            "  - grafana:        13000",
        ),
        forbidden=(
            "/tmp/journal-remote-validation",
            "config profile:remote",
        ),
    )


def main() -> int:
    args = parse_args()
    remote_output = run_start("remote", args.dsn)
    remote_errors = validate_remote_output(remote_output)
    local_output = run_start("local", args.dsn)
    local_errors = validate_local_output(local_output)

    if remote_errors or local_errors:
        print("Remote rollback dry-run failed:")
        for error in remote_errors:
            print(f"- remote:{error}")
        for error in local_errors:
            print(f"- local:{error}")
        return 1

    print("Remote rollback dry-run passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
