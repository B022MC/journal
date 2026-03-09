---
description: 添加新的自治模块功能（评分、贡献、降解、举报相关）
---

# 自治模块开发

## 前置检查

⚠️ **在开始之前必须确认**：
- 新功能是否涉及治理决策？
- 如果是，是否使用**透明公式 + 人类行为驱动**？
- 是否引入了任何 AI/ML 黑箱（BERT、Word2Vec、GPT）？→ **必须拒绝**

## 步骤

1. 确认功能属于哪个模块：
   - **贡献度** → `common/contribution/calculator.go`
   - **降解/举报** → `common/degradation/engine.go` + `model/flagmodel.go`
   - **限流** → `common/ratelimit/limiter.go`
   - **Score 计算** → `model/papermodel.go` 中的 `CalcShitScoreV2`
   - **定时任务** → `cmd/cron/main.go`

2. 如需数据库变更 → 执行 `/db-migration` 工作流

3. 实现业务逻辑：
   - 所有公式必须在代码注释中写明数学表达式
   - 阈值必须定义为常量或可配置变量，不能硬编码在逻辑中
   - 评分/贡献相关计算必须使用 `float64`，保留至少 2 位小数

4. 如果新功能需要定时执行：
   - 在 `cmd/cron/main.go` 的统一调度器中注册
   - 选择合适的执行频率（高频操作实时触发，低频操作放 cron）

5. 编译验证
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && go build ./...
```

6. 更新 ROADMAP.md 中对应条目状态
