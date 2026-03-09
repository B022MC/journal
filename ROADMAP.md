# S.H.I.T Journal 技术路线图

> 本文档梳理 ShitJournal 后端的中长期演进路径，分为 **架构** 和 **功能** 两个维度。  
> 优先级标记：🔴 高优 · 🟡 中优 · 🟢 低优  

> [!IMPORTANT]
> **设计原则：人类社区自治，拒绝 AI 黑箱**  
> 本项目是一项社会实验，探讨在无行政干预、无传统同行评审的环境下，学术平权与内容质量自治的可能性。  
> 所有治理逻辑（评分、晋升、降解、角色分配）均由 **人类行为驱动 + 透明公式量化**，不接受以 AI 为核心的治理方案。  
> 搜索与分词等工具性模块亦应选用 **规则可解释** 的方案（如 BM25、IK 分词、编辑距离），避免引入神经网络黑箱。

---

## 零、已完成 ✅

### 0.1 项目基础设施
- [x] Monorepo 项目结构（`backend/` + `frontend/`）
- [x] go-zero 微服务骨架：`goctl` 生成 4 个 RPC + 1 个 API 网关
- [x] Docker Compose：MySQL 8.0（1 Master + 2 Replica）+ Redis 7 + Etcd 3.5
- [x] MySQL GTID 主从复制配置（`deploy/mysql/` 下 cnf + init 脚本）
- [x] PostgreSQL DDL → MySQL DDL 迁移（`model/schema.sql`）
- [x] Makefile 全套 target（build / run / proto-gen / env-up / init-repl）

### 0.2 后端核心服务
- [x] **User Service** `:9001` — 注册（bcrypt 哈希）/ 登录（JWT 签发）/ 获取信息 / 更新资料
- [x] **Paper Service** `:9002` — 投稿 / 列表（Zone+学科+排序）/ 详情（异步浏览量+1）/ FULLTEXT 搜索 / Zone 更新 / 用户论文
- [x] **Rating Service** `:9003` — 评分 Upsert / 防自评 / S.H.I.T Score 实时重算（AVG + STDDEV→争议度）/ 按论文或用户查评分
- [x] **News Service** `:9004` — 新闻 CRUD / 置顶排序

### 0.3 数据层
- [x] go-zero `sqlx.SqlConn` Model 层：4 个 model 文件，使用 `QueryRowCtx` / `QueryRowsCtx` / `ExecCtx`
- [x] **MySQL 读写分离**：`sqlx.SqlConf` + `Replicas` + `Policy: round-robin`
- [x] 注册唯一性检查走主库 `sqlx.WithReadPrimary(ctx)` 避免主从延迟
- [x] S.H.I.T Score 算法：`0.4×norm_avg + 0.25×log(count) + 0.15×log(views) - 0.2×controversy`
- [x] MySQL FULLTEXT 搜索：`MATCH(title, abstract, keywords) AGAINST(? IN BOOLEAN MODE)`
- [x] Rating Upsert：`ON DUPLICATE KEY UPDATE`

### 0.4 API 网关
- [x] 16 个 API Logic 文件（HTTP → gRPC 转发）
- [x] 17 个 Handler 文件（含 routes.go）
- [x] JWT 鉴权分组：公开路由 / 认证路由 / 管理员路由
- [x] API 网关 `ServiceContext` 连接 4 个 RPC Client

### 0.5 公共包 & 工具
- [x] `common/jwt` — JWT 生成 / 解析
- [x] `common/errorx` — 双语错误码
- [x] `common/result` — 统一 JSON 响应封装
- [x] `cmd/lifecycle/main.go` — 论文生命周期定时任务（四区自动晋升/降级，每小时执行）

### 0.6 社区自治模块
- [x] **S.H.I.T Score v2 算法** — 六维度复合评分：加权平均 + 加权评分数 + 浏览量 + 争议度 + 评审者权威度 + 新鲜度
- [x] **多维度晋升/降级阈值** — 最低评分数 + 最低 Score + 最低加权评分数 + 最低评审者权威 + 最短存在时间 + 举报次数
- [x] **贡献度计算器** `common/contribution/calculator.go` — Shapley 值启发：AuthorScore(0.40) + ReviewerScore(0.35) + ActivityScore(0.15) - DecayPenalty(0.10)
- [x] **自动角色分配** — 基于 `contribution_score` 自动赋予 member/scooper/editor/admin，角色可因衰减降级
- [x] **降解引擎** `common/degradation/engine.go` — Quorum 法定人数机制：举报 → 观察 → 限流 → 封存 四级降解
- [x] **举报系统** `model/flagmodel.go` — Flag 表 + 目标类型举报 + 加权举报统计 + 批量处置
- [x] **Redis Token Bucket 限流器** `common/ratelimit/limiter.go` — 评分 20/h、举报 10/d、搜索 10/min
- [x] **统一定时任务引擎** `cmd/cron/main.go` — 合并 lifecycle/contribution-decay/role-audit/degradation-sweep 四类定时任务
- [x] **数据库迁移** `model/migrations/002_governance_modules.sql` — flag 表 + paper/user/rating 治理字段

---

## 一、架构演进

### 1.1 可观测性三件套

- [x] 🔴 引入 **Jaeger** 全链路追踪
  - go-zero 原生支持 OpenTelemetry，配置 `Telemetry` 即可上报
  - 跨服务追踪：API Gateway → RPC Services，通过 `traceparent` header 透传 trace context
  - Docker Compose 新增 `jaeger-all-in-one` 容器，UI 端口 16686
- [x] 🔴 引入 **Prometheus** + **Grafana** 监控
  - go-zero 内置 `/metrics` 端点，开启 `Prometheus: true` 即可暴露
  - 自定义 metrics：QPS、P99 延迟、S.H.I.T Score 分布、论文各 Zone 数量
  - Grafana Dashboard 模板：RPC 服务健康、MySQL 主从延迟、Redis 命中率
- [x] 🟡 引入 **降级熔断**
  - go-zero 自带 `breaker` 和 `shedding`（自适应降载），已在 RPC client 配置中启用
  - 对 Paper/Rating 等高频读服务配置熔断阈值
  - 降级策略：搜索服务不可用时 fallback 到 MySQL FULLTEXT；评分服务不可用时返回缓存快照

### 1.2 数据层抽象

- [x] 🔴 **抽离 DAO Init，按 Key 获取数据库实例**
  - 当前每个 RPC 的 `ServiceContext` 各自 `sqlx.MustNewConn()`，耦合配置
  - 改造方案：创建 `common/dao` 包，注册多个数据源 (`master`/`readonly`/`analytics`)
  - 使用 key 获取实例：`dao.GetConn("paper-master")`、`dao.GetConn("paper-readonly")`
  - 支持按业务维度隔离：用户库 / 论文库 / 日志库

```go
// common/dao/dao.go
var connMap = sync.Map{}

func Register(key string, conf sqlx.SqlConf) {
    connMap.Store(key, sqlx.MustNewConn(conf))
}

func GetConn(key string) sqlx.SqlConn {
    v, _ := connMap.Load(key)
    return v.(sqlx.SqlConn)
}
```

### 1.3 冷热数据分离

- [x] 🟡 **定义冷热标准**
  - 热数据：最近 N 天有访问（浏览/评分/搜索命中），Zone 为 `latrine`/`septic_tank`
  - 冷数据：超过 90 天无访问 且 Zone 为 `sediment`，或 `status = 0`（已删除）
  - 标记方式：`paper` 表新增 `last_accessed_at` 字段，GetPaper 时异步更新
- [x] 🟡 **分离存储**
  - 方案 B（定时迁移）：cron job 每日 5:00 AM 扫描冷数据 `INSERT INTO cold_paper SELECT ...`，源表软删除
  - 冷库：`cold_paper` 表（InnoDB，同结构 + `archived_at`）
- [ ] 🟢 参考 Elasticsearch 的 ILM（Index Lifecycle Management）策略
  - Hot → Warm → Cold → Frozen → Delete 五段式生命周期
  - 类比：Latrine(hot) → SepticTank(warm) → Stone(cold) → Sediment(frozen)

### 1.4 存储演进路径

- [ ] 🟢 **MySQL → StarRocks（OLAP）**
  - 当前 MySQL 8.0 + FULLTEXT 足够支撑初期（百万级）
  - 中期：论文量达千万级后，正排索引查询瓶颈，MySQL 需分库分表
  - 终态：StarRocks 单表亿级毫秒查询，支持实时 OLAP 分析
  - 迁移路径：MySQL binlog → Flink CDC → StarRocks，双写过渡期
  - StarRocks 适合的场景：论文热度排行榜、学科分布统计、S.H.I.T Score 聚合分析

---

## 二、功能迭代

### 2.1 搜索引擎核心

- [ ] 🔴 **索引构建并发化**
  - 当前单线程构建索引太慢
  - 引入 goroutine pool（`ants`），按 discipline 分片并发构建
  - 传文件地址改为传文件流（`io.Reader`），减少磁盘 IO

- [ ] 🔴 **倒排索引优化**
  - 当前方案：MySQL FULLTEXT（基于块的索引）
  - 改造目标：自建 inverted index，存 offset 而非全文
  - 使用 `mmap` 映射索引文件，减少内存拷贝
  - 索引压缩：Variable Byte Encoding / PForDelta

- [ ] 🟡 **Roaring Bitmap 存储 DocID**
  - 替换传统 posting list，大幅压缩稀疏 ID 集合
  - 支持高效交并差运算（位运算）
  - Go 实现：`github.com/RoaringBitmap/roaring`

- [ ] 🟡 **分布式 MapReduce 构建索引**
  - 参考 MIT 6.824 Lab1，实现分布式索引构建
  - 计算存储分离：索引构建节点（无状态）+ 索引存储节点
  - 在此基础上支持动态索引（增量更新而非全量重建）

### 2.2 相关性排序

- [ ] 🔴 **TF-IDF / BM25 相关性计算**
  - 实现 `TF` 类，计算词频和逆文档频率
  - BM25 参数调优：`k1=1.2`, `b=0.75`
  - 学术论文场景：标题权重 > 关键词 > 摘要 > 正文

- [ ] 🟡 **排序器优化**
  - 多路召回后统一排序：`BM25_score * 0.6 + shit_score * 0.3 + recency * 0.1`
  - 并发进行多路召回（FULLTEXT + 倒排索引 + 同义词扩展），合并排序

- [ ] 🟡 **PageRank**
  - 基于论文引用关系构建有向图
  - 计算学术影响力权重，融入 S.H.I.T Score

### 2.3 分词 & 文本处理

> [!NOTE]
> 本节所有方案均采用 **规则可解释** 的算法，不引入神经网络 / 大模型等黑箱。
> 搜索结果影响论文可见度，间接关联治理，因此必须保持算法透明。

- [ ] 🔴 **IK / Jieba 分词器集成**
  - 中文论文标题/摘要分词
  - 自定义词典：学术术语 + S.H.I.T 专有名词
  - 分词结果用于倒排索引构建 + 搜索 query 解析

- [ ] 🟡 **同义词词典 & 学术术语扩展**
  - 手工维护同义词表："量子计算" ↔ "量子信息" ↔ "quantum computing"
  - 搜索时自动扩展 query：用户输入 "量子计算" → 同时搜索同义词
  - 高贡献用户（editor+）可提交同义词建议，社区投票通过后生效

- [ ] 🟡 **编辑距离 & 拼音模糊匹配**
  - 利用 Levenshtein Distance 实现中英文模糊搜索
  - 拼音索引："liangzi" → "量子"，覆盖拼音输入场景
  - 完全基于确定性算法，无黑箱

### 2.4 搜索体验

- [ ] 🔴 **前缀树 (Trie) 联想提示**
  - 输入 "东方明" → 提示 "东方明珠"
  - 哈夫曼编码压缩 Trie Tree，减少内存占用
  - 分离 Trie Tree 的 build 和 recall 过程
  - 一致性哈希分片存储：Trie Tree + Inverted Index 按 hash(term) 分片

- [ ] 🟡 **Query 纠错**
  - 输入 "陆加嘴" → 纠正为 "陆家嘴"
  - 方案：编辑距离 + 拼音模糊匹配 + 自定义纠错词典

- [ ] 🟡 **分页 & 排序**
  - 搜索结果深分页优化：search_after 替代 offset
  - 支持按 `shit_score` / `created_at` / `rating_count` / 相关性 排序

- [ ] 🟢 **搜索缓存隔离**
  - 当前问题：连续搜索不同 query，未清除缓存导致结果产生并集
  - 修复：每次搜索前清空上一次的召回结果集

### 2.5 工程效率

- [ ] 🔴 **一键启动脚本**
  - 写一个 `start.sh`，按依赖顺序启动所有服务
  - 支持 `./start.sh dev` / `./start.sh prod` 环境切换
  - 服务健康检查 + 自动重启

```bash
#!/bin/bash
# start.sh - 一键启动 ShitJournal 所有服务
services=("user-rpc" "paper-rpc" "rating-rpc" "news-rpc" "api")
for svc in "${services[@]}"; do
    echo "Starting $svc..."
    make $svc &
    sleep 1
done
echo "All services started. PID list:"
jobs -l
```

- [ ] 🟡 **Proto 变更热加载**
  - 问题：子模块 proto 改动后 gateway 需要重启才生效
  - 原因：gRPC client stub 编译时生成，proto 变更需重新 `goctl` + 重编译
  - 方案：Makefile 增加 `watch-proto` target，用 `fswatch` 监听 `.proto` 文件变更自动重新生成 + 重启

### 2.6 社区自治演进

> 基于已完成的四大自治模块（§0.6），持续增强去中心化治理能力。

#### 2.6.1 贡献度系统增强

- [ ] 🔴 **增量贡献度计算**
  - 当前方案：cron 每日全量重算所有用户的 `contribution_score`
  - 改造方案：评分时实时增量更新评分者 + 论文作者的贡献分
  - 减少 cron 压力，实现近实时贡献度更新
- [ ] 🟡 **贡献度可视化仪表盘**
  - 用户个人页展示贡献分组成（创作 / 审阅 / 活跃度 / 衰减）
  - 贡献度排行榜（Top 50 Reviewers / Authors）
- [ ] 🟡 **跨论文审阅一致性检测**
  - 当前仅检测连续 5 篇相同评分的 spam
  - 增强：检测评分分布异常（所有评分集中在 1 或 10 分）、与社区均值的系统性偏差
  - 引入 Z-Score 异常检测
- [ ] 🟢 **IP / 设备指纹检测**
  - 防小号互评：记录 `rating` 时关联的 IP / User-Agent
  - 检测同一 IP 大量不同用户评分同一论文

#### 2.6.2 S.H.I.T Score v3

- [ ] 🟡 **引入引用网络权重**
  - 当论文间引用关系建立后（§2.2 PageRank），将引用权重融入 Score
  - `v3 = v2 + 0.10 × citation_influence - 0.05 × self_citation_ratio`
- [ ] 🟡 **学科归一化**
  - 不同学科的评分分布差异大（如理工科评分偏低、人文社科偏高）
  - 按学科 `discipline` 进行 Z-Score 归一化后再计算 Score
- [ ] 🟢 **时间序列 Score**
  - 记录 `shit_score` 的历史变化曲线
  - 新增 `paper_score_history` 表：快照 score + timestamp
  - 支持"趋势上升""争议爆发"等论文状态标签

#### 2.6.3 冲突解决增强

- [ ] 🔴 **Flag RPC 服务**
  - 新增 `proto/flag.proto` + `rpc/flag/` RPC 服务
  - 三个方法：`SubmitFlag`、`GetFlagStatus`、`ListFlagsByTarget`
  - API 网关新增举报端点：`POST /papers/:id/flag`、`POST /ratings/:id/flag`
- [ ] 🟡 **关键词黑名单热更新**
  - 当前 `common/degradation/keyword_filter.go` 为静态配置
  - 改造：关键词存 Redis / MySQL，支持管理员（贡献分 ≥200）动态添加
  - 支持正则表达式 + 拼音模糊匹配
- [ ] 🟡 **SimHash 内容重复检测**
  - 论文投稿时计算 SimHash 指纹
  - 与已有论文比对 Hamming Distance，≤ 3 位差异 → 标记疑似抄袭
  - 在 `paper` 表新增 `simhash` BIGINT 字段
- [ ] 🟡 **恶意评分实时检测**
  - 短时间内大量评分（burst detection）
  - 双峰分布检测（bimodality coefficient > 0.55）
  - 同 IP 聚集检测
- [ ] 🟢 **举报申诉机制**
  - 被降解论文的作者可提交申诉
  - 申诉需 3 名 editor 级别用户（贡献分 ≥50）投票通过
  - 新增 `appeal` 表 + Appeal RPC 服务

#### 2.6.4 激励与资源优化

- [ ] 🔴 **限流中间件集成**
  - 将 `common/ratelimit/limiter.go` 集成到 API Gateway 的 middleware 中
  - 评分 / 举报 / 搜索的 3 个限流器作为中间件自动拦截
  - 返回 `429 Too Many Requests` + `Retry-After` header
- [ ] 🟡 **缓存体系建设**
  - 各区热门论文缓存（Top 100，TTL 5 min）
  - 用户贡献分缓存（TTL 1h）
  - 举报 quorum 缓存（TTL 10 min）
  - 论文降解等级缓存（TTL 30 min）
- [ ] 🟡 **事件驱动架构**
  - 评分时异步发布事件 → 消费者更新贡献分、检测 spam、检查 quorum
  - 使用 go-zero 内置的 `queue` 或 Redis Pub/Sub
  - 减少评分 API 的响应时间（从同步计算改为异步）
- [ ] 🟢 **贡献度 NFT / 成就徽章**
  - 里程碑成就系统：首次投稿、首篇晋升 Sediment、审阅 100 篇等
  - 徽章显示在用户个人页，增强社区参与动机

### 2.7 前端开发

- [ ] 🔴 **核心页面开发**
  - 首页：论文热榜 + 各 Zone 概览 + 新闻公告
  - 论文列表页：Zone/学科筛选 + 排序 + 分页 + 搜索
  - 论文详情页：元信息 + 全文阅读 + 评分面板 + 评分分布图
  - 用户主页：个人论文 + 评分历史 + 贡献度雷达图
  - 投稿页：Markdown 编辑器 + 元信息表单 + 文件上传
  - 登录/注册页：表单验证 + JWT 存储
- [ ] 🔴 **技术选型 & 脚手架**
  - 框架：Next.js 14+ (App Router) 或 Nuxt 3
  - UI 库：Shadcn/UI + Radix Primitives
  - 状态管理：Zustand / Pinia
  - 请求层：封装 Axios / ofetch，统一拦截 JWT 刷新 + 错误码映射
- [ ] 🟡 **数据可视化**
  - 论文 S.H.I.T Score 变化曲线（ECharts / Recharts）
  - 评分分布直方图
  - 学科热度气泡图
  - 贡献度雷达图（创作 / 审阅 / 活跃 / 声望）
- [ ] 🟡 **搜索体验**
  - 搜索框联想提示（对接 Trie Tree API）
  - 搜索结果高亮
  - 高级筛选面板：学科 + Zone + 时间范围 + 评分区间
- [ ] 🟡 **暗色模式 & 国际化**
  - CSS 变量 + `prefers-color-scheme` 媒体查询
  - i18n：中文 / English 双语切换
  - 日期格式本地化
- [ ] 🟢 **SSR & SEO**
  - 论文详情页 SSR（提升搜索引擎收录）
  - Open Graph / Twitter Cards meta 标签
  - Sitemap 自动生成
- [ ] 🟢 **PWA & 移动适配**
  - Service Worker 缓存离线可读论文
  - 响应式布局：Mobile / Tablet / Desktop

### 2.8 CI/CD & 部署

- [ ] 🔴 **GitHub Actions CI**
  - Push / PR 触发：`go build ./...` + `go test ./...` + `go vet ./...`
  - Proto lint：`buf lint` 检查 `.proto` 文件规范
  - Docker 镜像构建：多阶段构建，最终镜像 < 30MB
- [ ] 🟡 **K8s 部署迁移**
  - 从 Docker Compose 迁移到 Kubernetes / K3s
  - 每个 RPC 服务一个 Deployment + Service
  - HPA 自动扩缩：基于 CPU / 自定义 metrics（QPS）
  - Ingress 暴露 API Gateway
- [ ] 🟡 **配置中心**
  - 将硬编码的 DSN、JWT Secret、限流阈值等迁移到配置中心
  - 方案：Etcd KV（已有 Etcd 基础设施）或 Nacos
  - 支持配置热更新，无需重启服务
- [ ] 🟡 **日志聚合**
  - 引入 Loki + Promtail 或 ELK Stack
  - 结构化日志：go-zero 原生 `logx` JSON 输出
  - 统一日志查询：按 trace_id / service / 时间范围
- [ ] 🟢 **蓝绿 / 灰度发布**
  - Kubernetes 蓝绿发布：新旧版本共存，流量切换
  - 灰度发布：按用户 ID 哈希或 header 标记分流

---

## 三、优先级总览

| 优先级 | 方向 | 具体项 |
|--------|------|--------|
| 🔴 P0 | 自治 | Flag RPC 服务、限流中间件集成、增量贡献度计算 |
| 🔴 P0 | 前端 | 核心页面开发、技术选型 & 脚手架 |
| 🔴 P0 | 可观测 | Jaeger 全链路追踪、Prometheus + Grafana 监控 |
| 🔴 P0 | 数据层 | DAO Init 抽离、一键启动脚本 |
| 🔴 P0 | 搜索 | 索引并发构建、TF-IDF/BM25、IK 分词 |
| 🔴 P0 | CI/CD | GitHub Actions CI |
| 🟡 P1 | 自治 | SimHash 查重、恶意评分检测、缓存体系、事件驱动 |
| 🟡 P1 | 自治 | S.H.I.T Score v3（引用网络 + 学科归一化）、关键词热更新 |
| 🟡 P1 | 前端 | 数据可视化、搜索体验、暗色模式 & i18n |
| 🟡 P1 | 架构 | 熔断降级、冷热分离 |
| 🟡 P1 | 搜索 | Trie 联想、Query 纠错、同义词扩展、PageRank |
| 🟡 P1 | 索引 | 倒排索引压缩 + mmap、Roaring Bitmap |
| 🟡 P1 | CI/CD | K8s 部署迁移、配置中心、日志聚合 |
| 🟢 P2 | 自治 | IP/设备指纹检测、申诉机制、成就徽章 |
| 🟢 P2 | 自治 | 时间序列 Score 历史 |
| 🟢 P2 | 前端 | SSR/SEO、PWA & 移动适配 |
| 🟢 P2 | 存储 | StarRocks OLAP 迁移、分布式 MapReduce 索引 |
| 🟢 P2 | 搜索 | 拼音索引、动态索引、一致性哈希分片 |
| 🟢 P2 | CI/CD | 蓝绿 / 灰度发布 |
