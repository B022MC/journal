# Hotspot Sharding And Partitioning PoC

Date: 2026-03-13

Purpose: define the first-batch physical split candidates without expanding the
DB merge into a full sharding rollout.

## Scope

Only these objects are in scope for first-batch physical split design:

- `biz_rating`
- `biz_flag`
- `adm_audit_log`

Everything else, especially `biz_paper` and `biz_cold_paper`, stays on the
single-db hot/cold model.

## Candidate Matrix

| Table | Physical strategy | Route key | Enable threshold | Main query cost | Aggregation cost | Archive strategy |
| --- | --- | --- | --- | --- | --- | --- |
| `biz_rating` | horizontal shards | `paper_id` or `paper_id hash` | row count > 50 million, hottest index > 40 GB, or sustained write QPS > 2,000 | paper-centric reads stay single shard because `paper_id` is the dominant lookup | user-centric views require scatter-gather or an async projection; do not make them cross-shard OLTP queries | export or compact old rating detail by time window once aggregate snapshots are materialized |
| `biz_flag` | horizontal shards | `target_type + target_id hash` | row count > 20 million, hottest index > 15 GB, or sustained write QPS > 500 | moderation lookups by target stay single shard | backlog or dashboard aggregates should use async rollups, not live fanout over every shard | archive resolved flags after SLA expiry and keep only unresolved or recent flags hot |
| `adm_audit_log` | monthly partitions first, tables only if partitions become too large | `created_at` month partition, secondary access by `actor_user_id` | monthly growth > 10 million rows, total rows > 100 million, or sustained write QPS > 500 | time-range queries prune partitions and stay cheap | cross-month reports require partition fan-in but remain bounded by the selected month range | detach or export partitions past retention and keep only recent partitions online |

## Design Rules

- No cross-shard transaction is allowed in the PoC path.
- Any user-facing aggregate that would require fanout must move to async
  rollups or cached summaries before physical split is enabled.
- The first batch must not introduce shard routing for `biz_paper` or
  `biz_cold_paper`.
- Partitioning `adm_audit_log` by month is preferred over a router until the
  monthly partition count itself becomes operationally expensive.

## Query And Rollup Notes

- `biz_rating`
  - Primary lookup: all ratings for one paper, histogram, weighted aggregate.
  - Safe because `paper_id` naturally scopes writes and reads to one shard.
  - Unsafe query to avoid in the PoC: "all ratings by one user across the whole
    corpus" as an OLTP fanout.
- `biz_flag`
  - Primary lookup: flags for one target under moderation.
  - Safe because `target_type + target_id` groups moderation traffic naturally.
  - Unsafe query to avoid in the PoC: global unresolved-flag dashboard without
    a separate rollup or search index.
- `adm_audit_log`
  - Primary lookup: time-range review by actor or target.
  - Safe because retention and compliance reviews already bound queries by time.
  - Unsafe query to avoid in the PoC: unbounded full-history scans across every
    partition.

## Exit Criteria

- Do not leave PoC mode until explain plans, benchmark results, and rollback
  steps exist for the selected candidate.
- If the threshold is not met, keep the table on the single-db path even if the
  routing design is already documented.
