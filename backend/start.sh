#!/usr/bin/env bash

set -euo pipefail

MODE="${1:-dev}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${ROOT_DIR}/logs"
RUN_DIR="${ROOT_DIR}/run"

mkdir -p "${LOG_DIR}" "${RUN_DIR}" "${ROOT_DIR}/bin"

SUPERVISOR_PIDS=()

usage() {
  cat <<'EOF'
Usage:
  ./start.sh dev
  ./start.sh prod

Environment variables:
  SKIP_INFRA=1       Skip docker-compose infrastructure startup
  SKIP_REPL_INIT=1   Skip MySQL replication initialization
  NO_RESTART=1       Disable automatic restart when a service exits unexpectedly
EOF
}

if [[ "${MODE}" != "dev" && "${MODE}" != "prod" ]]; then
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
  for pid in "${SUPERVISOR_PIDS[@]}"; do
    kill "${pid}" >/dev/null 2>&1 || true
  done
  wait >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

start_infra() {
  local compose
  compose="$(compose_cmd)"

  echo "[infra] starting docker-compose services"
  (
    cd "${ROOT_DIR}"
    ${compose} up -d
  )

  wait_for_port 127.0.0.1 13306 "mysql-master" 120
  wait_for_port 127.0.0.1 13307 "mysql-replica1" 120
  wait_for_port 127.0.0.1 13308 "mysql-replica2" 120
  wait_for_port 127.0.0.1 16379 "redis" 60
  wait_for_port 127.0.0.1 12379 "etcd" 60
  wait_for_port 127.0.0.1 4318 "jaeger-otlp" 60

  if [[ "${SKIP_REPL_INIT:-0}" != "1" ]]; then
    echo "[infra] initializing mysql replication"
    (
      cd "${ROOT_DIR}"
      bash deploy/mysql/init-replication.sh
    )
  else
    echo "[infra] skipping replication initialization"
  fi
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

service_command() {
  local service="$1"
  case "${MODE}:${service}" in
    dev:user-rpc) echo "go run rpc/user/user.go -f rpc/user/etc/user.yaml" ;;
    dev:paper-rpc) echo "go run rpc/paper/paper.go -f rpc/paper/etc/paper.yaml" ;;
    dev:rating-rpc) echo "go run rpc/rating/rating.go -f rpc/rating/etc/rating.yaml" ;;
    dev:news-rpc) echo "go run rpc/news/news.go -f rpc/news/etc/news.yaml" ;;
    dev:admin-rpc) echo "go run rpc/admin/admin.go -f rpc/admin/etc/admin.yaml" ;;
    dev:api) echo "go run api/journal.go -f api/etc/journal-api.yaml" ;;
    dev:admin-api) echo "go run admin-api/admin.go -f admin-api/etc/admin-api.yaml" ;;
    dev:cron) echo "go run cmd/cron/main.go" ;;

    prod:user-rpc) echo "./bin/user-rpc -f rpc/user/etc/user.yaml" ;;
    prod:paper-rpc) echo "./bin/paper-rpc -f rpc/paper/etc/paper.yaml" ;;
    prod:rating-rpc) echo "./bin/rating-rpc -f rpc/rating/etc/rating.yaml" ;;
    prod:news-rpc) echo "./bin/news-rpc -f rpc/news/etc/news.yaml" ;;
    prod:admin-rpc) echo "./bin/admin-rpc -f rpc/admin/etc/admin.yaml" ;;
    prod:api) echo "./bin/api -f api/etc/journal-api.yaml" ;;
    prod:admin-api) echo "./bin/admin-api -f admin-api/etc/admin-api.yaml" ;;
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
  cat <<EOF

All services are up.

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
  - mysql-master:   13306
  - mysql-replica1: 13307
  - mysql-replica2: 13308
  - redis:          16379
  - etcd:           12379
  - jaeger ui:      16686
  - prometheus:     19090
  - grafana:        13000

Logs:
  ${LOG_DIR}

Press Ctrl-C to stop the supervisor.
EOF
}

if [[ "${SKIP_INFRA:-0}" != "1" ]]; then
  start_infra
else
  echo "[infra] skipped by SKIP_INFRA=1"
fi

if [[ "${MODE}" == "prod" ]]; then
  build_prod_binaries
fi

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

for service in "${SERVICES[@]}"; do
  start_service "${service}" "$(service_command "${service}")" "$(service_port "${service}")"
done

print_summary
wait
