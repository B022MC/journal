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

- no versioned segment artifacts or atomic active-version publication
- no `active_index_version` field or rollback cache for the latest successful build
- no structured metrics and only free-form comparison or explain logs
- benchmark inputs are synthetic fixtures, not a versioned golden set with cutover thresholds
- rollback evidence is described in docs, but not yet bound to a recurring evidence package

## Checklist

| Area | Current evidence | Status | Follow-up |
| --- | --- | --- | --- |
| Release defaults stay on FULLTEXT | `backend/rpc/paper/etc/paper.yaml` keeps `Search.DefaultEngine: fulltext`; runbook keeps `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` until gates pass | Present | Keep as invariant in every follow-up PR |
| Engine routing (`fulltext` or `hybrid`) | `Config.Normalized` and `Service.resolveEngine` only promote `hybrid` when explicitly requested; default normalizes to FULLTEXT | Present | Covered, monitor in `SB1-040` |
| Shadow compare path | `Service.Search` runs FULLTEXT as primary answer and compares the new engine in shadow mode; tests cover `ShadowCompared` behavior | Present | Extend observability in `SB1-050` |
| Explain payloads | `Snapshot.Search` returns explain data and `service_test.go` asserts explain output for Chinese queries | Present | Preserve API and log shape in `SB1-040` |
| Snapshot build stability | `TestBuildSnapshotStableAcrossRuns` proves deterministic signatures for repeated in-memory rebuilds | Present | Carry forward into versioned publication in `SB1-030` |
| Versioned segment artifacts | `BuildSnapshot` only builds an in-memory `Snapshot`; there is no persisted segment, version string, or checksum artifact | Gap | `SB1-030` |
| Atomic active-version publication | `Service.ensureSnapshot` caches one in-memory snapshot only; there is no publish step or rollback to last known-good artifact | Gap | `SB1-030` |
| `active_index_version` in runtime metadata | `BuildMetadata` exposes counts, duration, and signature, but not an operator-facing active index version | Gap | `SB1-030`, `SB1-050` |
| Structured metrics for engine, mode, result, and shadow deltas | no Prometheus counters or histograms exist under `internal/search`; current logs are free-form `logx.Infof` lines | Gap | `SB1-050` |
| Structured fallback reason taxonomy | fallback currently returns `empty_query`, `batch_one_disabled`, and raw `hybrid_error:<message>` strings | Partial | `SB1-040` |
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
