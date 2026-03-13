# Single-DB Merge Runbook

This runbook closes `JDB-040` by defining a reversible migration path from `journal_biz` and `journal_admin` into a single `journal` schema with `biz_*` and `adm_*` tables.

## Default Strategy

1. Bootstrap target prefixed tables inside `journal`.
2. Backfill data by copy instead of `RENAME TABLE`, so the old schemas stay intact.
3. Create compatibility views for the legacy business table names inside `journal`, then freeze application writes, run a final delta copy, verify row counts and primary-key ranges, and cut application traffic to the new tables.
4. Keep `journal_biz` and `journal_admin` read-only for the rollback window. If cutover fails, disable read/write split first and switch DSNs back to the old schemas without doing a full data reverse-migration.

## Table Mapping

| Source schema.table | Target schema.table |
| --- | --- |
| `journal_biz.user` | `journal.biz_user` |
| `journal_biz.user_achievement` | `journal.biz_user_achievement` |
| `journal_biz.paper` | `journal.biz_paper` |
| `journal_biz.cold_paper` | `journal.biz_cold_paper` |
| `journal_biz.rating` | `journal.biz_rating` |
| `journal_biz.news` | `journal.biz_news` |
| `journal_biz.flag` | `journal.biz_flag` |
| `journal_biz.keyword_rule` | `journal.biz_keyword_rule` |
| `journal_admin.adm_role` | `journal.adm_role` |
| `journal_admin.adm_permission` | `journal.adm_permission` |
| `journal_admin.adm_role_permission` | `journal.adm_role_permission` |
| `journal_admin.adm_user_role` | `journal.adm_user_role` |
| `journal_admin.adm_audit_log` | `journal.adm_audit_log` |

Admin identity remains on `journal.biz_user`; `adm_user_role.user_id` continues
to point at that table after the merge, so there is no standalone `adm_user`
target in the single-db baseline.

## Compatibility Views

- Keep `journal.user`, `journal.user_achievement`, `journal.paper`,
  `journal.cold_paper`, `journal.rating`, `journal.news`, `journal.flag`, and
  `journal.keyword_rule` as view-based aliases to the `biz_*` tables during the
  cutover window.
- Treat those views as a temporary compatibility layer. New DDL and new code
  should target the prefixed tables directly.

## Dry-Run And Rehearsal

- Render the full SQL workbook: `make single-db-merge-dry-run`
- Render only a single phase when you want to review or execute it separately:
  - `python3 scripts/rehearse_single_db_merge.py --phase bootstrap`
  - `python3 scripts/rehearse_single_db_merge.py --phase backfill`
  - `python3 scripts/rehearse_single_db_merge.py --phase verify`
  - `python3 scripts/rehearse_single_db_merge.py --phase rollback`
- Execute the reviewed SQL in rehearsal or production windows with `mysql`:
  - `python3 scripts/rehearse_single_db_merge.py --phase all > /tmp/single-db-merge.sql`
  - `mysql --default-character-set=utf8mb4 < /tmp/single-db-merge.sql`

## Go/No-Go Checks

- Every mapped table must pass the generated `row_count/min_id/max_id` reconciliation query before application cutover.
- Compatibility views for the legacy business table names must resolve against the copied `biz_*` tables before application cutover.
- Old schemas must remain untouched after bootstrap and backfill; any rollback should be a DSN/config switch, not a reverse `RENAME`.
- If the final delta copy cannot finish inside the write freeze window, stop the cutover and keep serving from the old schemas.
- If zero-downtime is mandatory, enable temporary dual-write only for the final delta phase. It is a fallback, not the default migration path.

## Rollback Window

- Keep `journal_biz` and `journal_admin` available as the rollback source of truth until post-cutover validation and monitoring checks pass.
- On failure, execute rollback in this order:
  1. Disable read/write split and force all reads back to the primary.
  2. Restore application DSNs to `journal_biz` and `journal_admin`.
  3. Remove or bypass the compatibility views if they obscure drift during rollback verification.
  4. Re-run the verification queries to document divergence between old and new tables.
  5. Leave `journal.*` prefixed tables in place for diffing and a later retry.

## Validation Notes

- The provided rehearsal script is intentionally dry-run first. It renders repeatable SQL for bootstrap, backfill, verification, and rollback review without requiring a live database connection from the repo.
- A full rehearsal still needs a pre-production or equivalent MySQL environment, because this repository does not carry a disposable dataset large enough to prove row-count reconciliation or rollback timing locally.
