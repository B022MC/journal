# Legacy DB Reference Batch 1 Plan

Date: 2026-03-14

Purpose: close `JRV-080` by freezing the current legacy-reference baseline and
defining the first `biz_*` convergence batch without touching `adm_*`.

## Baseline Status

`cd backend && python3 scripts/check_legacy_db_refs.py` currently reports:

```text
db naming freeze OK: 64 tracked tuples, 309 legacy hits remain within baseline 2026-03-13-db-merge-legacy-baseline.csv
```

This means the freeze guard is still reproducible and no new legacy reference
escaped the audited baseline after the remote-validation prep work.

## Excluded From Batch 1

- All `adm_*` tables and `backend/model/adminrbacmodel.go`
- Historical migrations and `model/schema.sql`
- `deploy/k8s/middleware/init/mysql-init.yaml` compatibility-view bootstrap
- `user_achievement` and `cold_paper`, which are lower-frequency follow-ups
  after the main user or paper paths have moved

## Batch 1 Runtime Targets

| Legacy object | Primary file(s) | Current baseline hits | Why it is batch 1 |
| --- | --- | --- | --- |
| `paper` | `backend/model/papermodel.go`, `backend/model/papermodel_admin.go` | 49 + 8 | Highest runtime concentration and shared by list, detail, submit, moderation, and search |
| `rating` | `backend/model/ratingmodel.go` | 32 | High-write and high-read path for scoring, aggregates, and anti-abuse checks |
| `user` | `backend/model/usermodel.go`, `backend/model/usermodel_admin.go` | 28 + 8 | Registration, login, profile, and admin identity all depend on it |
| `flag` | `backend/model/flagmodel.go`, `backend/model/flagmodel_admin.go` | 12 + 6 | Shared by user report flows and admin moderation |
| `keyword_rule` | `backend/model/keywordrulemodel.go` | 8 | Critical for admin keyword filtering and paper-side degradation checks |
| `news` | `backend/model/newsmodel.go` | 8 | Small enough to converge early and shared by public plus admin paths |

## Batch 1 Rewrite Order

1. `PaperModel` and `UserModel`
2. `RatingModel`
3. `FlagModel`
4. `KeywordRuleModel`
5. `NewsModel`

This order keeps the first pass focused on the hottest shared runtime models
before touching lower-frequency helpers.

## Rewrite Rules

- Replace bare business table names with `biz_*` only inside the batch-1 model
  or SQL paths listed above.
- Keep compatibility views in place until `JRV-040`, `JRV-050`, and `JRV-060`
  have real green evidence on the remote profile.
- Update `2026-03-13-db-merge-legacy-baseline.csv` only in the same commit as
  each actual code rewrite so the guard stays count-accurate.
- Do not move `adm_role`, `adm_permission`, `adm_user_role`, or `adm_audit_log`
  during batch 1. Their runtime boundary stays on `adm_*`.

## Expected Follow-Ups

- `user_achievement` can move after `user` and `paper` stop using the legacy
  names in the hot path.
- `cold_paper` stays behind the paper batch because it is a lifecycle archive
  path rather than the primary validation blocker.
- Migration SQL cleanup belongs to a later batch after the runtime layer has
  already stopped depending on compatibility views.
