# Read-Write Split Rollout Matrix

Date: 2026-03-13

Purpose: define which journal read paths may use replicas, which paths must stay
on the primary, and which config toggles or drills must be executed before any
service enables replica reads in production.

## Service Toggles

| Service config | Current toggle | Intended use |
| --- | --- | --- |
| `api/etc/journal-api.yaml` | `ReadWriteSplit: false` for both `BizDB` and `AdminDB` | Keep the gateway on primary until M3 and M4 evidence is green |
| `admin-api/etc/admin-api.yaml` | `ReadWriteSplit: false` for both `BizDB` and `AdminDB` | Keep admin gateway on primary until permission and audit-log drills are green |
| `rpc/paper/etc/paper.yaml` | `BizDB.ReadWriteSplit: true` | Candidate for paper list, detail, and search-oriented reads |
| `rpc/news/etc/news.yaml` | `BizDB.ReadWriteSplit: true` | Candidate for news list and detail reads |
| `rpc/rating/etc/rating.yaml` | `BizDB.ReadWriteSplit: true` | Candidate for rating aggregate and list reads after lag drills |
| `rpc/user/etc/user.yaml` | `BizDB.ReadWriteSplit: true` | Use carefully; registration uniqueness and write-after-read flows still pin to primary |
| `rpc/admin/etc/admin.yaml` | `BizDB.ReadWriteSplit: true`, `AdminDB.ReadWriteSplit: true` | Only list or lookup flows may use replicas; permission checks and admin mutations stay primary-first |

## Replica-Eligible Flows

| Flow | Code anchor | Why replica-safe |
| --- | --- | --- |
| Paper list and author list | `backend/model/papermodel.go` `List` and `ListByAuthor` | Read-mostly browsing paths tolerate short lag and already avoid explicit primary pinning |
| News list and detail | `backend/model/newsmodel.go` `List` and `FindById` | Public news reads are append-heavy and do not require read-after-write consistency for end users |
| Rating list and aggregate views | `backend/model/ratingmodel.go` `ListByPaper`, `GetPaperRatingStats`, `GetWeightedRatingStats`, `GetPaperScoreHistogram` | Aggregates are read-mostly and can tolerate bounded lag during rollout drills |
| Search snapshot document reads | `backend/model/papermodel.go` `ListSearchDocuments` | Search indexing is already an asynchronous workload and should not force gateway traffic onto primary |

## Primary-Only Or Explicit Primary Flows

| Flow | Code anchor | Why it stays on primary |
| --- | --- | --- |
| Registration uniqueness checks | `backend/model/usermodel.go` `ExistsByUsernamePrimary` and `ExistsByEmailPrimary` | Write-after-read path; replica lag would allow duplicate usernames or emails |
| Login and user hydration after writes | `backend/model/usermodel.go` `FindByIdPrimary`; `backend/admin-api/internal/logic/manage/adminauth.go` | Session bootstrap and admin auth must see the newest role or status change |
| Permission checks and role resolution | `backend/model/adminrbacmodel.go` `HasPermission`, `ListPermissionCodesByUserId`, `HasAnyAdminRole` | Lag here is a security regression, not just stale UI |
| Rating fraud and post-write checks | `backend/model/ratingmodel.go` `FindByIdPrimary`, `CountByUserSince`, `CountDistinctUsersByPaperIPSince`, `CountDistinctUsersByPaperFingerprintSince` | Anti-abuse logic must observe the latest writes before allowing more traffic |
| Paper moderation or zone update readback | `backend/model/papermodel.go` `FindByIdPrimary`; `backend/rpc/paper/internal/logic/paper/updatezonelogic.go` | Moderation and write confirmation must read their own write |

## Rollout Order

1. Keep `api` and `admin-api` on primary-only while validating the RPC layer.
2. Start with `news-rpc` and paper list or search reads, because these are read-heavy and public.
3. Expand to rating aggregate reads only after replica-lag drills prove metrics and rollback steps.
4. Keep registration, login, permission checks, admin mutations, and any write-after-read path on primary for the entire rollout.

## Failure Drills

| Drill | How to execute | Pass criteria | Rollback switch |
| --- | --- | --- | --- |
| Replica lag | Introduce lag on one replica, then hit paper list, news list, rating aggregate, login, and permission paths | Replica-safe paths may be briefly stale; primary-only paths remain correct | Set `ReadWriteSplit: false` in the affected service YAML and restart the service |
| Replica unavailable | Stop one replica or break replica DSN resolution | Services recover by removing replica reads from the affected service and keeping primary traffic healthy | Same service-level `ReadWriteSplit: false` toggle plus service restart |
| Query timeout under replica load | Run rating aggregate and paper list against a degraded replica | Service-level rollback path restores primary reads fast enough to keep the user-visible path healthy | Disable split for that service, keep explicit primary-read methods unchanged |
| Write-after-read validation | Register a user, rate a paper, update admin permissions, then read back immediately | Primary-pinned methods observe the new state without waiting for replication | None; these paths should never leave primary during rollout |

## Monitoring And Evidence

- Collect MySQL replica lag or health evidence outside the app, for example
  `Seconds_Behind_Master` or equivalent managed-mysql metrics.
- Keep service restart records showing when `ReadWriteSplit` was toggled on or
  off for each service.
- Archive logs or request traces for paper list, news list, rating aggregate,
  login, and permission checks during every lag or outage drill.
- Treat any permission mismatch, duplicate registration, or write-after-read
  inconsistency as an automatic rollback trigger.
