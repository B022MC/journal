# Remote Journal Startup Path

Date: 2026-03-14

Purpose: close `JRV-030` by defining a reusable startup command for remote
single-db validation that does not start the local MySQL master or replicas.

## Commands

- Default local mode is unchanged: `./start.sh dev`
- Remote validation dry-run:
  `DRY_RUN=1 REMOTE_JOURNAL_DSN='<redacted>' ./start.sh dev remote`
- Remote validation live run:
  `REMOTE_JOURNAL_DSN='<redacted>' ./start.sh dev remote`
- Remote validation via the prepared local env file:
  `backend/scripts/with_remote_journal_validation_env.sh ./start.sh dev remote`

## Remote Profile Behavior

- Renders the remote YAMLs from `JRV-020` into
  `${REMOTE_CONFIG_DIR:-/tmp/journal-remote-validation}`.
- Rewrites `journal-api` and `admin-api` to use direct localhost RPC endpoints
  (`127.0.0.1:9001-9005`) so remote validation does not depend on live etcd
  registration.
- Removes the RPC `Etcd` blocks from the rendered remote YAMLs, which avoids
  publishing validation-only services into the live service-discovery namespace.
- Starts only `redis`, `etcd`, and `jaeger` from `docker-compose.yaml`.
- Does not start `mysql-master`, `mysql-replica1`, or `mysql-replica2`.
- Does not run `deploy/mysql/init-replication.sh`.
- Starts services in order:
  `user.rpc -> paper.rpc -> rating.rpc -> news.rpc -> admin.rpc -> journal-api -> admin-api`
- Leaves the default local profile unchanged, including MySQL replication init
  and the existing `cron` process.

## Local Env Notes

- `start.sh` now launches each service with `bash -c` instead of `bash -lc`,
  so any injected Go toolchain overrides (`GOROOT`, `PATH`, `GOCACHE`) are
  preserved during remote validation.
- When local Docker infra is unavailable, `SKIP_INFRA=1` plus the prepared
  remote Redis and Jaeger tunnels is sufficient for the direct-RPC smoke path.

## Failure Signals

- Missing `REMOTE_JOURNAL_DSN` fails fast before any remote profile startup.
- The wrapper also fails fast if `backend/.env.remote-validation.local` is
  missing, so the validation session does not depend on ad-hoc shell history.
- `DRY_RUN=1` prints the exact infra and service commands without launching
  containers or Go processes.
- `SKIP_INFRA=1` still works for both profiles when infra is already running.

## Rollback

- To return to the default local stack, stop the current supervisor and run
  `./start.sh dev`.
- Remote rendered YAMLs stay outside the repository by default, so no git state
  rollback is required for the config files themselves.
