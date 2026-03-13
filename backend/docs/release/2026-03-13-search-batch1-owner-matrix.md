# Search Batch 1 Owner Matrix And Batch 2 Queue

Date: 2026-03-13

Purpose: keep Batch 1 cutover ownership, handoff points, and non-blocking Batch
2 work in one place so the release review can distinguish required evidence from
future enhancements.

## Owner Matrix

| Role | Owned deliverables | Handoff trigger | Handoff artifact |
| --- | --- | --- | --- |
| search maintainer | `paper-rpc` lifecycle, fallback semantics, metrics or logs, golden benchmark summaries, and shadow-compare evidence | D1 or D2 evidence is green, or D3 shadow evidence is ready for review | updated evidence ledger, benchmark summary, fallback or shadow-compare notes, and current blocking list |
| frontend maintainer | `/papers` validation entry, release-default engine messaging, retry or rollback copy, and smoke coverage for `engine=auto` plus explicit overrides | the route reflects the current release flag and smoke stays green under both validation and rollback scenarios | smoke result, route-level notes, and any UX copy that release captain should validate |
| release captain | runbook updates, milestone sign-off, release-flag changes, and rollback drill record | search maintainer and frontend maintainer have both handed off green evidence for the current stage | signed runbook step, rollback drill record, and go or no-go decision for the next milestone |

## Current Batch 1 Blocking List

Only the items below may block Batch 1 cutover review:

- D3 recurring shadow-compare evidence archive is still missing, so the default
  route must stay on `fulltext`
- D4 release-captain sign-off cannot begin until D3 remains green over repeated
  checks

Everything else in the Batch 2 queue below is explicitly non-blocking for Batch
1.

## Batch 2 Queue

| Item | Owner when scheduled | Why it is non-blocking for Batch 1 | Re-entry condition |
| --- | --- | --- | --- |
| Trie suggestions beyond the current validation surface | search maintainer plus frontend maintainer | Batch 1 only needs `/papers` validation controls and stable rollback messaging | start after D4 or a new issues CSV refresh |
| Synonym fusion tuning and broader ranking enhancements | search maintainer | the current cutover gates only require benchmark parity, fallback safety, and shadow evidence | start after D4 with a fresh benchmark plan |
| Search highlighting or richer result presentation | frontend maintainer | it is a UI enhancement and does not affect the answering path or rollback safety | schedule as a separate P1 or Batch 2 issue |
| Governance or visualization features | release captain plus frontend maintainer | these features do not change query routing, fallback, or cutover evidence | schedule in a separate snapshot after Batch 1 |

## Review Rule

- Reject any PR or issue that tries to move a Batch 2 row into the Batch 1
  blocker list without first updating the plan or the issues CSV snapshot.
- If a handoff artifact is missing, the owning role remains responsible for the
  stage; do not assume another role will infer or recreate the evidence.
