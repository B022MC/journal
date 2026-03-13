# Journal DB Merge Legacy Inventory

This artifact closes `JDB-010` by freezing the pre-merge legacy naming surface before the repo starts moving from dual schemas to a single `journal` schema.

## Freeze Rules

1. No new runtime or deployment config may introduce `journal_biz`, `journal_admin`, `BizDB`, or `AdminDB`.
2. No new DDL or SQL may introduce the bare business tables `user`, `user_achievement`, `paper`, `cold_paper`, `rating`, `news`, `flag`, or `keyword_rule`.
3. During the migration window, only paths already recorded in [`2026-03-13-db-merge-legacy-baseline.csv`](./2026-03-13-db-merge-legacy-baseline.csv) may retain legacy references.
4. Every new or increased legacy reference must fail `make db-naming-freeze-check` before merge.

## Coverage Snapshot

The audited baseline contains 122 `(path, kind, token)` tuples and 383 legacy hits across the required surfaces:

- API: 10 hits across `api/etc/journal-api.yaml`, `api/internal/config/config.go`, and `api/internal/svc/serviceContext.go`.
- admin-api: 10 hits across `admin-api/etc/admin-api.yaml`, `admin-api/internal/config/config.go`, and `admin-api/internal/svc/serviceContext.go`.
- admin-rpc: 14 hits across `rpc/admin/etc/admin.yaml`, `rpc/admin/internal/config/config.go`, and `rpc/admin/internal/svc/serviceContext.go`.
- paper/user/news/rating RPC: 29 hits across each service's `etc/*.yaml`, `internal/config/*.go`, `internal/svc/*.go`, plus `rpc/rating/internal/eventing/postrate.go`.
- cron and offline entrypoints: `cmd/cron/main.go` and `cmd/lifecycle/main.go` still hardcode `journal_biz`.
- docker-compose and k8s manifests: 43 hits across `docker-compose.yaml`, `deploy/k8s/base/secrets.yaml`, `deploy/k8s/middleware/init/mysql-init.yaml`, and the backend service manifests.
- schema baseline and migrations: 106 hits across `model/schema.sql` and `model/migrations/*.sql`.
- shared model SQL layer: 169 hits across the shared DAO files in `model/*.go`, which is where bare table names are currently concentrated.

## Audited Artifacts

- Machine-readable allowlist and counts: [`2026-03-13-db-merge-legacy-baseline.csv`](./2026-03-13-db-merge-legacy-baseline.csv)
- Freeze-window guard: [`../../scripts/check_legacy_db_refs.py`](../../scripts/check_legacy_db_refs.py)
- Make entrypoint: `make db-naming-freeze-check`

## Operating Notes

- The guard only scans runtime, migration, deployment, and shared model paths. It intentionally ignores docs, generated protobuf stubs, tests, and `api/journal.json`.
- The baseline is count-based. Deleting or reducing legacy references is allowed; introducing a new file/token pair or increasing a recorded count is blocked.
- When a later migration issue legitimately removes or rewrites legacy references, update the baseline in the same commit as the code change so the freeze remains auditable.
