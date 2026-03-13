---
mode: plan
cwd: /Users/b022mc/project/shit/journal
task: 基于 Report.md 聚焦搜索 Batch 1 与切流证据链的执行计划
complexity: complex
planning_method: builtin
created_at: 2026-03-13T13:17:33+08:00
---

# Plan: 搜索 Batch 1 与切流计划

🎯 任务概述
`Report.md` 已把当前优先级从“前端主站建设”收敛到“搜索重构 Batch 1 + 切流证据链”。本计划替代 2026-03-12 那份覆盖前端、搜索、平台的宽计划，聚焦 `paper-rpc` 搜索引擎的落地、验证、观测与回滚，不扩散到 P1/P2 体验增强。

📋 执行计划
1. 冻结本轮范围和验收门槛：以 `Report.md` 的 P0 三项为唯一主线，显式声明 `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` 与 `Search.DefaultEngine=fulltext` 在全部门槛通过前不得变更；同步列出 Milestone D 的 owner、证据清单和“不做项”（SSR/SEO、UI 体验增强、治理功能），防止范围回弹。
2. 盘点现状并形成 ADR 差异清单：审计 `paper-rpc` 现有的配置脚手架、hybrid/fulltext 路由、explain、shadow compare、`make search-bench` 能力，与 ADR 要求的 `active_index_version`、版本化 segment、结构化 metrics/logs、golden benchmark 数据集逐项对照；输出 implementation checklist，作为后续 PR 拆分依据。
3. 补齐 Batch 1 的索引生命周期而不是继续堆功能：在现有内存 `Snapshot` 基础上补出“构建 -> 校验 -> 原子发布 -> 加载 -> 回退”的索引流程，引入 index version、checksum、segment 元数据和最近可用版本缓存；要求失败时直接降级到 FULLTEXT，并保留最近成功构建产物用于离线排障。
4. 强化检索编排与 fallback 语义：把 query parsing、retrieval orchestration、ranker、explain renderer 的边界进一步固定下来，统一 fallback reason 枚举和超时、缺索引、加载失败处理，确保 FULLTEXT 回退既能保住请求成功，也能给 benchmark 和日志留下可比较证据。
5. 补齐搜索观测面：按 ADR 增加按 `engine`、`mode`、`result` 维度区分的 Prometheus metrics，并把 request id、engine mode、index version、normalized query、token list、top result ids、fallback reason 写入统一日志；shadow compare 还要输出 delta 分类，方便判断是召回差异、排序差异还是污染问题。
6. 建立固定 benchmark 数据集与自动评估：把当前合成 benchmark 扩展为可复用的 golden set/fixture，覆盖中文精确匹配、中英混合、discipline filter、prefix query，以及未来 synonym case 的占位样本；同时把 Recall@10、p50/p95 latency、explain completeness、重复构建稳定性纳入可重复执行的 bench/test target。
7. 衔接前端控制台与发布开关，但不提前切默认：继续保持 `/papers` 页面上的 `engine` / `shadow_compare` 只是验证入口，默认发布仍由 `JOURNAL_SEARCH_RELEASE_ENGINE` 和 `paper.yaml` 控制；在 shadow compare 期间优先保证 API contract 和页面提示稳定，而不是先做联想高亮等 UI 增强。
8. 以里程碑方式推进并保留回滚演练：Milestone D1 为“版本化索引 + fallback 完成”，D2 为“metrics/logs/benchmark 齐备”，D3 为“shadow compare 连续通过且无污染”，D4 为“具备切到 hybrid 的候选资格”；每个里程碑都要求跑后端测试、`make search-bench`、前端 smoke，以及一次 FULLTEXT 回滚 drill。
9. 明确依赖协调和后续队列：search maintainer 负责 `paper-rpc` 与 benchmark 证据，frontend maintainer 负责 `/papers` 的开关呈现与回滚提示，release captain 负责 runbook 执行；Trie、synonym fusion、搜索高亮、治理可视化继续留在 Batch 2 / P1，不允许阻塞 Batch 1 合入。

## Batch 1 Scope Freeze

This snapshot is the only delivery boundary for the current cutover work.

### In scope

- search rearchitecture batch-one work in `paper-rpc`
- fallback and rollback evidence that keeps FULLTEXT as the safe path
- observability, benchmark, and shadow-compare proof needed for Milestone D
- `/papers` validation entry and release-default messaging only

### Release invariants

- `JOURNAL_SEARCH_RELEASE_ENGINE=fulltext` stays unchanged until every Milestone D gate is signed off
- `backend/rpc/paper/etc/paper.yaml` keeps `Search.DefaultEngine: fulltext` until the frontend route-level cutover has already proven stable
- any new scope that changes these defaults or expands beyond this snapshot requires a new plan or issues CSV refresh before code changes

### Explicitly out of scope

- SSR or SEO work
- UI enhancement work such as suggestion, highlight, or broader browse-mode redesign
- governance or visualization features
- Batch 2 search capabilities: Trie, synonym expansion, and fusion ranking
- irreversible schema or storage changes bundled into the same release step as a search cutover

⚠️ 风险与注意事项
- ADR 文档把当前状态描述得比代码更早期，但仓库里已经有 `hybrid`、explain、shadow compare 和合成 benchmark；如果不先做差异盘点，容易重复开发或误判缺口。
- 当前索引仍是进程内 `Snapshot`，缺少持久化 segment/version publication，重启一致性、回滚证据和多次重建稳定性都不足。
- `search-bench` 目前主要基于合成数据，若不补真实或固定 golden set，就无法作为切流门槛证据。
- `paper.yaml` 已默认开启部分 Batch 2 开关，若不在计划里明确“Batch 2 不阻塞 Batch 1”，验收口径会继续漂移。

📎 参考
- `../Report.md:24`
- `../Report.md:43`
- `../Report.md:58`
- `backend/docs/adr/2026-03-12-search-rearchitecture.md:79`
- `backend/docs/adr/2026-03-12-search-rearchitecture.md:98`
- `backend/docs/adr/2026-03-12-search-rearchitecture.md:134`
- `backend/docs/release/2026-03-12-main-site-rollout.md:30`
- `backend/docs/release/2026-03-12-main-site-rollout.md:59`
- `backend/rpc/paper/etc/paper.yaml:21`
- `backend/rpc/paper/internal/search/service.go:56`
- `backend/rpc/paper/internal/search/index.go:103`
- `backend/model/papermodel.go:264`
- `backend/Makefile:105`
- `frontend/src/app/papers/page.tsx:19`
