# DB Merge Milestone Runbook

Date: 2026-03-13

Purpose: keep the journal DB merge milestones, owners, go or no-go gates, and
rollback steps in one place so the cutover review can reject incomplete stages
before they reach production.

## Global Rollback Rule

Every rollback starts in the same order:

1. Disable replica reads and force all traffic back to the primary.
2. Restore the previous DSN or schema mapping before changing any data-copy
   state.
3. Re-run the verification queries from the single-db merge workbook.
4. Preserve the new target tables, metrics, and logs for diffing instead of
   trying a full reverse data migration.

## Owner Matrix

| Role | Owned scope | Handoff trigger | Handoff artifact |
| --- | --- | --- | --- |
| DB maintainer | schema baseline, migration workbook, reconciliation queries, and rollback SQL review | bootstrap or copy plan is ready for rehearsal | updated runbook, mapping diff, and blocker list |
| Service maintainer | config wiring, DSN switches, explicit primary-read paths, and service restart plan | config diff and restart steps are reviewable | service-specific config diff, restart order, and fallback toggles |
| SRE or release captain | replica health, metrics, alerting, and cutover window approval | DB and service evidence are both green for the milestone | signed go or no-go note, rollback drill record, and monitoring snapshot |

## Milestone Gates

| Milestone | Owner | Go or no-go gate | Required evidence | Rollback action | Current blocker |
| --- | --- | --- | --- | --- | --- |
| M1: naming and DDL baseline | DB maintainer | legacy naming freeze passes and new baseline artifacts are reviewed | `make db-naming-freeze-check`, legacy inventory, updated DDL or init diff | revert the baseline commit, rerun `make db-naming-freeze-check`, and keep old schemas untouched | `JDB-020` still needs the actual single-schema DDL baseline |
| M2: application entry unification | Service maintainer | service configs no longer require dual-schema wiring and strong-consistency reads remain explicit | config diff for API, admin-api, admin-rpc, paper or user or news or rating RPC; startup proof; primary-read audit | restore prior DSNs or config structs, keep `ReadWriteSplit: false`, restart services in reverse order | `JDB-050` is still open |
| M3: single-db cutover | DB maintainer plus release captain | full rehearsal copy succeeds, row_count or min_id or max_id match, and rollback workbook is ready | `make single-db-merge-dry-run`, applied rehearsal log, reconciliation output, rollback drill record | run `python3 scripts/rehearse_single_db_merge.py --phase rollback`, restore old DSNs, rerun verify queries | live rehearsal still blocked by external MySQL env in `JDB-040` and `JDB-080` |
| M4: read or write split rollout | Service maintainer plus SRE | only approved read-mostly flows hit replicas and lag or outage drills fall back safely | per-service rollout matrix, replica lag drill note, monitoring screenshot, service-level toggle record | set `ReadWriteSplit: false` in affected YAMLs, restart services, confirm strong-consistency paths stay on primary | `JDB-060` still needs rollout ownership and live drills |
| M5: hotspot sharding PoC | DB maintainer plus feature owner | PoC stays inside the approved whitelist and does not widen runtime blast radius | sharding proposal, routing threshold, benchmark result, and rollback switch | disable the PoC flag or router, keep logical tables on the single-db path, and stop the experiment | `JDB-030` and `JDB-070` are still open |

## Rollback Entry Points

- Replica-read rollback:
  - Set `ReadWriteSplit: false` in the affected service YAMLs.
  - Restart the service and verify primary-read paths still use `sqlx.WithReadPrimary` for login, permission, and write-after-read flows.
- Schema rollback:
  - Restore DSNs from `journal` or prefixed tables back to `journal_biz` and `journal_admin`.
  - Run the verify section from `python3 scripts/rehearse_single_db_merge.py --phase verify`.
- Milestone review rollback:
  - If any required evidence artifact is missing, do not advance the milestone. The current owner keeps the stage until the artifact exists in the repo or release packet.

## Current Go or No-Go Rule

- No milestone may advance unless its blocker column is empty or explicitly signed off by the release captain with updated evidence.
- No one may skip straight to M4 or M5 while M2 or M3 is still blocked.
- Any incident during M3 or M4 immediately falls back to the global rollback rule above before deeper investigation begins.
