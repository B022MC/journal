# Journal DB Merge Legacy Inventory

This artifact closes `JDB-010` by freezing the pre-merge legacy naming surface before the repo starts moving from dual schemas to a single `journal` schema.

## Freeze Rules

1. No new runtime or deployment config may introduce `journal_biz`, `journal_admin`, `BizDB`, or `AdminDB`.
2. No new DDL or SQL may introduce the bare business tables `user`, `user_achievement`, `paper`, `cold_paper`, `rating`, `news`, `flag`, or `keyword_rule`.
3. During the migration window, only paths already recorded in [`2026-03-13-db-merge-legacy-baseline.csv`](./2026-03-13-db-merge-legacy-baseline.csv) may retain legacy references.
4. Every new or increased legacy reference must fail `make db-naming-freeze-check` before merge.

## Coverage Snapshot

The current audited baseline contains 64 `(path, kind, token)` tuples and 309
legacy hits across the required surfaces:

- API, admin-api, admin-rpc, cron, lifecycle, docker-compose, and service
  manifests: 0 legacy hits after the `DB` and single-schema cutover.
- RPC layer: 1 remaining bare-table hit in `rpc/rating/internal/eventing/postrate.go`.
- k8s bootstrap init: 9 legacy-table hits in `deploy/k8s/middleware/init/mysql-init.yaml`,
  all from the intentional compatibility views for old business table names.
- schema baseline and migrations: 130 hits across `model/schema.sql` and
  `model/migrations/*.sql`; these are now concentrated in historical split-db
  migrations plus the explicit `009` merge workbook.
- shared model SQL layer: 169 hits across the shared DAO files in `model/*.go`,
  which is where the remaining bare table names still live while compatibility
  views carry the cutover.

## Audited Artifacts

- Machine-readable allowlist and counts: [`2026-03-13-db-merge-legacy-baseline.csv`](./2026-03-13-db-merge-legacy-baseline.csv)
- Freeze-window guard: [`../../scripts/check_legacy_db_refs.py`](../../scripts/check_legacy_db_refs.py)
- Make entrypoint: `make db-naming-freeze-check`

## Operating Notes

- The guard only scans runtime, migration, deployment, and shared model paths. It intentionally ignores docs, generated protobuf stubs, tests, and `api/journal.json`.
- The baseline is count-based. Deleting or reducing legacy references is allowed; introducing a new file/token pair or increasing a recorded count is blocked.
- When a later migration issue legitimately removes or rewrites legacy references, update the baseline in the same commit as the code change so the freeze remains auditable.
- Compatibility views and rollback workbooks are allowed to retain legacy table
  names only when they are explicitly part of the migration path.
