# Main-Site Rollout And Rollback Runbook

Date: 2026-03-12

Primary owner: release captain on current deploy

Escalation owners:

- Homepage routing or shell regressions: frontend maintainer on duty
- Search quality or fallback regressions: search maintainer on duty

## Release Flags

Homepage flag:

- `JOURNAL_HOME_VARIANT=main` keeps `/` on the live editorial desk
- `JOURNAL_HOME_VARIANT=roadmap` rolls `/` back to the legacy roadmap surface

Search flag:

- `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` keeps `/papers?query=...` on MySQL FULLTEXT by default
- `JOURNAL_SEARCH_RELEASE_ENGINE=hybrid` sends search traffic to the new hybrid engine by default
- `JOURNAL_SEARCH_RELEASE_ENGINE=auto` delegates to the backend default and should only be used when the backend rollout state is already controlled and observed

Backend safety default:

- `backend/rpc/paper/etc/paper.yaml` keeps `Search.DefaultEngine: fulltext`
- `backend/docs/adr/2026-03-12-search-rearchitecture.md` defines the quality gates required before changing that backend default

## Milestones

| Milestone | Definition | Status on 2026-03-13 | Evidence |
| --- | --- | --- | --- |
| A | Design baseline and scope are frozen; roadmap remains available as rollback content, not the default delivery baseline. | Completed | `RPT-010`, `frontend/FRONTEND_DESIGN.md`, issues CSV |
| B | Core homepage, list, detail, login, and register routes are live on the main shell. | Completed | `RPT-020`, `RPT-030`, `npm run smoke`, `npx next build --webpack` |
| C | Protected submit/workspace flow and public user pages are live behind the shared auth/session bridge. | Completed | `RPT-040`, `go test ./api/... ./rpc/... ./model`, cookie bridge drill |
| D1 | Versioned index publication, validation, and FULLTEXT fallback are complete. | Completed | `backend/docs/release/2026-03-13-search-batch1-milestone-evidence.md` |
| D2 | Metrics, logs, benchmark thresholds, and `/papers` validation entry are complete. | Completed | `backend/docs/release/2026-03-13-search-batch1-milestone-evidence.md` |
| D3 | Shadow compare remains clean on the agreed dataset and leaves recurring evidence. | In progress | `backend/docs/release/2026-03-13-search-batch1-milestone-evidence.md` |
| D4 | Release captain has a complete sign-off package and may consider `hybrid` as a default candidate. | Blocked on D3 | `backend/docs/release/2026-03-13-search-batch1-milestone-evidence.md` |

## Milestone D Scope And Sign-Off

Milestone D covers only Batch 1 search cutover work. The following items are
explicitly out of scope and must not block sign-off:

- SSR or SEO work
- UI enhancement work beyond `/papers` validation controls and release-default messaging
- governance or visualization features
- Batch 2 search capabilities such as Trie, synonym expansion, and fusion ranking
- irreversible schema or storage changes bundled with a flag flip

Milestone D ownership is split by evidence surface:

- release captain: final go or no-go decision, release-flag change, rollback drill record
- search maintainer: `paper-rpc` benchmark, fallback, shadow-compare, and observability evidence
- frontend maintainer: `/papers` validation entry, release-default engine display, and rollback messaging

The detailed owner handoff matrix, current Batch 1 blocking list, and Batch 2
parking lot live in
`backend/docs/release/2026-03-13-search-batch1-owner-matrix.md`. Keep that file
in sync before any cutover review.

Milestone D cannot promote the release default while either
`JOURNAL_SEARCH_RELEASE_ENGINE` or `Search.DefaultEngine` has already moved away
from `fulltext` before the checklist below is complete.

### Milestone D Evidence Checklist

Every sign-off package must include all of the following:

1. `cd backend && go test ./api/... ./rpc/... ./model`
2. `cd backend && BENCHTIME=1x make search-bench`
3. shadow-compare evidence showing no cross-query contamination on the agreed golden set
4. `cd frontend && npm run lint`
5. `cd frontend && npx next build --webpack`
6. `cd frontend && npm run smoke`
7. one FULLTEXT rollback drill record showing `/papers?query=...` returns to the stable FULLTEXT default without route or contract changes

### Milestone D Evidence Ledger

Update `backend/docs/release/2026-03-13-search-batch1-milestone-evidence.md`
after each D-stage boundary. Missing command output, benchmark summaries, smoke
results, or rollback-drill records means the current stage is not signed off.

## Pre-Release Checklist

Run these before changing either release flag:

1. `cd backend && go test ./api/... ./rpc/... ./model`
2. `cd frontend && npm run lint`
3. `cd frontend && npx next build --webpack`
4. `cd frontend && npm run smoke`
5. If considering `JOURNAL_SEARCH_RELEASE_ENGINE=hybrid`, also run `cd backend && BENCHTIME=1x make search-bench`

Do not combine irreversible database/schema changes with a homepage or search flag flip in the same release step.

## Rollout Steps

### Homepage

1. Keep `JOURNAL_HOME_VARIANT=main` for the normal release path.
2. Verify `/` still serves the editorial desk.
3. If a homepage regression appears, switch `JOURNAL_HOME_VARIANT=roadmap` and reload the frontend process. No route code change is required.

### Search

1. Keep `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` until the hybrid engine satisfies the ADR cutover gates.
2. When the gates pass, switch `JOURNAL_SEARCH_RELEASE_ENGINE=hybrid` for the frontend route-level default.
3. Only after the frontend route-level default is stable should `Search.DefaultEngine` in paper-rpc be considered for promotion beyond `fulltext`.
4. Treat `/papers?query=...&engine=auto` as the release-default path that follows `JOURNAL_SEARCH_RELEASE_ENGINE`.
5. Treat explicit `/papers?query=...&engine=fulltext` or `engine=hybrid` URLs, plus `shadow_compare=true`, as validation overrides only; they must not be used as evidence that the default route has changed.

## Rollback Triggers

Rollback the homepage flag immediately when any of the following appears:

- `/` fails smoke or renders a broken shell
- navigation no longer reaches `/papers`, `/login`, or `/submit`
- critical SSR error replaces the homepage content

Rollback search immediately to FULLTEXT when any of the following appears:

- benchmark gate regression from the search ADR
- cross-query contamination in shadow comparison
- hybrid latency or timeout incidents
- incorrect fallback behavior or missing explain/debug evidence
- user-facing result instability after sort, refresh, or back/forward navigation

## Rollback Procedure

Homepage rollback:

1. Set `JOURNAL_HOME_VARIANT=roadmap`
2. Reload the frontend process
3. Verify `/` serves the roadmap surface and `/papers` remains reachable

Search rollback:

1. Set `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext`
2. Keep `Search.DefaultEngine=fulltext` in paper-rpc
3. Reload the affected services
4. Verify `/papers?query=...&engine=auto` shows FULLTEXT as the active default while explicit override URLs remain validation-only paths

## Rollback Drill Record

Executed on 2026-03-12 in local dev:

- Started the frontend with `JOURNAL_HOME_VARIANT=roadmap JOURNAL_SEARCH_RELEASE_ENGINE=fulltext npm run dev -- --hostname 127.0.0.1 --port 3000`
- Verified `/` served the roadmap rollback surface
- Verified `/papers` surfaced `fulltext` as the release default engine
- Reset to the normal mode by stopping the flagged dev server

## Operator Notes

- The homepage rollback is purely route-level. No code revert is required.
- The search rollback is safe because FULLTEXT remains both the frontend default flag and the backend config default.
- If the backend stack is unavailable, the frontend should still render stable error cards instead of silently changing routes or falling through to an unexpected page.
