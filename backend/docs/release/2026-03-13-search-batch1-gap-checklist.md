# Search Batch 1 Gap Checklist

Date: 2026-03-13

Purpose: map the current `paper-rpc` search implementation to the accepted ADR
and the Batch 1 cutover gates, so follow-up issues can land against a fixed
checklist instead of re-auditing the baseline.

## Audit Summary

The codebase already contains a usable hybrid-search skeleton:

- `fulltext` and `hybrid` routing already exists in the search service
- shadow compare already keeps FULLTEXT as the answering path
- explain payloads and deterministic snapshot tests already exist
- `make search-bench` already exercises snapshot build and service-search benchmarks

The remaining Batch 1 gaps are operational rather than conceptual:

- benchmark inputs are synthetic fixtures, not a versioned golden set with cutover thresholds
- rollback evidence is described in docs, but not yet bound to a recurring evidence package

## Checklist

| Area | Current evidence | Status | Follow-up |
| --- | --- | --- | --- |
| Release defaults stay on FULLTEXT | `backend/rpc/paper/etc/paper.yaml` keeps `Search.DefaultEngine: fulltext`; runbook keeps `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` until gates pass | Present | Keep as invariant in every follow-up PR |
| Engine routing (`fulltext` or `hybrid`) | `Config.Normalized` and `Service.resolveEngine` only promote `hybrid` when explicitly requested; default normalizes to FULLTEXT | Present | Covered, monitor in `SB1-040` |
| Shadow compare path | `Service.Search` keeps FULLTEXT as the answer path, logs compare-path outcomes separately, and only marks `ShadowCompared` when the comparison actually succeeds | Present | Extend observability in `SB1-050` |
| Explain payloads | `Snapshot.Search` returns explain data and `service_test.go` asserts explain output for Chinese queries | Present | Preserve API and log shape in `SB1-040` |
| Snapshot build stability | `TestBuildSnapshotStableAcrossRuns` proves deterministic signatures for repeated in-memory rebuilds | Present | Carry forward into versioned publication in `SB1-030` |
| Versioned segment artifacts | `BuildMetadata` now carries deterministic `version`, `checksum`, and per-discipline segment metadata for every build artifact | Present | Completed by `SB1-030`; keep exposing it in logs and metrics |
| Atomic active-version publication | `Service` now builds a candidate artifact, validates it, and only then swaps the active pointer while retaining the latest successful artifact cache | Present | Completed by `SB1-030`; preserve this publish order |
| `active_index_version` in runtime metadata | active and cached artifacts now expose a stable `version` field via `BuildMetadata` for follow-up observability work | Present | Wire into logs and metrics in `SB1-050` |
| Structured metrics for engine, mode, result, and shadow deltas | search now registers request, latency, build, shadow-delta, and active-version collectors; answer, explain, shadow, and compare logs share request id, engine or mode, index version, normalized query, token list, top ids, and fallback reason fields | Present | Completed by `SB1-050`; keep label cardinality bounded |
| Structured fallback reason taxonomy | fallback now maps timeout, missing version, segment load or validation failure, query-parse failure, ranker failure, and build failure onto stable reason enums instead of raw error text | Present | Completed by `SB1-040`; keep contract stable |
| Golden benchmark dataset and thresholds | `benchmark_test.go` uses generated synthetic documents; `make search-bench` does not yet load a versioned golden set or assert Recall@10 / latency thresholds | Gap | `SB1-060` |
| Rollback evidence package | runbook documents rollback steps, but no recurring artifact bundle links benchmark, shadow compare, smoke, and rollback drill together | Gap | `SB1-080` |
| Batch 2 features do not block Batch 1 | Trie, synonym, and fusion toggles already exist in config and tests, but they are not required for the cutover gate | Out of scope | Keep explicitly non-blocking in `SB1-090` |

## Existing Commands Used For This Audit

- `cd backend && go test ./rpc/paper/internal/search -run 'Test(BuildSnapshotStableAcrossRuns|SearchFallsBackToFulltextWhenIndexBuildFails|SearchShadowCompareKeepsFulltextPrimary)$'`
- `cd backend && BENCHTIME=1x make search-bench`

## Key Code Anchors

- `backend/rpc/paper/internal/search/config.go`
- `backend/rpc/paper/internal/search/service.go`
- `backend/rpc/paper/internal/search/index.go`
- `backend/rpc/paper/internal/search/service_test.go`
- `backend/rpc/paper/internal/search/benchmark_test.go`
- `backend/Makefile`
