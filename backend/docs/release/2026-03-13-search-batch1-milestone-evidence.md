# Search Batch 1 Milestone Evidence Package

Date: 2026-03-13

Purpose: keep Milestone D1 to D4 gated by one recurring evidence ledger so the
release default cannot move on undocumented proof or one-off judgment.

## Release Invariants

- `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` remains the release default until D4
  is explicitly signed off.
- `backend/rpc/paper/etc/paper.yaml` keeps `Search.DefaultEngine: fulltext`
  until the frontend route-level default is already stable.
- Explicit `/papers?query=...&engine=fulltext` or `engine=hybrid` URLs are
  validation overrides only; they are not evidence that the release default has
  changed.

## Milestone Ledger

| Milestone | Definition | Current status | Required recurring evidence | Latest recorded evidence |
| --- | --- | --- | --- | --- |
| D1 | Versioned index publication, validation, and FULLTEXT fallback are complete. | Completed | backend tests, `make search-bench`, frontend smoke, FULLTEXT rollback drill | `SB1-030`, `SB1-040`, and the 2026-03-13 command set below |
| D2 | Metrics, logs, benchmark thresholds, and `/papers` validation entry are complete. | Completed | backend tests, `make search-bench`, frontend smoke, FULLTEXT rollback drill | `SB1-050`, `SB1-060`, `SB1-070`, and the 2026-03-13 command set below |
| D3 | Shadow compare stays clean on the agreed dataset and leaves recurring evidence without cross-query contamination. | In progress | backend tests, `make search-bench`, frontend smoke, FULLTEXT rollback drill, recurring shadow-compare evidence bundle | Pending a recurring shadow-compare archive; keep the release default on `fulltext` |
| D4 | Release captain has a complete sign-off package and may consider `hybrid` as a default candidate. | Blocked on D3 | backend tests, `make search-bench`, frontend smoke, FULLTEXT rollback drill, release sign-off note | Blocked until D3 remains green across repeated checks |

## Command Set Executed On 2026-03-13

Backend validation:

- `cd backend && GOCACHE=/tmp/journal-go-build go test ./rpc/paper/internal/search ./api/internal/types`
  - Result: passed
- `cd backend && GOCACHE=/tmp/journal-go-build BENCHTIME=1x make search-bench`
  - Result: passed
  - Golden summary: `recall_at_10 hybrid=1.00 fulltext=1.00 latency_ms_p50 hybrid=0.024 fulltext=0.030 latency_ms_p95 hybrid=0.045 fulltext=0.079 explain_completeness=1.00 rebuild_stable=true version=idx-ca0086dceef1`

Frontend validation:

- `cd frontend && npm run lint`
  - Result: passed
- `cd frontend && npx next build --webpack`
  - Result: passed on 2026-03-13 while stabilizing `/papers` validation routing
- `cd frontend && JOURNAL_SEARCH_RELEASE_ENGINE=hybrid npm run smoke`
  - Result: passed
- `cd frontend && JOURNAL_SEARCH_RELEASE_ENGINE=fulltext npm run smoke`
  - Result: passed

## FULLTEXT Rollback Drill On 2026-03-13

1. Started the production smoke flow with `JOURNAL_SEARCH_RELEASE_ENGINE=hybrid`
   and verified the `/papers` surface stayed reachable.
2. Re-ran the same smoke flow with
   `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext`.
3. Verified the release-default path `/papers?query=...&engine=auto` rendered
   the `Release default engine:` marker with the current env value on both runs,
   which proves the default route can move back to FULLTEXT without route or
   contract changes.
4. Verified the explicit validation URL
   `/papers?query=...&engine=hybrid&shadow_compare=true` stayed reachable during
   the drill, keeping overrides separate from the default path.

## Promotion Rule

- Do not promote the release default while any D-stage row above is not marked
  Completed.
- Treat missing evidence in this ledger as a failed gate, even if individual
  code changes look complete.
- Append the next command set and rollback drill record to this file before any
  future `JOURNAL_SEARCH_RELEASE_ENGINE` promotion.
