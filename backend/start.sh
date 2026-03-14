#!/usr/bin/env bash

set -euo pipefail

MODE="${1:-dev}"
PROFILE="${2:-local}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${ROOT_DIR}/logs"
RUN_DIR="${ROOT_DIR}/run"
REMOTE_CONFIG_DIR="${REMOTE_CONFIG_DIR:-/tmp/journal-remote-validation}"

mkdir -p "${LOG_DIR}" "${RUN_DIR}" "${ROOT_DIR}/bin"

SUPERVISOR_PIDS=()

usage() {
  cat <<'EOF'
Usage:
  ./start.sh dev [local|remote]
  ./start.sh prod [local|remote]

Environment variables:
  SKIP_INFRA=1       Skip docker-compose infrastructure startup
  SKIP_REPL_INIT=1   Skip MySQL replication initialization
  NO_RESTART=1       Disable automatic restart when a service exits unexpectedly
  DRY_RUN=1          Print the remote or local startup plan without launching processes
  REMOTE_JOURNAL_DSN Required for the remote profile
  REMOTE_CONFIG_DIR  Output dir for rendered remote YAMLs (default: /tmp/journal-remote-validation)
EOF
}

if [[ "${MODE}" != "dev" && "${MODE}" != "prod" ]]; then
  usage
  exit 1
fi

if [[ "${PROFILE}" != "local" && "${PROFILE}" != "remote" ]]; then
  usage
  exit 1
fi

compose_cmd() {
  if docker compose version >/dev/null 2>&1; then
    echo "docker compose"
    return
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
    return
  fi
  echo "docker compose command not found" >&2
  exit 1
}

wait_for_port() {
  local host="$1"
  local port="$2"
  local label="$3"
  local timeout="${4:-60}"
  local start_ts
  start_ts="$(date +%s)"

  if [[ "${DRY_RUN:-0}" == "1" ]]; then
    echo "[dry-run] wait for ${label} on ${host}:${port} within ${timeout}s"
    return
  fi

  while ! nc -z "${host}" "${port}" >/dev/null 2>&1; do
    if (( $(date +%s) - start_ts >= timeout )); then
      echo "[health] ${label} did not become ready on ${host}:${port} within ${timeout}s" >&2
      exit 1
    fi
    sleep 1
  done

  echo "[health] ${label} is ready on ${host}:${port}"
}

cleanup() {
  local pid
  if [[ ${#SUPERVISOR_PIDS[@]} -eq 0 ]]; then
    return
  fi
  for pid in "${SUPERVISOR_PIDS[@]}"; do
    kill "${pid}" >/dev/null 2>&1 || true
  done
  wait >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

compose_up() {
  local compose="$1"
  shift
  echo "[infra] starting docker-compose services: $*"
  if [[ "${DRY_RUN:-0}" == "1" ]]; then
    return
  fi
  (
    cd "${ROOT_DIR}"
    ${compose} up -d "$@"
  )
}

render_remote_configs() {
  if [[ -z "${REMOTE_JOURNAL_DSN:-}" ]]; then
    echo "[remote] REMOTE_JOURNAL_DSN is required for profile=remote" >&2
    exit 1
  fi

  echo "[remote] rendering configs into ${REMOTE_CONFIG_DIR}"
  if [[ "${DRY_RUN:-0}" == "1" ]]; then
    return
  fi

  (
    cd "${ROOT_DIR}"
    python3 scripts/render_remote_journal_configs.py \
      --dsn "${REMOTE_JOURNAL_DSN}" \
      --output-dir "${REMOTE_CONFIG_DIR}"
  )
}

start_local_infra() {
  local compose
  compose="$(compose_cmd)"

  compose_up "${compose}" mysql-master mysql-replica1 mysql-replica2 redis etcd jaeger prometheus grafana

  wait_for_port 127.0.0.1 13306 "mysql-master" 120
  wait_for_port 127.0.0.1 13307 "mysql-replica1" 120
  wait_for_port 127.0.0.1 13308 "mysql-replica2" 120
  wait_for_port 127.0.0.1 16379 "redis" 60
  wait_for_port 127.0.0.1 12379 "etcd" 60
  wait_for_port 127.0.0.1 4318 "jaeger-otlp" 60

  if [[ "${SKIP_REPL_INIT:-0}" != "1" ]]; then
    echo "[infra] initializing mysql replication"
    if [[ "${DRY_RUN:-0}" != "1" ]]; then
      (
        cd "${ROOT_DIR}"
        bash deploy/mysql/init-replication.sh
      )
    fi
  else
    echo "[infra] skipping replication initialization"
  fi
}

start_remote_infra() {
  local compose
  compose="$(compose_cmd)"

  compose_up "${compose}" redis etcd jaeger

  wait_for_port 127.0.0.1 16379 "redis" 60
  wait_for_port 127.0.0.1 12379 "etcd" 60
  wait_for_port 127.0.0.1 4318 "jaeger-otlp" 60
}

service_port() {
  case "$1" in
    user-rpc) echo "9001" ;;
    paper-rpc) echo "9002" ;;
    rating-rpc) echo "9003" ;;
    news-rpc) echo "9004" ;;
    admin-rpc) echo "9005" ;;
    api) echo "8888" ;;
    admin-api) echo "8889" ;;
    cron) echo "" ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

service_config_path() {
  case "$1" in
    user-rpc)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/rpc/user/etc/user.remote.yaml"
      else
        echo "rpc/user/etc/user.yaml"
      fi
      ;;
    paper-rpc)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/rpc/paper/etc/paper.remote.yaml"
      else
        echo "rpc/paper/etc/paper.yaml"
      fi
      ;;
    rating-rpc)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/rpc/rating/etc/rating.remote.yaml"
      else
        echo "rpc/rating/etc/rating.yaml"
      fi
      ;;
    news-rpc)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/rpc/news/etc/news.remote.yaml"
      else
        echo "rpc/news/etc/news.yaml"
      fi
      ;;
    admin-rpc)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/rpc/admin/etc/admin.remote.yaml"
      else
        echo "rpc/admin/etc/admin.yaml"
      fi
      ;;
    api)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/api/etc/journal-api.remote.yaml"
      else
        echo "api/etc/journal-api.yaml"
      fi
      ;;
    admin-api)
      if [[ "${PROFILE}" == "remote" ]]; then
        echo "${REMOTE_CONFIG_DIR}/admin-api/etc/admin-api.remote.yaml"
      else
        echo "admin-api/etc/admin-api.yaml"
      fi
      ;;
    cron)
      echo ""
      ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

service_command() {
  local service="$1"
  local config_file
  config_file="$(service_config_path "${service}")"
  case "${MODE}:${service}" in
    dev:user-rpc) echo "go run rpc/user/user.go -f ${config_file}" ;;
    dev:paper-rpc) echo "go run rpc/paper/paper.go -f ${config_file}" ;;
    dev:rating-rpc) echo "go run rpc/rating/rating.go -f ${config_file}" ;;
    dev:news-rpc) echo "go run rpc/news/news.go -f ${config_file}" ;;
    dev:admin-rpc) echo "go run rpc/admin/admin.go -f ${config_file}" ;;
    dev:api) echo "go run api/journal.go -f ${config_file}" ;;
    dev:admin-api) echo "go run admin-api/admin.go -f ${config_file}" ;;
    dev:cron) echo "go run cmd/cron/main.go" ;;

    prod:user-rpc) echo "./bin/user-rpc -f ${config_file}" ;;
    prod:paper-rpc) echo "./bin/paper-rpc -f ${config_file}" ;;
    prod:rating-rpc) echo "./bin/rating-rpc -f ${config_file}" ;;
    prod:news-rpc) echo "./bin/news-rpc -f ${config_file}" ;;
    prod:admin-rpc) echo "./bin/admin-rpc -f ${config_file}" ;;
    prod:api) echo "./bin/api -f ${config_file}" ;;
    prod:admin-api) echo "./bin/admin-api -f ${config_file}" ;;
    prod:cron) echo "./bin/cron" ;;
    *)
      echo "unsupported mode/service pair: ${MODE}:${service}" >&2
      exit 1
      ;;
  esac
}

start_service() {
  local service="$1"
  local command="$2"
  local port="$3"
  local logfile="${LOG_DIR}/${service}.log"
  local pidfile="${RUN_DIR}/${service}.pid"

  echo "[start] ${service}"
  if [[ "${DRY_RUN:-0}" == "1" ]]; then
    echo "[dry-run] ${service}: ${command}"
    if [[ -n "${port}" ]]; then
      echo "[dry-run] expected port ${port}"
    fi
    return
  fi

  (
    cd "${ROOT_DIR}"
    while true; do
      exit_code=0
      echo "[$(date '+%F %T')] starting ${service}: ${command}" >> "${logfile}"
      set +e
      bash -lc "${command}" >> "${logfile}" 2>&1
      exit_code=$?
      set -e
      echo "[$(date '+%F %T')] ${service} exited with code ${exit_code}" >> "${logfile}"
      if [[ "${NO_RESTART:-0}" == "1" ]]; then
        exit "${exit_code}"
      fi
      sleep 2
    done
  ) &

  local supervisor_pid=$!
  SUPERVISOR_PIDS+=("${supervisor_pid}")
  echo "${supervisor_pid}" > "${pidfile}"

  if [[ -n "${port}" ]]; then
    wait_for_port 127.0.0.1 "${port}" "${service}" 120
  else
    sleep 2
    echo "[health] ${service} supervisor pid=${supervisor_pid}"
  fi
}

build_prod_binaries() {
  echo "[build] compiling production binaries"
  (
    cd "${ROOT_DIR}"
    make build
  )
}

print_summary() {
  local services=("$@")
  local infra_lines
  if [[ "${PROFILE}" == "remote" ]]; then
    infra_lines=$'  - redis:          16379\n  - etcd:           12379\n  - jaeger ui:      16686'
  else
    infra_lines=$'  - mysql-master:   13306\n  - mysql-replica1: 13307\n  - mysql-replica2: 13308\n  - redis:          16379\n  - etcd:           12379\n  - jaeger ui:      16686\n  - prometheus:     19090\n  - grafana:        13000'
  fi
  cat <<EOF

Service supervisor is up.
Profile:
  - mode:          ${MODE}
  - config profile:${PROFILE}

Public APIs:
  - user api:    http://127.0.0.1:8888
  - admin api:   http://127.0.0.1:8889

RPC ports:
  - user-rpc:    9001
  - paper-rpc:   9002
  - rating-rpc:  9003
  - news-rpc:    9004
  - admin-rpc:   9005

Infra:
${infra_lines}

Logs:
  ${LOG_DIR}

Services:
$(printf '  - %s\n' "${services[@]}")

Press Ctrl-C to stop the supervisor.
EOF
}

if [[ "${PROFILE}" == "remote" ]]; then
  render_remote_configs
fi

if [[ "${SKIP_INFRA:-0}" != "1" ]]; then
  if [[ "${PROFILE}" == "remote" ]]; then
    start_remote_infra
  else
    start_local_infra
  fi
else
  echo "[infra] skipped by SKIP_INFRA=1"
fi

if [[ "${MODE}" == "prod" ]]; then
  build_prod_binaries
fi

if [[ "${PROFILE}" == "remote" ]]; then
  SERVICES=(
    "user-rpc"
    "paper-rpc"
    "rating-rpc"
    "news-rpc"
    "admin-rpc"
    "api"
    "admin-api"
  )
else
  SERVICES=(
    "user-rpc"
    "paper-rpc"
    "rating-rpc"
    "news-rpc"
    "admin-rpc"
    "api"
    "admin-api"
    "cron"
  )
fi

for service in "${SERVICES[@]}"; do
  start_service "${service}" "$(service_command "${service}")" "$(service_port "${service}")"
done

print_summary "${SERVICES[@]}"
wait
