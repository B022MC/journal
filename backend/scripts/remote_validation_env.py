from __future__ import annotations

import os
import shlex
from pathlib import Path


DEFAULT_ENV_FILE = Path(__file__).resolve().parents[1] / ".env.remote-validation.local"


def parse_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values

    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if not key:
            continue
        if value and value[0] in {"'", '"'}:
            value = shlex.split(value)[0]
        values[key] = value
    return values


def load_values(keys: tuple[str, ...], env_file: Path | None = None) -> dict[str, str]:
    values = {key: os.environ.get(key, "").strip() for key in keys}
    path = env_file or DEFAULT_ENV_FILE
    file_values = parse_env_file(path)
    for key in keys:
        if not values[key]:
            values[key] = file_values.get(key, "").strip()
    return values
