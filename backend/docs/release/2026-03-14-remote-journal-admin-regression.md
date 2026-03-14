# Remote Journal Admin Regression

Date: 2026-03-14

Purpose: close `NRJ-050` by proving the minimum remote admin regression against
the live remote-profile stack with disposable data only.

## Environment

- Command: `cd backend && env GOROOT=/Users/b022mc/sdk/go1.25.6 PATH=/Users/b022mc/sdk/go1.25.6/bin:$PATH GOCACHE=/tmp/journal-go-build SKIP_INFRA=1 NO_RESTART=1 scripts/with_remote_journal_validation_env.sh ./start.sh dev remote`
- Admin API: `http://127.0.0.1:8889/api/v1/admin`
- Disposable admin identity: `rv-reporter-1773476041`
- Evidence snapshot: `/tmp/nrj050-admin-regression.json`

## Evidence

1. Admin login succeeded with the disposable remote account and returned a JWT
   whose claims use `user_id`.
2. `GET /api/v1/admin/me/permissions` returned `is_super=true`; visible
   permissions included `admin.role.manage`, `admin.audit.view`,
   `admin.paper.zone.update`, `admin.role.view`, `admin.user.manage`, and
   `admin.user.view`.
3. `GET /api/v1/admin/papers?page=1&page_size=3` returned disposable paper `1`
   and `PUT /api/v1/admin/papers/1/zone` with the same `latrine` zone returned
   `{"success":true,"message":"zone updated"}`.
4. Role CRUD and RBAC binding passed with disposable role
   `rv_tmp_role_1773476924`:
   `POST /roles` -> role `4`;
   `PUT /roles/4` updated the name and description;
   `PUT /roles/4/permissions` set permission ids `[8,10,12]`;
   `POST /users/8/roles` assigned the role to `rv-rater-1773476041`;
   `DELETE /users/8/roles/4` revoked it;
   `DELETE /roles/4` removed the disposable role.
5. Keyword rule CRUD passed with disposable rule
   `rv_tmp_keyword_1773476924`:
   `POST /keyword-rules` -> rule `4`;
   `DELETE /keyword-rules/4` removed it;
   a second `DELETE /keyword-rules/4` returned
   `{"success":false,"message":"keyword rule not found"}` as the expected
   non-existent-object regression case.
6. `GET /api/v1/admin/flags?status=0` returned pending test flags and
   `PUT /api/v1/admin/flags/3/resolve` moved disposable rating flag `3` from
   `status=0` to `status=1`.
7. `GET /api/v1/admin/audit-logs?page=1&page_size=50` returned two persisted
   audit rows for keyword rule `4`:
   `create keyword rule` and `delete keyword rule`, both by actor `7`.

## Cleanup

- Disposable role `4` was revoked from user `8` and deleted.
- Disposable keyword rule `4` was deleted.
- Disposable flag `3` remains resolved because the regression intentionally
  exercised the admin resolution path.

## Residual Risk

- `GET /api/v1/admin/me/permissions` and `GET /api/v1/admin/permissions` did
  not expose `admin.keyword.manage`, but super-admin keyword rule operations and
  audit logs still passed because permission checks bypass for `is_super=1`.
  This points to remote permission inventory drift rather than a runtime block,
  but a non-super role cannot currently discover or bind that permission from
  the live permission list.
