---
description: ShitJournal 项目约束与架构概览，所有开发任务必须遵循
---

# 项目核心约束

## 🚫 核心禁令
1. **拒绝 AI 治理黑箱**：治理逻辑（评分、晋升、降解、角色）只允许透明公式 + 人类行为驱动。禁止 BERT/GPT/Word2Vec 等用于治理。
2. **三层架构**：Handler → Logic → Model，不得跳层。Handler/Types 自动生成勿改。
3. **搜索/分词**也必须用规则可解释方案（IK、BM25、编辑距离），不引入神经网络。

## 架构速查

```
backend/
├── api/                    HTTP 网关 :8888（shitjournal.api DSL）
├── rpc/                    gRPC: user:9001, paper:9002, rating:9003, news:9004
├── proto/                  .proto 定义
├── model/                  DAO（usermodel/papermodel/ratingmodel/newsmodel/flagmodel）
├── common/                 jwt / errorx / result / contribution / degradation / ratelimit
└── cmd/cron/               统一定时任务
```

## 编码规范
- Context 永远第一参数
- 错误用 `errorx.NewCodeError`，不用 `fmt.Errorf`
- 写操作注释 `// === 写操作 → 主库 ===`
- 写后读用 `sqlx.WithReadPrimary(ctx)`
- 数据库：表名单数小写，主键 BIGINT UNSIGNED，utf8mb4，InnoDB

## 自治模块公式
- **Score v2** = 0.30×weighted_avg + 0.20×log(count+1) + 0.10×log(views+1) - 0.15×controversy + 0.15×reviewer_authority + 0.10×freshness
- **贡献度** = 0.40×Author + 0.35×Reviewer + 0.15×Activity - 0.10×Decay
- **降解** = quorum=max(3,√rating_count)，4 级降解（正常→观察→限流→封存）
- **角色** = 0~10 member / 10~50 scooper / 50~200 editor / 200+ admin（可降级）

## 错误码范围
- 10001~10006 通用 | 20001~20005 用户 | 30001~30003 论文 | 40001~40003 评分 | 50001~50004 举报
