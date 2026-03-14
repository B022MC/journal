# Remote Journal Milestone Tracker

Date: 2026-03-14

Purpose: close `JRV-090` by keeping one milestone-oriented evidence index for
the remote single-db validation work.

## Milestone Board

| Milestone | Current status | Evidence | Primary command | Rollback or next step |
| --- | --- | --- | --- | --- |
| `M1` remote config landed and startup path ready | Ready for live env, not yet proven against a real remote DSN | `ce18986`, `e736a88`, `7b6652f`, `8d2da13`; `2026-03-14-remote-journal-validation-preflight.md`; `2026-03-14-remote-journal-entrypoint-matrix.md`; `2026-03-14-remote-journal-config-overlay.md`; `2026-03-14-remote-journal-startup-path.md` | `cd backend && REMOTE_JOURNAL_DSN='<redacted>' ./start.sh dev remote` | If startup fails, use `2026-03-14-remote-journal-go-no-go.md` rollback steps and return to `./start.sh dev` |
| `M2` frontend and admin minimum regression passed | Complete with residual RBAC inventory drift noted | `e4cfd16`; `cf25d24`; `9504bdf`; `2026-03-14-remote-journal-startup-path.md`; `2026-03-14-remote-journal-frontend-regression.md`; `2026-03-14-remote-journal-admin-regression.md` | `cd backend && SKIP_INFRA=1 scripts/with_remote_journal_validation_env.sh ./start.sh dev remote`, then replay the startup, frontend, and admin evidence docs | Keep the residual permission inventory drift visible in `2026-03-14-remote-journal-admin-regression.md`; `NRJ-080` can start from the batch-1 runtime model plan |
| `M3` legacy reference inventory frozen | Complete in repo | `592ce1f`; `2026-03-13-db-merge-legacy-baseline.csv`; `2026-03-13-db-merge-legacy-inventory.md`; `2026-03-14-legacy-db-ref-batch1-plan.md` | `cd backend && python3 scripts/check_legacy_db_refs.py` | Update the baseline only in the same commit as future `biz_*` rewrites |
| `M4` first `biz_*` convergence batch complete | Planned, not started | `592ce1f`; `2026-03-14-legacy-db-ref-batch1-plan.md` defines the target files and rewrite order | Batch-1 runtime order: `PaperModel/UserModel -> RatingModel -> FlagModel -> KeywordRuleModel -> NewsModel` | Wait until `M2` is green so compatibility views stop carrying unverified runtime traffic |

## Commit Index

- `ce18986` `docs: [JRV-000] 冻结远端单库验证边界`
- `e736a88` `docs: [JRV-010] 盘点单库切换涉及的全部服务入口`
- `7b6652f` `feat: [JRV-020] 落远端单库配置覆盖层`
- `8d2da13` `feat: [JRV-030] 提供跳过本地 MySQL 主从的启动路径`
- `5224779` `docs: [JRV-070] 固化 go/no-go 与回滚动作`
- `592ce1f` `docs: [JRV-080] 固化旧表名 inventory 并规划首批 biz 收敛`
- `9504bdf` `fix: [NRJ-050] close remote admin regression`

## Current Blockers

- `JRV-080`: batch-1 `biz_*` convergence is not started yet; use the now-green
  `M2` evidence set as the regression baseline for each follow-up batch.
- `JRV-090`: remains blocked on actual `JRV-080` runtime model rewrites and the
  post-change naming-freeze reruns.

## Operating Rule

- Update this tracker whenever a milestone status changes.
- Every milestone row must point to a concrete command, evidence document, or
  commit instead of a free-form conclusion.
