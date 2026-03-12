# S.H.I.T Journal 前端设计方案

## 1. 目标

这份文档只定义 **主站前端** 的产品设计与实现基线，不直接覆盖管理后台。

## 1.1 2026-03-12 执行基线

从 `2026-03-12` 这轮执行开始，主站第一阶段默认以这份文档作为正式视觉与实现基线，不再以当前 `roadmap` 展示页的视觉风格作为交付参考。

本轮冻结范围只包含下面三类事项：

- 主站 Phase 0 / Phase 1：全局壳层、首页、论文列表、论文详情、登录、注册
- 搜索核心能力：搜索 ADR、索引构建、分词、倒排索引、排序、Explain，以及接回主站前必须保留的降级链路
- 必要工程补强：统一 API 访问、环境变量约定、错误处理、契约/冒烟测试、README 与 CI 约束

本轮明确不提前展开：

- Phase 2 用户闭环页面：`/submit`、`/me`、用户公开主页、徽章展示
- 体验增强项：新闻、图表、manifesto、搜索高级体验
- 平台级优化：SEO、i18n、PWA、额外发布包装

边界说明：

- 鉴权只要求预留统一接口与后续迁移位，不在本轮内强行定死 `httpOnly cookie` 与浏览器存储的最终实现
- 新搜索未达到 ADR 约定阈值前，默认回退路径仍然是 MySQL FULLTEXT
- 后续如果要扩大范围，必须同步更新这份文档和对应的执行快照，不能只靠口头变更

当前仓库现状很明确：

- `frontend/` 是 Next.js 16 + React 19 的空骨架
- `frontend/admin-web/` 是独立的后台应用
- 后端 API 已经具备论文、评分、举报、新闻、用户信息、徽章等核心能力

因此前端应拆成两条线：

- **主站**：面向普通用户，负责阅读、搜索、投稿、评分、社区自治可视化
- **后台**：面向管理员，继续维持独立应用，不和主站合并

## 2. 设计结论

### 2.1 产品定位

主站不是普通 SaaS 仪表盘，也不是学术论文检索站的翻版。它应该更像：

- 一个带实验气质的学术废料档案馆
- 一个公开展示社区自治结果的“编辑实验室”
- 一个以文本阅读和治理透明度为中心的内容产品

### 2.2 视觉方向

主站视觉不建议延续当前 `roadmap` 页那种纯黑科技感。那套风格更像工程看板，不像内容产品。

建议主站使用：

- **视觉主题**：`Archive Lab`，即“档案室 + 实验台”
- **背景基调**：暖纸色，不以黑底为默认
- **信息对比**：深墨色正文 + 高密度边框 + 少量风险色
- **气质关键词**：编辑感、审阅感、沉积感、可考据

建议色板：

- `bg.base`: `#f3efe6`
- `bg.panel`: `#e4dccd`
- `ink.strong`: `#171411`
- `ink.muted`: `#6b6258`
- `accent.rust`: `#8a4b2a`
- `accent.moss`: `#426b54`
- `accent.slate`: `#53606d`
- `signal.warn`: `#b06a2d`
- `signal.danger`: `#8b312e`

Zone 颜色不要做成花哨徽章，保持治理语义：

- `latrine`: moss
- `septic_tank`: rust
- `stone`: slate
- `sediment`: dark olive

### 2.3 字体方向

不要继续用 `Inter` 作为主站默认字体。

建议：

- 标题：`Newsreader`
- 正文/UI：`IBM Plex Sans`
- 数据/编号/DOI/分数：`IBM Plex Mono`

这样能同时保留“论文阅读感”和“系统界面感”。

## 3. 信息架构

## 3.1 主站路由

```text
/
/papers
/papers/[id]
/search
/submit
/login
/register
/news
/users/[id]
/me
/manifesto
```

说明：

- `/` 是内容门户，不是营销首页
- `/papers` 是主内容索引页
- `/papers/[id]` 是最重要的阅读页
- `/search` 可以先与 `/papers` 共用实现，后续再拆高级搜索
- `/users/[id]` 用于公开个人主页
- `/me` 是登录用户的工作台
- `/manifesto` 用来承接 S.H.I.T Journal 的治理理念，不要埋在 README 里

## 3.2 页面优先级

以下优先级描述的是长期产品路线，不等于当前迭代交付范围；本轮执行边界以前面的 `1.1 2026-03-12 执行基线` 为准。

P0：

- 首页
- 论文列表页
- 论文详情页
- 登录/注册页
- 投稿页
- 我的主页 `/me`

P1：

- 用户公开主页
- 新闻页
- 搜索增强
- 评分分布图
- 贡献度图
- 徽章展示

P2：

- manifesto 页面
- SSR/SEO 完整优化
- PWA
- i18n

## 4. 页面蓝图

## 4.1 首页 `/`

首页不是大 Banner + 几个按钮，而是一个“实时编辑台”。

应包含 6 个区块：

1. 顶部导语
   - 一句清晰定义产品
   - 两个主动作：`Browse Papers`、`Submit Paper`
2. 热门论文带
   - 按 `shit_score` 排的前 6 篇
3. Zone 状态总览
   - 四个 Zone 的数量、特征、最新流入
4. 新闻公告侧栏
   - 置顶新闻 + 最近公告
5. 社区治理说明
   - 解释评分、举报、降级、角色如何运作
6. 最近活跃
   - 最近被评分 / 被举报 / 晋升到新 Zone 的论文

首页布局建议：

- Desktop: `8 / 4` 双栏
- Mobile: 单栏，新闻和治理说明后置

## 4.2 论文列表页 `/papers`

这是主流量入口，必须优先设计得像“可检索的内容档案”，不是普通卡片瀑布流。

结构建议：

- 顶部固定搜索条
- 左侧筛选面板
  - Zone
  - Discipline
  - Sort
  - Rating range
  - Time range
- 右侧结果区
  - 结果总数
  - 当前筛选 breadcrumb
  - 论文列表

单个论文卡片应至少展示：

- 标题
- 作者
- 学科
- Zone
- S.H.I.T Score
- Avg rating / rating count
- 浏览量
- 举报/降级状态
- 简短摘要

排序默认值建议：

- 默认 `highest_rated`
- 可切 `newest` / `most_viewed`

## 4.3 论文详情页 `/papers/[id]`

这页是整个产品体验中心，必须用“阅读页”思路设计，而不是信息罗列。

布局建议：

- 主阅读栏：标题、作者、DOI、关键词、摘要、正文
- 右侧治理栏：
  - 当前 Zone
  - S.H.I.T Score
  - 平均分 / 评分人数
  - 争议度
  - 举报状态
  - 评分入口

详情页需要的模块：

- `PaperHeader`
- `ZoneBadge`
- `ScoreDial`
- `ModerationMeter`
- `RatingComposer`
- `RatingHistogram`
- `RatingsFeed`
- `RelatedPapers`

交互原则：

- 阅读优先，评分操作永远不遮挡正文
- 举报入口放在治理栏，不抢主视觉
- 评分分布图优先于花哨图表

## 4.4 投稿页 `/submit`

投稿页不能只是一个大表单，应该是“分步式编辑台”。

建议做成三段：

1. 元信息
   - 标题
   - 英文标题
   - 学科
   - 关键词
2. 内容编辑
   - 摘要
   - 英文摘要
   - 正文
3. 提交前检查
   - 必填项校验
   - 重复检测提示
   - 预估初始状态说明

编辑区建议：

- Desktop 左右分栏：编辑 / 预览
- Mobile 切换 tab：编辑 / 预览

## 4.5 登录/注册页

不要做成单纯居中的 SaaS 登录框。

建议布局：

- 左侧表单
- 右侧产品理念摘要
- 右侧同时展示：
  - 社区评分机制
  - 论文四区流转
  - 治理透明原则

这样新用户在登录页就知道这不是普通论坛。

## 4.6 我的主页 `/me`

这页是“个人工作台”，不是单纯 profile。

应有 4 个主 tab：

- `Overview`
- `My Papers`
- `My Ratings`
- `Governance`

内容建议：

- 贡献分总览
- 角色与权限说明
- 已解锁徽章
- 最近投稿
- 最近评分
- 被举报记录或参与举报记录

## 4.7 用户公开主页 `/users/[id]`

这页偏“公开档案页”。

展示：

- 公开身份信息
- 角色
- 贡献分
- 徽章
- 论文列表
- 评分历史摘要

不展示：

- 私有操作入口
- 账号设置

## 4.8 新闻页 `/news`

新闻页承担两个作用：

- 公告流
- 社区治理变更记录

建议不是单独“博客页”，而是带分类切换：

- `Announcement`
- `Governance`
- `Release`

## 5. 组件体系

## 5.1 Shell 级组件

- `SiteHeader`
- `PrimaryNav`
- `GlobalSearchDock`
- `FooterManifest`
- `ZoneLegend`

## 5.2 内容组件

- `PaperCard`
- `PaperRow`
- `PaperMetaStrip`
- `ScoreDial`
- `ZonePill`
- `ModerationMeter`
- `KeywordCluster`
- `NewsRail`

## 5.3 社区治理组件

- `RatingComposer`
- `RatingHistogram`
- `FlagStatusPanel`
- `ContributionBreakdown`
- `AchievementShelf`
- `GovernanceTimeline`

## 5.4 表单组件

- `AuthPanel`
- `SubmitStepper`
- `MetadataForm`
- `MarkdownWorkbench`
- `PreviewPane`

## 6. 与现有后端 API 的对接

当前主站第一阶段只围绕已经稳定存在的接口设计。

核心接口映射：

| 页面 | 接口 |
|------|------|
| 首页热榜 / 列表页 | `GET /api/v1/papers` |
| 搜索页 | `GET /api/v1/papers/search` |
| 论文详情页 | `GET /api/v1/papers/:id` |
| 详情页评分列表 | `GET /api/v1/papers/:id/ratings` |
| 详情页评分动作 | `POST /api/v1/papers/:id/rate` |
| 举报论文 | `POST /api/v1/papers/:id/flag` |
| 举报评分 | `POST /api/v1/ratings/:id/flag` |
| 举报状态 | `GET /api/v1/papers/:id/flag-status` / `GET /api/v1/ratings/:id/flag-status` |
| 登录/注册 | `POST /api/v1/user/login` / `POST /api/v1/user/register` |
| 当前用户信息 | `GET /api/v1/user/info` |
| 我的论文/评分 | `GET /api/v1/user/papers` / `GET /api/v1/user/ratings` |
| 投稿 | `POST /api/v1/papers/submit` |
| 新闻 | `GET /api/v1/news` / `GET /api/v1/news/:id` |

## 7. 前端技术方案

## 7.1 保留现有基础

主站继续使用现有：

- Next.js App Router
- React 19
- Tailwind v4

不建议此时重选 Nuxt 或另起一套栈。

## 7.2 建议补齐

- 数据获取：优先原生 `fetch` + 封装 `lib/api`
- 服务端数据页：默认 Server Components
- 客户端交互区：评分、举报、筛选器、编辑器
- 本地状态：`Zustand` 只用于 session 和轻量 UI 状态
- 图表：`Recharts`
- 表单：`react-hook-form` + `zod`
- Markdown 编辑：先用轻量编辑器，不要一开始就上复杂协同

## 7.3 鉴权策略

当前后端是 JWT Bearer 风格。

前端目标策略：

- 优先把 token 落到 `httpOnly cookie` 的桥接方案
- 如果第一阶段来不及做 BFF，短期用浏览器存储，但实现时要预留迁移位

不要把鉴权设计成全站 CSR。

## 8. 布局与响应式规则

## 8.1 栅格

- Desktop：12 栏，最大宽度 `1440px`
- Tablet：8 栏
- Mobile：4 栏

## 8.2 响应式原则

- 列表页：筛选器在移动端收进抽屉
- 详情页：治理栏在移动端下移到正文后
- 投稿页：预览与编辑在移动端用 tab 切换
- 首页：热榜卡片从 `3 列` 收到 `1 列`

## 8.3 动效

只做三类动效：

- 页面首屏渐入
- 区块顺序浮现
- 评分/举报结果状态切换

不要上大量 hover 装饰和悬浮玻璃效果。

## 9. 第一阶段实现顺序

### Phase 0: 基础设施

- 重做 `layout.tsx`
- 建立字体、色板、间距、圆角 token
- 建 `SiteHeader` / `Footer` / `Container`
- 建 `lib/api`

### Phase 1: 核心主路径

- 首页
- 论文列表页
- 论文详情页
- 登录/注册

### Phase 2: 用户闭环

- 投稿页
- `/me`
- 用户公开主页
- 徽章展示

### Phase 3: 体验增强

- 新闻页
- 搜索增强
- 图表
- manifesto

### Phase 4: 平台级优化

- SSR / SEO
- i18n
- PWA

## 10. 明确不这么做

- 不把主站做成黑底后台面板
- 不先做大而全的数据可视化屏
- 不把主站和后台合并成一个应用
- 不先做主题切换器再做内容页
- 不先做复杂动画再做阅读体验

## 11. 下一步执行建议

从这份设计出发，真正开做时应按下面顺序：

1. 重做主站全局视觉系统
2. 先完成首页 / 列表 / 详情 / 登录
3. 再补投稿与个人页
4. 最后接图表、新闻、SEO、PWA

如果后面继续实现，默认以这份文档和后续对应的执行快照为准，不再以当前 `roadmap` 页的视觉风格为准。
