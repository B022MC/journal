export type Priority = "🔴 High" | "🟡 Medium" | "🟢 Low";

export interface RoadmapItem {
    id: string;
    title: string;
    priority?: Priority;
    completed: boolean;
    details?: string[];
}

export interface RoadmapCategory {
    id: string;
    title: string;
    items: RoadmapItem[];
}

export const roadmapData: RoadmapCategory[] = [
    {
        id: "completed",
        title: "零、已完成 ✅",
        items: [
            {
                id: "c-1",
                title: "项目基础设施",
                completed: true,
                details: [
                    "Monorepo 项目结构（backend/ + frontend/）",
                    "go-zero 微服务骨架：goctl 生成 4 个 RPC + 1 个 API 网关",
                    "Docker Compose：MySQL 8.0（1 Master + 2 Replica）+ Redis 7 + Etcd 3.5",
                    "MySQL GTID 主从复制配置（deploy/mysql/ 下 cnf + init 脚本）",
                    "PostgreSQL DDL → MySQL DDL 迁移（model/schema.sql）",
                    "Makefile 全套 target（build / run / proto-gen / env-up / init-repl）"
                ]
            },
            {
                id: "c-2",
                title: "后端核心服务",
                completed: true,
                details: [
                    "User Service :9001 — 注册/登录/获取信息/更新资料",
                    "Paper Service :9002 — 投稿/列表/详情/搜索/Zone更新/用户论文",
                    "Rating Service :9003 — 评分Upsert/防自评/S.H.I.T Score实时重算/按论文或用户查评分",
                    "News Service :9004 — 新闻 CRUD / 置顶排序"
                ]
            },
            {
                id: "c-3",
                title: "数据层",
                completed: true,
                details: [
                    "go-zero sqlx.SqlConn Model 层使用",
                    "MySQL 读写分离",
                    "注册唯一性检查走主库",
                    "S.H.I.T Score 算法",
                    "MySQL FULLTEXT 搜索",
                    "Rating Upsert: ON DUPLICATE KEY UPDATE"
                ]
            },
            {
                id: "c-4",
                title: "API 网关",
                completed: true,
                details: [
                    "16 个 API Logic 文件",
                    "17 个 Handler 文件（含 routes.go）",
                    "JWT 鉴权分组",
                    "API 网关 ServiceContext 连接 4 个 RPC Client"
                ]
            },
            {
                id: "c-5",
                title: "公共包 & 工具",
                completed: true,
                details: [
                    "common/jwt — JWT 生成 / 解析",
                    "common/errorx — 双语错误码",
                    "common/result — 统一 JSON 响应封装",
                    "cmd/lifecycle/main.go — 论文生命周期定时任务"
                ]
            },
            {
                id: "c-6",
                title: "社区自治模块",
                completed: true,
                details: [
                    "S.H.I.T Score v2 算法",
                    "多维度晋升/降级阈值",
                    "贡献度计算器 common/contribution/calculator.go",
                    "自动角色分配",
                    "降解引擎 common/degradation/engine.go",
                    "举报系统 model/flagmodel.go",
                    "Redis Token Bucket 限流器",
                    "统一定时任务引擎 cmd/cron/main.go",
                    "数据库迁移 model/migrations/002_governance_modules.sql"
                ]
            }
        ]
    },
    {
        id: "architecture",
        title: "一、架构演进",
        items: [
            {
                id: "a-1",
                title: "引入 Jaeger 全链路追踪",
                priority: "🔴 High",
                completed: true
            },
            {
                id: "a-2",
                title: "引入 Prometheus + Grafana 监控",
                priority: "🔴 High",
                completed: true
            },
            {
                id: "a-3",
                title: "引入 降级熔断",
                priority: "🟡 Medium",
                completed: true
            },
            {
                id: "a-4",
                title: "数据层抽象 (抽离 DAO Init，按 Key 获取数据库实例)",
                priority: "🔴 High",
                completed: true
            },
            {
                id: "a-5",
                title: "冷热数据分离 (定义冷热标准和分离存储)",
                priority: "🟡 Medium",
                completed: true
            },
            {
                id: "a-6",
                title: "参考 Elasticsearch 的 ILM 策略",
                priority: "🟢 Low",
                completed: false
            },
            {
                id: "a-7",
                title: "MySQL → StarRocks（OLAP）",
                priority: "🟢 Low",
                completed: false
            }
        ]
    },
    {
        id: "features",
        title: "二、功能迭代",
        items: [
            {
                id: "f-1",
                title: "搜索引擎核心",
                completed: false,
                details: [
                    "🔴 High: 索引构建并发化",
                    "🔴 High: 倒排索引优化",
                    "🟡 Medium: Roaring Bitmap 存储 DocID",
                    "🟡 Medium: 分布式 MapReduce 构建索引"
                ]
            },
            {
                id: "f-2",
                title: "相关性排序",
                completed: false,
                details: [
                    "🔴 High: TF-IDF / BM25 相关性计算",
                    "🟡 Medium: 排序器优化",
                    "🟡 Medium: PageRank"
                ]
            },
            {
                id: "f-3",
                title: "分词 & 文本处理",
                completed: false,
                details: [
                    "🔴 High: IK / Jieba 分词器集成",
                    "🟡 Medium: 同义词词典 & 学术术语扩展",
                    "🟡 Medium: 编辑距离 & 拼音模糊匹配"
                ]
            },
            {
                id: "f-4",
                title: "搜索体验",
                completed: false,
                details: [
                    "🔴 High: 前缀树 (Trie) 联想提示",
                    "🟡 Medium: Query 纠错",
                    "🟡 Medium: 分页 & 排序",
                    "🟢 Low: 搜索缓存隔离"
                ]
            },
            {
                id: "f-5",
                title: "工程效率",
                completed: false,
                details: [
                    "🔴 High: 一键启动脚本",
                    "🟡 Medium: Proto 变更热加载"
                ]
            },
            {
                id: "f-6",
                title: "社区自治演进",
                completed: false,
                details: [
                    "🔴 High: 增量贡献度计算",
                    "🟡 Medium: 贡献度可视化仪表盘",
                    "🟡 Medium: 跨论文审阅一致性检测",
                    "🟢 Low: IP / 设备指纹检测（已完成）",
                    "🟡 Medium: 引入引用网络权重 (Score v3)",
                    "🟡 Medium: 学科归一化 (Score v3)",
                    "🟢 Low: 时间序列 Score",
                    "🔴 High: Flag RPC 服务",
                    "🟡 Medium: 关键词黑名单热更新",
                    "🟡 Medium: SimHash 内容重复检测",
                    "🟡 Medium: 恶意评分实时检测（已完成）",
                    "🟢 Low: 举报申诉机制",
                    "🔴 High: 限流中间件集成",
                    "🟢 Low: 缓存体系建设（已完成）",
                    "🟢 Low: 事件驱动架构（评分 / 举报后处理异步化已完成）",
                    "🟢 Low: 贡献度 NFT / 成就徽章（后端成就系统已完成）"
                ]
            },
            {
                id: "f-7",
                title: "前端开发",
                completed: false,
                details: [
                    "🔴 High: 核心页面开发",
                    "🔴 High: 技术选型 & 脚手架",
                    "🟡 Medium: 数据可视化",
                    "🟡 Medium: 搜索体验",
                    "🟡 Medium: 暗色模式 & 国际化",
                    "🟢 Low: SSR & SEO",
                    "🟢 Low: PWA & 移动适配"
                ]
            },
            {
                id: "f-8",
                title: "CI/CD & 部署",
                completed: false,
                details: [
                    "🔴 High: GitHub Actions CI",
                    "🟡 Medium: K8s 部署迁移",
                    "🟡 Medium: 配置中心",
                    "🟡 Medium: 日志聚合",
                    "🟢 Low: 蓝绿 / 灰度发布"
                ]
            }
        ]
    }
];
