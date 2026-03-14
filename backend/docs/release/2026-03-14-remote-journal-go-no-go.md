# Remote Journal Go Or No-Go Template

Date: 2026-03-14

Purpose: close `JRV-070` by defining a fixed decision template and rollback
sequence for remote single-db validation. The team should use this template
instead of improvising during a failure window.

## Decision Rule

- Go only when `JRV-040`, `JRV-050`, and `JRV-060` all have green evidence.
- No-go immediately if any service log shows the wrong DSN target, a missing
  compatibility object, a read-only view write failure, or an admin permission
  mismatch.
- No-go if disposable test data cannot be traced and cleaned up.
- No-go means rollback first, diagnosis second. Do not start `biz_*`
  convergence during the incident window.

## Go Or No-Go Table

| Gate | Required owner | Go condition | No-go trigger | Evidence |
| --- | --- | --- | --- | --- |
| Service startup | `JRV-040` backend smoke owner | `user.rpc -> paper.rpc -> rating.rpc -> news.rpc -> admin.rpc -> journal-api -> admin-api` all start with the remote profile and no service touches `127.0.0.1:13306/13307/13308` | any startup log still points at local MySQL or cannot connect to remote `journal` | startup command log, effective config path, DSN proof, health check log |
| Frontend minimum regression | `JRV-050` frontend owner | register, login, `/user/info`, paper list or detail or search, rating, paper report, and rating report all succeed with disposable data | any compatibility view is missing, stale, not writable, or produces wrong-side effects | request or response log, related backend log, captured disposable ids |
| Admin minimum regression | `JRV-060` admin owner | admin login, my permissions, role CRUD, role-permission update, user-role assign or revoke, keyword-rule CRUD, flag resolve, and audit-log list all succeed | permission check mismatch, audit log missing, keyword rule write fails, or admin mutation hits wrong DB | request or response log, audit-log ids, cleanup record |
| Data cleanup | release owner | every disposable user, role, keyword rule, news row, rating, and flag created during validation is either deleted or explicitly recorded for cleanup | orphan disposable data remains without ids or cleanup owner | cleanup SQL log or command log |

## 15-Minute Rollback Sequence

1. Stop the remote-profile supervisor in the active shell with `Ctrl-C`.
2. If the remote profile was launched in the background, stop every service
   supervisor recorded by pidfiles:

   ```bash
   cd backend
   for pidfile in run/*.pid; do
     kill "$(cat "$pidfile")" 2>/dev/null || true
   done
   ```

3. Return to the default local profile:

   ```bash
   cd backend
   ./start.sh dev
   ```

4. If local infra is already healthy and you only need to restore services:

   ```bash
   cd backend
   SKIP_INFRA=1 ./start.sh dev
   ```

5. Archive the failing remote-profile logs before retrying:

   ```bash
   cd backend
   tar -czf /tmp/remote-journal-rollback-logs.tgz logs
   ```

6. Keep `/tmp/journal-remote-validation` for diffing until the incident review
   ends. The rendered remote configs are disposable, but they are useful
   evidence while comparing against the tracked local YAMLs.

## Dry-Run Drill

Use this drill before the first live remote session:

```bash
cd backend
DRY_RUN=1 REMOTE_JOURNAL_DSN='<redacted>' ./start.sh dev remote
DRY_RUN=1 ./start.sh dev
```

Or run the automated verifier from the repo root:

```bash
python3 backend/scripts/check_remote_journal_rollback_dry_run.py
```

The dry-run passes when:

- the remote profile shows only `redis`, `etcd`, and `jaeger`
- the remote profile points every service at `/tmp/journal-remote-validation`
- the local profile still shows MySQL master or replicas and the existing `cron`
  process

## Incident Notes Template

Record the following fields for every no-go call:

- `decision_time`
- `decision_owner`
- `trigger`
- `first_bad_service`
- `rollback_started_at`
- `rollback_completed_at`
- `local_profile_recovered`
- `followup_owner`

Ready-to-paste template:

```text
decision_time:
decision_owner:
trigger:
first_bad_service:
rollback_started_at:
rollback_completed_at:
local_profile_recovered:
followup_owner:
```
