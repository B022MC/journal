# Table Governance Boundary

Date: 2026-03-13

Purpose: keep the DB merge scoped to single-db governance by default and define
exactly which tables are merely sharding candidates rather than immediate
physical split work.

## Default Rule

- Single table is the default.
- "Candidate for sharding or partitioning" is not the same as "enabled now".
- Hot or cold separation for `biz_paper` and `biz_cold_paper` stays a storage
  lifecycle rule, not a first-batch sharding trigger.

## Governance Matrix

| Table | Default governance | Trigger threshold | Route key or partition key | Archive strategy | First batch status |
| --- | --- | --- | --- | --- | --- |
| `biz_user` | Single table | Review only when row count exceeds 50 million and write QPS exceeds 300 sustained | none | cold user analytics move to offline warehouse, not online split | Keep single table |
| `biz_user_achievement` | Single table | Review only when row count exceeds 100 million or per-user lookup latency regresses | none | archive unlocked history older than 24 months to offline analytics | Keep single table |
| `biz_news` | Single table | Review only when row count exceeds 20 million and publish latency regresses | none | archive old news content by year to object storage or offline search snapshot | Keep single table |
| `biz_keyword_rule` | Single table | Review only when rule count or regex cost causes write-path latency regression | none | no partitioning; keep rule history in audit trail instead | Keep single table |
| `adm_role` | Single table | No sharding candidate in current scope | none | none | Keep single table |
| `adm_user` | Single table | Review only when admin population or login latency materially exceeds current platform scope | none | archive disabled admin accounts outside the online path if needed | Keep single table |
| `adm_permission` | Single table | No sharding candidate in current scope | none | none | Keep single table |
| `adm_role_permission` | Single table | Review only when role-permission fanout makes RBAC checks exceed SLO | none | rebuild from source-of-truth permission config before considering split | Keep single table |
| `adm_user_role` | Single table | Review only when join cardinality materially changes admin auth latency | none | rebuild from role assignments before considering split | Keep single table |
| `biz_rating` | Sharding candidate only | row count > 50 million, hottest secondary index > 40 GB, or sustained write QPS > 2,000 | `paper_id` or `paper_id hash` | age out raw write trails older than 24 months after aggregate snapshots are preserved | Candidate only |
| `biz_flag` | Sharding candidate only | row count > 20 million, sustained write QPS > 500, or moderation backlog causes hotspot indexes | `target_type + target_id hash` | archive resolved flags by time window after moderation SLA closes | Candidate only |
| `adm_audit_log` | Partition candidate only | monthly growth > 10 million rows, table > 100 million rows, or retention queries exceed SLO | monthly partition by `created_at` | drop or export partitions past retention window | Candidate only |
| `biz_paper` | Hot/cold separation only | keep using `biz_cold_paper` when cold-read ratio or storage cost requires it | none | move cold papers into `biz_cold_paper` by lifecycle | Excluded from sharding batch |
| `biz_cold_paper` | Cold archive table | not a sharding target in the first batch | none | already the archive target for cold papers | Excluded from sharding batch |

## Review Rules

- Reject any proposal that introduces physical sharding for a table not listed
  as a candidate above.
- Reject any plan that uses "data is getting big" without one of the explicit
  row-count, index-size, or write-QPS thresholds.
- Treat `biz_paper` and `biz_cold_paper` as a lifecycle pair. Do not describe
  them as a sharded set in design review.

## Current Limitation

- The thresholds above are governance gates, not proof that the current system
  has already reached them. Production or pre-production metrics are still
  required before any candidate leaves the "candidate only" state.
