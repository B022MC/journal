# Remote Journal Entrypoint Matrix

Date: 2026-03-14

Purpose: close `JRV-010` by inventorying every runtime entrypoint that must
switch together for remote single-db validation. A partial switch is not valid
because both API gateways still keep direct DB model access in addition to RPC
clients.

## Cutover Rule

- Remote validation is green only when `journal-api`, `admin-api`, `user.rpc`,
  `paper.rpc`, `rating.rpc`, `news.rpc`, and `admin.rpc` all point at the same
  remote `journal` DSN.
- `journal-api` is not a pure gateway. Reporting, admin-permission lookup, and
  some write-after-read safeguards still hit its local DB connection.
- `admin-api` is not a pure gateway either. Admin login, permission checks, and
  keyword-rule CRUD also hit its local DB connection before or alongside RPCs.
- Therefore `JRV-020` must apply the remote DB overlay to both APIs and all
  five RPCs in one batch.

## Service Inventory

| Service | Config | Dev start command | Prod start command | Required deps | Validation owner |
| --- | --- | --- | --- | --- | --- |
| `journal-api` | `backend/api/etc/journal-api.yaml` | `go run api/journal.go -f api/etc/journal-api.yaml` | `./bin/api -f api/etc/journal-api.yaml` | remote `journal` DB, redis, etcd, otel, `user.rpc`, `paper.rpc`, `rating.rpc`, `news.rpc`, `admin.rpc` | `JRV-050` |
| `admin-api` | `backend/admin-api/etc/admin-api.yaml` | `go run admin-api/admin.go -f admin-api/etc/admin-api.yaml` | `./bin/admin-api -f admin-api/etc/admin-api.yaml` | remote `journal` DB, redis, etcd, otel, `admin.rpc`, `news.rpc`, `paper.rpc`, `user.rpc` | `JRV-060` |
| `user.rpc` | `backend/rpc/user/etc/user.yaml` | `go run rpc/user/user.go -f rpc/user/etc/user.yaml` | `./bin/user-rpc -f rpc/user/etc/user.yaml` | remote `journal` DB, etcd, otel | `JRV-040` |
| `paper.rpc` | `backend/rpc/paper/etc/paper.yaml` | `go run rpc/paper/paper.go -f rpc/paper/etc/paper.yaml` | `./bin/paper-rpc -f rpc/paper/etc/paper.yaml` | remote `journal` DB, redis, etcd, otel | `JRV-040` |
| `rating.rpc` | `backend/rpc/rating/etc/rating.yaml` | `go run rpc/rating/rating.go -f rpc/rating/etc/rating.yaml` | `./bin/rating-rpc -f rpc/rating/etc/rating.yaml` | remote `journal` DB, redis, etcd, otel | `JRV-040` |
| `news.rpc` | `backend/rpc/news/etc/news.yaml` | `go run rpc/news/news.go -f rpc/news/etc/news.yaml` | `./bin/news-rpc -f rpc/news/etc/news.yaml` | remote `journal` DB, redis, etcd, otel | `JRV-040` |
| `admin.rpc` | `backend/rpc/admin/etc/admin.yaml` | `go run rpc/admin/admin.go -f rpc/admin/etc/admin.yaml` | `./bin/admin-rpc -f rpc/admin/etc/admin.yaml` | remote `journal` DB, etcd, otel | `JRV-040` |

## Frontend Flow Mapping

| User-visible flow | HTTP entry | Concrete service chain | Notes |
| --- | --- | --- | --- |
| Register and login | `/api/v1/user/register`, `/api/v1/user/login` | `journal-api -> user.rpc`, then `journal-api` local `AdminRBAC` for admin permission enrichment | Login is not fully isolated to `user.rpc`; the gateway also reads RBAC locally. |
| `/user/info` | `/api/v1/user/info` | `journal-api -> user.rpc`, then `journal-api` local `AdminRBAC` | Same direct DB dependency as login. |
| Paper list, detail, search, user papers | `/api/v1/papers`, `/api/v1/papers/:id`, `/api/v1/papers/search`, `/api/v1/users/:id/papers`, `/api/v1/user/papers` | `journal-api -> paper.rpc` | `paper.rpc` also loads `keyword_rule`, `flag`, `user`, and `rating` models locally. |
| Paper rating and rating list | `/api/v1/papers/:id/rate`, `/api/v1/papers/:id/ratings`, `/api/v1/user/ratings` | `journal-api -> rating.rpc`, plus `journal-api` local `RatingModel` fingerprint update on rate | Rating write proof must cover both RPC and gateway DB access. |
| Paper or rating report | `/api/v1/papers/:id/flag`, `/api/v1/ratings/:id/flag` | `journal-api` local `FlagService -> FlagModel/PaperModel/RatingModel/UserModel` | No RPC hop here. Missing the API DB overlay will leave report flows on the wrong DB. |
| Public news | `/api/v1/news`, `/api/v1/news/:id` | `journal-api -> news.rpc` | Needed even if only admin creates news. |

## Admin Flow Mapping

| Admin flow | HTTP entry | Concrete service chain | Notes |
| --- | --- | --- | --- |
| Admin login | `/api/v1/admin/login` | `admin-api` local `UserModel -> AdminRBAC` | No RPC hop. This is a direct proof that `admin-api` must switch DB with the RPCs. |
| My permissions | `/api/v1/admin/me/permissions` | `admin-api` local `AdminRBAC` | Also direct DB. |
| Keyword-rule list, create, delete | `/api/v1/admin/keyword-rules` | `admin-api` local `KeywordRuleModel`, then `admin.rpc` audit logging | This path validates both local DB and `admin.rpc`. |
| Role CRUD and permission binding | `/api/v1/admin/roles*`, `/api/v1/admin/permissions` | `admin-api -> admin.rpc` | Core RBAC chain. |
| User-role assign or revoke | `/api/v1/admin/users/:id/roles*`, `/api/v1/admin/users/:id/role` | `admin-api -> admin.rpc` | Covers `adm_user_role`. |
| Audit log list | `/api/v1/admin/audit-logs` | `admin-api -> admin.rpc` | Covers `adm_audit_log`. |
| Flag list and resolve | `/api/v1/admin/flags*` | `admin-api -> admin.rpc` | Covers moderation loop. |
| Admin create news | `/api/v1/admin/news` | `admin-api -> news.rpc` | `news.rpc` is shared by public and admin paths. |
| Admin paper status and zone | `/api/v1/admin/papers`, `/api/v1/admin/papers/:id/status`, `/api/v1/admin/papers/:id/zone` | `admin-api -> admin.rpc` for list or status, `admin-api -> paper.rpc` for zone update | Paper admin paths are split across two RPCs. |

## Startup And Verification Order

The repository startup script already encodes the runnable commands for all
services. For remote validation, keep the service order below so downstream
RPC discovery is ready before the APIs start:

1. `user.rpc`
2. `paper.rpc`
3. `rating.rpc`
4. `news.rpc`
5. `admin.rpc`
6. `journal-api`
7. `admin-api`

For each service capture:

- the exact config path used
- the command used to start it
- the DSN target visible in the effective config or startup log
- the downstream dependency it needs to reach next
- the matching validation owner from `JRV-040`, `JRV-050`, or `JRV-060`

## Current Risk

- If only the five RPC YAMLs are switched, paper or rating report flows still
  hit the old DB through `journal-api`.
- If only `admin.rpc` is switched, admin login and keyword-rule CRUD still hit
  the old DB through `admin-api`.
- If `news.rpc` is omitted, public news and admin news creation will diverge
  even when the rest of the stack points at the remote DB.
