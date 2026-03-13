# DB Merge Validation Matrix

Date: 2026-03-13

Purpose: turn the single-db merge, read/write split rollout, and rollback plan
into a fixed go/no-go matrix so every rehearsal produces the same evidence
package instead of ad-hoc notes.

## Summary

Current repository evidence is enough to lock the validation entrypoints, but
not enough to claim production readiness. The dry-run migration workbook and
legacy naming freeze guard are in place; a disposable pre-production MySQL
environment is still required for live copy, replica delay injection, explain
capture, and rollback timing evidence.

## Matrix

| Gate | Environment | Entry point | Pass criteria | Current status | Evidence or blocker |
| --- | --- | --- | --- | --- | --- |
| Schema diff and naming freeze | Local CI or dev shell | `cd backend && make db-naming-freeze-check` | No new legacy schema or bare-table reference escapes the audited baseline | Ready now | `backend/docs/release/2026-03-13-db-merge-legacy-inventory.md` and `backend/docs/release/2026-03-13-db-merge-legacy-baseline.csv` |
| Target-schema bootstrap dry-run | Local CI or dev shell | `cd backend && make single-db-merge-dry-run` | Workbook renders `journal` bootstrap, backfill, compatibility-view, verify, and rollback sections without direct `RENAME TABLE` | Ready now | `backend/docs/release/2026-03-13-single-db-merge-runbook.md` and `backend/scripts/rehearse_single_db_merge.py` |
| Full migration rehearsal | Pre-production MySQL or equivalent sample dataset | `python3 scripts/rehearse_single_db_merge.py --phase all > /tmp/single-db-merge.sql` then `mysql < /tmp/single-db-merge.sql` | All mapped tables complete copy, row_count/min_id/max_id match, old schemas stay intact for rollback | Blocked on env | Needs disposable MySQL with `journal_biz`, `journal_admin`, and target `journal` privileges |
| Read/write split regression | Pre-production services with replicas | Toggle `ReadWriteSplit` in `api/etc/journal-api.yaml`, `rpc/rating/etc/rating.yaml`, `rpc/news/etc/news.yaml`, `rpc/paper/etc/paper.yaml`, `rpc/admin/etc/admin.yaml` | Read-mostly flows stay on replicas, login/permission/write-after-read flows remain on primary or explicit `sqlx.WithReadPrimary` paths | Partially ready | `backend/common/dbconfig/dbconfig.go` and existing `sqlx.WithReadPrimary` usage define the code anchors; still needs live service rehearsal |
| Replica delay and replica-down drill | Pre-production MySQL replication topology | Inject lag or stop one replica, then hit paper list, news list, rating aggregate, login, permission checks | Weak-consistency flows degrade back to primary or tolerate lag; strong-consistency flows stay unaffected | Blocked on env | Needs replication topology plus metrics split by primary or replica |
| Core query explain capture | Pre-production MySQL with representative data | Run `EXPLAIN` on paper list, rating aggregate, RBAC permission lookup, and audit-log listing after bootstrap | Query plans use expected indexes and do not regress into full scans for high-volume paths | Blocked on env | Candidate anchors: `biz_paper`, `biz_rating`, `adm_permission`, `adm_audit_log` |
| Hot-write benchmark | Pre-production or isolated load-test env | Replay rating writes, flag writes, and admin audit-log writes at target QPS before enabling replicas broadly | P95 write latency and error rate stay inside agreed SLO; no replica-induced backlog blocks rollback | Blocked on env | Needs replay dataset or load generator; repo does not include one yet |
| Admin permission chain validation | Pre-production services | Login, `/me/permissions`, role-permission update, and audit-log verification against admin flows | Permission checks, super-admin bypass, `biz_user`-backed admin identity, and audit logging remain correct before and after cutover | Blocked on env | Code anchors in `backend/model/adminrbacmodel.go`, `backend/admin-api/internal/logic/auth/adminLoginLogic.go`, and `backend/admin-api/internal/logic/manage/adminauth.go` |
| Rollback drill | Same env as full rehearsal | Disable read/write split, restore old DSNs, re-run generated verify queries | Service recovers old-schema read/write path without full reverse migration and leaves target tables for diffing | Blocked on env | Runbook entrypoint is `backend/docs/release/2026-03-13-single-db-merge-runbook.md` |

## Required Evidence Package

- Command log for `make db-naming-freeze-check` and `make single-db-merge-dry-run`
- Applied SQL workbook or equivalent orchestration log from the rehearsal window
- Verification output for every mapped table: `row_count`, `min_id`, `max_id`
- Replica lag or failure drill notes with observed fallback behavior
- `EXPLAIN` output for paper list, rating aggregate, RBAC permission lookup, and audit log queries
- Hot-write benchmark summary with target QPS, latency, and error rate
- Rollback drill log showing the order: disable replica reads, restore old DSNs, rerun verify queries

## Current Blockers

- No disposable MySQL environment is wired into this repo for a real copy-and-verify rehearsal.
- No checked-in load generator exists yet for rating or flag hot-write replay.
- Metrics and dashboards for primary versus replica split still need to be collected in the target environment rather than inferred locally.
