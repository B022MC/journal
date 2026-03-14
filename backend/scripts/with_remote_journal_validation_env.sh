#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${REMOTE_VALIDATION_ENV_FILE:-${ROOT_DIR}/.env.remote-validation.local}"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[remote-validation] env file not found: ${ENV_FILE}" >&2
  exit 1
fi

set -a
source "${ENV_FILE}"
set +a

exec "$@"
