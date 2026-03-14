# Remote Journal Validation Preflight

Date: 2026-03-14

Purpose: close `JRV-000` by freezing the remote single-db validation scope,
recording the disposable test fixtures, and defining a repeatable preflight
checklist before any service points at the remote `journal` schema.

## Scope Freeze

- This validation slice changes application configuration and validation flow
  only. It does not apply remote DDL, schema renames, or physical table moves.
- Business objects continue to enter through the compatibility views
  `user`, `paper`, `rating`, `news`, `flag`, and `keyword_rule`.
- Admin paths continue to use the base tables `adm_role`, `adm_permission`,
  `adm_user_role`, and `adm_audit_log`.
- Any failure in compatibility-view updatability, wrong-schema routing, or
  admin permission behavior is a rollback signal. Do not start `biz_*`
  convergence while those failures are still open.

## Required Objects

| Object | Type expected in remote `journal` | Why it matters | Live write proof owner |
| --- | --- | --- | --- |
| `user` | updatable view | register, login, `/user/info` | `JRV-050` |
| `paper` | updatable view | paper list, detail, search, report target | `JRV-050` |
| `rating` | updatable view | paper rating and rating report | `JRV-050` |
| `news` | updatable view | news list and admin news mutations | `JRV-040` |
| `flag` | updatable view | paper or rating report flow | `JRV-050` |
| `keyword_rule` | updatable view | admin keyword rule CRUD | `JRV-060` |
| `adm_role` | base table | RBAC role CRUD | `JRV-060` |
| `adm_permission` | base table | permission lookup and role binding | `JRV-060` |
| `adm_user_role` | base table | user-role assignment and revoke | `JRV-060` |
| `adm_audit_log` | base table | admin operation audit trail | `JRV-060` |

Use `python3 backend/scripts/render_remote_journal_validation_preflight.py`
to render the SQL inventory and read probes before the application rollout.
That script is intentionally read-only; write validation must use disposable
service-level fixtures so the same evidence package is reusable in later
smoke and regression issues.

## Connection And Fixture Record

Secrets stay out of git. The following record must be prepared locally before
running the remote validation:

```dotenv
REMOTE_JOURNAL_DSN=journal:<redacted>@tcp(<remote-host>:3306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
REMOTE_VALIDATION_OWNER=<name-or-slack-handle>
REMOTE_VALIDATION_DATE=2026-03-14
REMOTE_TEST_USER_EMAIL=codex+remote-validation-20260314@example.invalid
REMOTE_TEST_USER_NAME=remote-validation-20260314
REMOTE_TEST_USER_PASSWORD=<set-locally-only>
REMOTE_TEST_ADMIN_LOGIN=<set-locally-only>
REMOTE_TEST_ADMIN_PASSWORD=<set-locally-only>
REMOTE_TEST_ROLE_CODE=rv_tmp_role_20260314
REMOTE_TEST_ROLE_NAME=RV Temporary Role 20260314
REMOTE_TEST_KEYWORD_PATTERN=rv_tmp_keyword_20260314
REMOTE_TEST_NEWS_TITLE=RV temporary news 20260314
REMOTE_TEST_PAPER_ID=<existing-paper-id-for-detail-search-rating-report>
REMOTE_TEST_RATING_ID=<filled-after-rating-case>
REMOTE_TEST_FLAG_IDS=<filled-after-report-cases>
REMOTE_CLEANUP_OWNER=<same-as-REMOTE_VALIDATION_OWNER>
```

Replace every `<...>` placeholder locally before the checker runs. The checker
fails fast on placeholder values so the preflight cannot proceed with template
text or missing secrets.

The repository now prepares a local-only template at
`backend/.env.remote-validation.local`. That file is gitignored and should be
the single source for the real DSN, passwords, and sample paper id during the
validation window.

## Live Fixture Snapshot

The local-only env file is now populated and passes both:

- `python3 backend/scripts/check_remote_journal_validation_env.py`
- `backend/scripts/with_remote_journal_validation_env.sh python3 backend/scripts/check_remote_journal_validation_env.py`

The non-secret fixture summary for this validation window is:

- Validation owner and cleanup owner: `b022mc`
- Sample paper id: `1`
- Disposable role prefix: `rv_tmp_role_20260314`
- Disposable keyword prefix: `rv_tmp_keyword_20260314`
- Frontend evidence users:
  `rv-front-1773475922`, `rv-reporter-1773476041`, `rv-rater-1773476041`
- Admin evidence identity: `rv-reporter-1773476041`

Secrets still stay only in the gitignored local env file; the repository keeps
the executable shape and naming rules without storing live credentials.

## Disposable Data Rules

- Prefix every created record with `rv_tmp_` or `remote-validation-YYYYMMDD`.
- Never reuse a long-lived admin account or built-in super role.
- Prefer one disposable user, one disposable role, one disposable keyword rule,
  one disposable news item, and one known paper sample for the whole rehearsal.
- Every new identifier created during the run must be appended to the command
  log so cleanup can be replayed without guessing.

## Cleanup Contract

- `user` fixture: disable or delete the disposable account after validation.
- `rating` and `flag` fixtures: delete by captured ids or mark them with the
  validation label before removal.
- `adm_role`, `adm_user_role`, and `keyword_rule` fixtures: delete by the
  `rv_tmp_` code or pattern after regression passes.
- `news` fixture: delete the disposable news row after read and write checks.
- `adm_audit_log` is append-only evidence. Do not mutate historical rows;
  instead capture the ids or time range created by the validation window.

The generated SQL includes inventory and read probes plus cleanup query
templates for the disposable labels above. Execute the read-only checks first,
then run service-level write flows, and finally apply cleanup for disposable
rows only.

## Execution Checklist

1. Fill `backend/.env.remote-validation.local` with the real DSN, passwords, and
   sample paper id, or export the same keys in the current shell.
2. Validate that every required env var is present and matches the naming rules
   without printing secrets:
   `python3 backend/scripts/check_remote_journal_validation_env.py`
3. Render the SQL workbook:
   `python3 backend/scripts/render_remote_journal_validation_preflight.py > /tmp/remote-journal-preflight.sql`
4. Review the workbook and confirm it contains only inventory, read probes, and
   cleanup templates for disposable labels.
5. Execute the read-only section against the remote `journal` schema.
6. Run the service-level smoke or regression flows that prove write behavior:
   `JRV-040`, `JRV-050`, and `JRV-060`.
7. Capture created ids, then run only the cleanup statements that target the
   disposable labels created during the rehearsal.

If you prefer a single entry point for later commands, use the wrapper:

```bash
backend/scripts/with_remote_journal_validation_env.sh \
  python3 backend/scripts/check_remote_journal_validation_env.py
```

## Current Limitation

- The repository still does not track the live remote DSN or passwords.
  Preflight is operational through the gitignored local env file, but anyone
  reproducing the validation on another machine still needs their own local
  secret injection before the checker or wrapper can pass.
