# DB Merge Batch 1 Runtime Convergence

Date: 2026-03-14

Purpose: close `NRJ-080` by moving the batch-1 hot runtime models from the
compatibility view names to the `biz_*` tables and updating the legacy naming
baseline in the same commit.

## Scope

- `PaperModel` and `PaperModel` admin helpers now target `biz_paper`
- `UserModel` and `UserModel` admin helpers now target `biz_user`
- `RatingModel` now targets `biz_rating` and joins `biz_user`
- `FlagModel` admin and public paths now target `biz_flag`
- `KeywordRuleModel` now targets `biz_keyword_rule`
- `NewsModel` now targets `biz_news`
- `adm_*`, migrations, schema bootstrap, and `cold_paper` stay outside this
  batch

## Naming Freeze Result

- Before the rewrite: `db naming freeze OK: 64 tracked tuples, 309 legacy hits`
- After the rewrite and baseline update:
  `db naming freeze OK: 54 tracked tuples, 146 legacy hits`
- Baseline updated in:
  `backend/docs/release/2026-03-13-db-merge-legacy-baseline.csv`

## Verification

- Compile and package sweep:
  `cd backend && go test ./...`
- Freeze guard:
  `cd backend && python3 scripts/check_legacy_db_refs.py`
- Frontend smoke output: `/tmp/nrj080-frontend-smoke.json`
  - registered and logged in disposable user `rv-biz-batch1-1773477523`
  - `/api/v1/papers` and `/api/v1/papers/1` returned paper `1`
  - `/api/v1/news` returned HTTP `200`
  - `POST /api/v1/papers/1/rate` created rating `4`
  - `POST /api/v1/papers/1/flag` created flag `4`
- Admin smoke output: `/tmp/nrj080-admin-smoke.json`
  - `/api/v1/admin/users` listed user `11`, then same-value status and role
    updates returned success
  - `/api/v1/admin/papers` listed paper `1`, and same-value zone update
    returned success
  - `/api/v1/admin/flags?status=-1` showed flag `4`
  - disposable role `5` was created and deleted
  - disposable keyword rule `5` was created, listed, deleted, and recorded in
    audit logs

## Residual Risk

- The remote admin permission inventory still does not expose
  `admin.keyword.manage` in `/api/v1/admin/me/permissions` or
  `/api/v1/admin/permissions`; keyword rule CRUD still works because the
  disposable admin identity is backed by a super role.
