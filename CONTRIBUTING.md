# S.H.I.T Journal 开发指南

> **S**cholarly **H**ub for **I**ntellectual **T**rash — 学术论文 "屎" 评系统

---

## 设计哲学

> [!IMPORTANT]
> 本项目是一项**社会实验**：探讨在无行政干预、无传统同行评审（Peer Review）的环境下，学术平权与内容质量自治的可能性。

### 核心约束
1. **人类社区自治** — 所有治理（评分、晋升、降解、角色）由人类行为驱动 + 透明公式量化
2. **拒绝 AI 黑箱** — 不接受以 AI 为核心的治理方案（BERT、GPT、Word2Vec 等均禁止用于治理决策）
3. **不删除，只降解** — 问题内容通过降低可见度自然"沉淀"，不直接删除
4. **角色可逆** — 即使 admin 贡献分衰减也会降级为普通用户，无永久特权

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 框架 | go-zero v1.10（微服务） |
| 语言 | Go 1.25 |
| 通信 | gRPC + Protobuf |
| 数据库 | MySQL 8.0（1 Master + 2 Replica） |
| 缓存 | Redis 7 |
| 注册中心 | Etcd 3.5 |
| 代码生成 | goctl v1.9+ |

---

## 快速上手

```bash
# 1. 推荐：一键启动全部基础设施与服务
cd backend && ./start.sh dev

# 2. 生产模式启动
./start.sh prod
```

如果你需要手动调试单个进程：

```bash
make env-up
make init-repl
make tidy
make user-rpc      # :9001
make paper-rpc     # :9002
make rating-rpc    # :9003
make news-rpc      # :9004
make admin-rpc     # :9005
make api           # :8888
make admin-api     # :8889
make cron          # 统一定时任务
```

---

## 项目结构

```
backend/
├── api/                    # HTTP API 网关
│   ├── journal.api     # API DSL 定义
│   └── internal/
│       ├── handler/        # ⚠️ 自动生成，勿手动改
│       ├── logic/          # ✅ 在此编写业务逻辑
│       ├── svc/            # ServiceContext（依赖注入）
│       └── types/          # ⚠️ 自动生成
├── admin-api/              # 管理端 HTTP API
├── rpc/{user,paper,rating,news,admin}/  # gRPC 微服务
├── proto/                  # Protobuf 定义
├── model/                  # 数据访问层（DAO）
│   ├── schema.sql          # 完整 DDL
│   └── migrations/         # 增量迁移脚本
├── common/                 # 公共包
│   ├── contribution/       # 贡献度计算器（Shapley 值启发）
│   ├── degradation/        # 降解引擎（Quorum 法定人数）
│   ├── ratelimit/          # Redis Token Bucket 限流
│   ├── errorx/             # 双语错误码
│   ├── jwt/                # JWT 工具
│   └── result/             # 统一响应
├── cmd/cron/               # 统一定时任务调度
├── start.sh                # 一键启动脚本
└── deploy/mysql/           # MySQL 主从配置
```

### 三层架构（强制执行）

```
HTTP Request
    ↓
  Handler    ← 参数解析 + 响应封装（自动生成，勿改）
    ↓
  Logic      ← 业务逻辑（在此编写核心代码）
    ↓
  Model      ← 数据访问（SQL 查询）
```

---

## 代码生成工作流

### 添加新的 API 接口

```bash
# 1. 编辑 API DSL
vim api/journal.api

# 2. 重新生成代码
make api-gen

# 3. 在 logic/ 中实现业务
vim api/internal/logic/your_new_logic.go

# 4. 编译验证
go build ./...
```

### 添加新的 RPC 服务

```bash
# 1. 编辑 Proto 定义
vim proto/your_service.proto

# 2. 生成 RPC 代码
make proto-xxx   # 或 make proto-all

# 3. 在 logic/ 中实现业务
vim rpc/xxx/internal/logic/your_logic.go

# 4. 编译验证
go build ./...
```

---

## 编码规范

### Go

```go
// ✅ 正确：Context 第一参数，使用 errorx
func (l *SubmitPaperLogic) SubmitPaper(ctx context.Context, req *types.SubmitPaperReq) (*types.SubmitPaperResp, error) {
    if req.Title == "" {
        return nil, errorx.ErrInvalidParam
    }
    // 业务逻辑...
}

// ❌ 错误：直接写 SQL、使用 fmt.Errorf
func (l *SubmitPaperLogic) SubmitPaper(req *types.SubmitPaperReq) error {
    db.Exec("INSERT INTO paper ...")  // 应走 Model 层
    return fmt.Errorf("something wrong")  // 应用 errorx
}
```

### Model 层

```go
// 写操作标注主库
// === 写操作 → 主库 ===
func (m *PaperModel) Insert(ctx context.Context, p *Paper) (int64, error) { ... }

// 读操作自动走从库
// === 读操作 → 从库 (SELECT 自动路由) ===
func (m *PaperModel) FindById(ctx context.Context, id int64) (*Paper, error) { ... }

// 写后读走主库
func (m *PaperModel) FindByIdPrimary(ctx context.Context, id int64) (*Paper, error) {
    ctx = sqlx.WithReadPrimary(ctx)  // 强制走主库
    // ...
}
```

### 数据库

- 表名：单数小写（`user`、`paper`、`rating`）
- 主键：`id BIGINT UNSIGNED AUTO_INCREMENT`
- 字符集：`utf8mb4_unicode_ci`，引擎：`InnoDB`
- 逻辑删除：`status TINYINT`（0=删除, 1=正常）
- 新增字段/表：先写 `model/migrations/NNN_xxx.sql`，再同步 `schema.sql`

---

## 自治模块核心公式

### S.H.I.T Score v2

```
Score = 0.30 × weighted_avg
      + 0.20 × log(weighted_rating_count + 1)
      + 0.10 × log(view_count + 1)
      - 0.15 × controversy
      + 0.15 × reviewer_authority
      + 0.10 × freshness
```

> 所有维度均有明确数学定义，无黑箱成分。

### 贡献度

```
ContributionScore = 0.40×AuthorScore + 0.35×ReviewerScore + 0.15×ActivityScore - 0.10×DecayPenalty
```

- **AuthorScore** = Σ paper.shit_score × zone_weight(paper.zone)
- **ReviewerScore** = Σ review_accuracy × depth_bonus
- **ActivityScore** = log(reviews_30d + 1) × log(logins_30d + 1)
- **DecayPenalty** = max(0, days_inactive - 30) × 0.01

### 评审者权重

```
reviewer_weight = clamp(0.10 + contribution_score / 200, 0.10, 1.00)
```

- `reviewer_weight` 在评分写入时实时计算
- 用于加权平均分 `weighted_avg_rating` 与评审者权威度 `reviewer_authority`
- 该规则保持完全可解释，避免黑箱加权

### 降解协议

```
举报数 ≥ 2 或 加权分 ≥ 10  → Level 1（观察：标注⚠️）
举报 ≥ quorum 或 加权分 ≥ 50  → Level 2（限流：列表隐藏）
举报 ≥ 2×quorum 或 加权分 ≥ 100 → Level 3（封存：仅作者可见）
其中 quorum = max(3, √rating_count)
```

### 关键词黑名单

- 规则表：`journal_biz.keyword_rule`
- 热更新实现：`common/degradation/keyword_filter.go`
- 缓存层级：MySQL 主库 -> Redis -> 进程内编译缓存
- 支持规则类型：`keyword`、`regex`、`pinyin`
- `pinyin` 规则只接受 ASCII 拼音模式，禁止直接写中文原文
- 当前接入路径：`paper-rpc` 投稿、`news-rpc` 发新闻
- 管理要求：`admin.keyword.manage` 权限 + `contribution_score >= 200`

### SimHash 查重

- 指纹计算：`common/degradation/simhash.go`
- 落库字段：`paper.simhash`
- 当前接入路径：`paper-rpc` 投稿
- 检测规则：与已有论文做 Hamming Distance 比对，`<= 3` 视为疑似重复
- 处置方式：新论文写入 `status=flagged`，并自动生成一条系统 `plagiarism` 举报记录

### 热数据缓存

- 热门论文缓存键：`api:papers:hot:{zone|all}`，TTL `5 min`
- 用户贡献分缓存键：`api:user:contribution:{userId}`，TTL `1h`
- 举报状态缓存键：`api:flags:status:{targetType}:{targetId}`，TTL `10 min`
- 论文降级快照缓存键：`api:paper:moderation:{paperId}`，TTL `30 min`
- 当前失效路径：投稿、评分、论文举报

### 评分事件队列

- 队列键：`events:rating:postrate:v1`
- 死信键：`events:rating:postrate:v1:dead`
- 当前生产者：`rating-rpc` 评分写库成功后
- 当前消费者：`rating-rpc` 后台 worker（Redis `BLPOP`）
- 当前异步职责：
  - 刷新论文 `avg_rating / rating_count / controversy / weighted_score`
  - 刷新评分者与作者 `contribution_score / role`
  - 执行 `burst / bimodality` 恶意评分检测
  - 失效评分相关缓存：热门论文、贡献分、`paper flag-status`、`paper moderation`
- 当前兜底策略：Redis 不可用或入队失败时，立即回退到同步处理，保证后处理不丢

### 举报事件队列

- 队列键：`events:flag:postsubmit:v1`
- 死信键：`events:flag:postsubmit:v1:dead`
- 当前生产者：`api` 举报写库成功后
- 当前消费者：`api` 后台 worker（Redis `BLPOP`）
- 当前异步职责：
  - 执行 `quorum / degradation level` 计算
  - 对论文目标应用降级并结案 pending flags
  - 失效举报相关缓存：`paper flag-status`、`paper moderation`、热门论文缓存
- 当前兜底策略：Redis 不可用或入队失败时，立即回退到同步处理，保证后处理不丢

### 成就徽章

- 落库表：`user_achievement`
- 当前徽章规则：
  - `first_submission`：至少投稿 `1` 篇
  - `sediment_breakthrough`：至少有 `1` 篇论文进入 `sediment`
  - `reviewer_century`：累计评分 `>= 100`
- 当前同步路径：
  - `paper-rpc` 投稿成功后同步作者成就
  - `paper-rpc` 升区到 `sediment` 后同步作者成就
  - `rating-rpc` 评分后处理异步同步审稿人成就
  - `api` 登录 / `user/info` 读取前做一次兜底同步
- 当前返回位置：`/login`、`/user/info` 的 `user_info.achievements`

### 恶意评分实时检测

- 实现位置：`common/degradation/rating_guard.go`
- 当前接入路径：
  - `rating-rpc` 评分提交后：burst / bimodality
  - `api` 评分成功回写后：IP / 设备指纹聚集检测
- Burst 规则：用户 `10` 分钟内 `>= 8` 次评分，自动生成 `user` 目标的系统 `manipulation` 举报
- 双峰规则：论文评分分布 `bimodality coefficient > 0.55`，自动生成 `paper` 目标的系统 `manipulation` 举报
- IP 聚集规则：同一论文在 `24h` 内出现同 IP `>= 3` 个不同用户评分，自动生成 `paper` 目标的系统 `manipulation` 举报
- 设备指纹规则：同一论文在 `24h` 内出现同 `device_fingerprint` `>= 2` 个不同用户评分，自动生成 `paper` 目标的系统 `manipulation` 举报
- `device_fingerprint = sha256(source_ip + "|" + normalized_user_agent)`

---

## 错误码

| 范围 | 领域 | 示例 |
|------|------|------|
| `0` | 成功 | — |
| `10001~10006` | 通用 | ErrInternal, ErrInvalidParam |
| `20001~20005` | 用户 | ErrUserNotFound, ErrPasswordWrong |
| `30001~30003` | 论文 | ErrPaperNotFound, ErrInvalidZone |
| `40001~40003` | 评分 | ErrAlreadyRated, ErrSelfRating |
| `50001~50004` | 举报 | ErrAlreadyFlagged, ErrSelfFlag |

---

## 定时任务

所有定时任务统一在 `cmd/cron/main.go` 中管理：

| 频率 | 任务 | 说明 |
|------|------|------|
| 每小时 | Zone 生命周期 | 多维度晋升/降级 |
| 每天 03:00 | 贡献度衰减 | 不活跃 >30 天用户扣 5% |
| 每天 03:05 | 角色审计 | 按贡献分调整角色 |
| 每天 04:00 | 降解扫描 | 处理待审举报 |

---

## Git 约定

### 分支策略
- `main` — 稳定分支
- `feat/xxx` — 功能开发
- `fix/xxx` — Bug 修复
- `refactor/xxx` — 重构

### Commit Message 格式
```
<type>(<scope>): <description>

feat(paper): add SimHash duplicate detection
fix(rating): prevent self-rating bypass
refactor(model): extract contribution score to common package
docs(roadmap): add governance module section
```

类型：`feat` / `fix` / `refactor` / `docs` / `chore` / `test`

---

## AI 辅助开发

本项目已配置 go-zero 官方 AI 工具链：

| 工具 | 用途 | 路径 |
|------|------|------|
| **ai-context** | go-zero 通用 AI 规则 | `.github/ai-context/` |
| **mcp-zero** | MCP 工具（代码生成/验证） | `/Users/b022mc/go/bin/mcp-zero` |
| **项目规则** | ShitJournal 专属约束 | `.cursor/rules/journal.mdc` |

> **注意**：AI 助手仅用于辅助**开发者写代码**，不参与平台的**治理决策**。这是项目的核心设计原则。

---

> 📖 技术路线图请参阅 [ROADMAP.md](ROADMAP.md)  
> 📖 项目详情请参阅 [README.md](README.md)
