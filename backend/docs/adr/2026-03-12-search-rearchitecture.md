# ADR: Search Rearchitecture With FULLTEXT Fallback

- Status: Accepted
- Date: 2026-03-12
- Owners: paper-rpc / api gateway maintainers
- Related issues: `RPT-050`, `RPT-060`, `RPT-070`

## Context

The current search path is simple and production-usable, but it is also the limit
of what the platform can explain and evolve:

- Request path: `GET /api/v1/papers/search`
- API gateway forwards directly to `paper-rpc`
- `paper-rpc` calls `PaperModel.Search`
- `PaperModel.Search` executes MySQL `MATCH(title, abstract, keywords) AGAINST(? IN BOOLEAN MODE)`
- Sorting is effectively driven by MySQL FULLTEXT score plus pagination
- There is no feature flag, shadow comparison, query explain payload, or index versioning today

That baseline is intentionally not replaced in-place. We need a design that keeps
the current FULLTEXT path as the safe default while new search capabilities are
introduced behind explicit rollout controls.

## Decision

Search will be split into explicit modules and released in phases. The default
production path remains MySQL FULLTEXT until the new engine proves both quality
and operability.

### Module boundaries

1. Query parser
   - Input: raw query, filters, sort, pagination
   - Output: normalized query, token stream, optional synonym expansion, explain seed
   - Responsibilities: whitespace cleanup, language detection, IK/Jieba segmentation, deterministic synonym expansion

2. Index builder
   - Input: paper snapshots from MySQL
   - Output: immutable segment files plus version metadata
   - Responsibilities: concurrent build, discipline sharding, rebuild orchestration, checksum/version publication

3. Inverted index store
   - Input: normalized terms + postings
   - Output: term dictionary, postings offsets, document stats
   - Responsibilities: persist postings, expose segment metadata, keep read-only snapshots for rollback

4. Retrieval orchestrator
   - Input: parsed query, feature flags, active index version
   - Output: candidate document set plus per-engine debug info
   - Responsibilities: decide between FULLTEXT, inverted index, or hybrid recall; isolate shadow traffic; enforce timeouts

5. Ranker
   - Input: recall candidates + document stats + optional business signals
   - Output: ordered result set with explain payload
   - Responsibilities: BM25 primary score, optional fusion with `shit_score`, recency, and future synonym/Trie features

6. Suggestion subsystem
   - Input: normalized terms and query prefix traffic
   - Output: Trie-based suggestions
   - Responsibilities: prefix suggestion only; no default coupling to the first batch release

7. Explain renderer
   - Input: parser output, retrieval info, ranking breakdown
   - Output: stable debug/explain structure for logs, benchmarks, and future API exposure

### Target data flow

1. Paper data remains the source of truth in MySQL.
2. Index builder reads paper snapshots by shard or discipline and writes versioned segment artifacts.
3. Segment metadata is published atomically after a successful build.
4. Search request enters the API gateway and keeps the existing request/response contract.
5. Retrieval orchestrator chooses an engine by feature flag:
   - `fulltext`: MySQL FULLTEXT only
   - `shadow`: FULLTEXT answers user traffic, new engine runs in parallel for comparison only
   - `hybrid`: new engine answers user traffic, FULLTEXT remains hot for fallback
6. Ranker returns ordered results plus explain metadata.
7. Failures, timeouts, missing index versions, or threshold misses immediately demote the request to FULLTEXT.

## Rollout and feature flags

The search rewrite must be controlled by explicit flags, not code edits:

- `search.default_engine = fulltext | hybrid`
- `search.shadow_compare = true | false`
- `search.enable_inverted_index = true | false`
- `search.enable_trie = true | false`
- `search.enable_synonyms = true | false`
- `search.enable_fusion_ranker = true | false`
- `search.active_index_version = <string>`

Rollout rules:

- `fulltext` is the only allowed default at ADR acceptance time.
- `shadow_compare=true` is required before any user-facing cutover.
- `hybrid` can become the answering path only after the benchmark thresholds below are met.
- Trie, synonyms, and fusion ranking remain independently switchable so batch-two work is not tied to batch-one release.

## Benchmark and cutover gates

The new engine is not allowed to become default without a tracked benchmark set.

### Dataset

- Start with a fixed golden set derived from real paper titles, abstracts, keywords, and discipline filters.
- Cover at least:
  - Chinese exact-match queries
  - Mixed Chinese/English academic terms
  - Discipline-constrained queries
  - Prefix queries for future Trie validation
  - Known typo or synonym cases once those features ship

### Metrics

- Recall@10 versus FULLTEXT baseline on the golden set
- Precision review on a sampled query list
- p50 / p95 search latency by engine
- Index build duration and peak memory usage
- Result stability for repeated runs on the same dataset
- Explain completeness: top-ranked results must expose tokenization and score breakdown

### Cutover threshold

The new engine may become default only when all of the following hold for the
agreed benchmark dataset:

- Recall@10 is not worse than FULLTEXT by more than 5%
- p95 latency stays within 120% of FULLTEXT baseline
- Shadow comparison shows no incorrect cross-query contamination
- Index rebuild succeeds twice on the same snapshot with identical result counts
- Explain payload is present for benchmark inspection

If any gate fails, `search.default_engine` must remain `fulltext`.

## Failure handling and rollback

Rollback is operational, not conceptual:

1. Set `search.default_engine=fulltext`
2. Disable `search.shadow_compare` if it is causing extra pressure
3. Keep the latest working index artifacts for offline inspection; do not delete them during incident handling
4. Leave the public API contract untouched so frontend routes continue to hit the same endpoint
5. Log the failing query class, index version, and engine mode before reattempting rollout

Immediate FULLTEXT fallback triggers:

- Index version missing or checksum mismatch
- Retrieval timeout
- Segment load failure
- Parser panic or tokenization error
- Ranker error
- Benchmark threshold regression after deployment

## Observability requirements

Search work is incomplete unless the following are observable:

- `search_requests_total{engine,mode,result}`
- `search_latency_ms{engine,mode}`
- `search_index_build_duration_seconds{shard}`
- `search_index_build_failures_total`
- `search_shadow_delta_total{kind}`
- `search_active_index_version`

Every search log line should carry:

- request id
- engine mode
- index version
- normalized query
- token list
- top result ids
- fallback reason if FULLTEXT was used after a new-engine attempt

## Phase mapping

### Batch 1: required before frontend integration

- concurrent index build
- deterministic tokenization (IK/Jieba)
- inverted index
- BM25 ranking
- explain output
- FULLTEXT fallback wiring

### Batch 2: optional and independently gated

- Trie suggestions
- synonym expansion
- fusion ranker

Batch 2 is not allowed to block Batch 1 release.

## Code mapping from current baseline

- API handler and gateway entry: `backend/api/internal/logic/searchPapersLogic.go`
- RPC search orchestration today: `backend/rpc/paper/internal/logic/paper/searchpaperslogic.go`
- Current FULLTEXT query: `backend/model/papermodel.go`
- Current service context without search-specific modules: `backend/rpc/paper/internal/svc/servicecontext.go`

These files remain the anchor points for the first implementation PRs. New search
modules should be introduced alongside them, not by deleting the FULLTEXT path first.

## Consequences

- Positive:
  - rollout becomes reversible
  - search quality work gets measurable gates
  - later frontend integration can switch engines without changing route contracts

- Negative:
  - service context grows before the old path can be removed
  - shadow compare increases temporary operational cost
  - benchmark data and explain payload must be maintained as product assets

## Next steps

1. Introduce config and feature-flag scaffolding in `paper-rpc`
2. Land batch-one search modules behind flags while FULLTEXT remains default
3. Run benchmark and shadow comparison
4. Only then connect frontend search controls to the new engine
