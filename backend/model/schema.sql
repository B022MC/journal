-- =============================================================
-- S.H.I.T Journal — 完整初始化脚本
-- 包含：DDL（建库/建表）+ DML（种子数据）
-- 执行顺序：直接 source 此文件即可
-- =============================================================

-- ---------------------------------------------------------------
-- 1. 数据库 & 用户
-- ---------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `journal_biz`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS `journal_admin`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'journal'@'%' IDENTIFIED BY 'banishmentB022.';
GRANT ALL PRIVILEGES ON `journal_biz`.* TO 'journal'@'%';
GRANT ALL PRIVILEGES ON `journal_admin`.* TO 'journal'@'%';
FLUSH PRIVILEGES;


-- ===============================================================
-- 2. 业务库 journal_biz
-- ===============================================================
USE `journal_biz`;

-- ---------------------------------------------------------------
-- 2.1 用户表
-- role: 0=member, 1=scooper, 2=editor, 3=admin
-- status: 0=banned, 1=active
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user` (
  `id`                 bigint unsigned  NOT NULL AUTO_INCREMENT,
  `username`           varchar(64)      NOT NULL,
  `email`              varchar(128)     NOT NULL,
  `password_hash`      varchar(256)     NOT NULL,
  `nickname`           varchar(64)      DEFAULT '',
  `avatar`             varchar(512)     DEFAULT '',
  `role`               tinyint          NOT NULL DEFAULT 0
                         COMMENT '0=member,1=scooper,2=editor',
  `contribution_score` decimal(10,2)    DEFAULT 0.00,
  `last_active_at`     timestamp        NULL DEFAULT NULL
                         COMMENT '最后活跃时间',
  `review_count_30d`   int              DEFAULT 0
                         COMMENT '近30天评审数',
  `status`             tinyint          NOT NULL DEFAULT 1
                         COMMENT '0=banned,1=active',
  `created_at`         timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`         timestamp        DEFAULT CURRENT_TIMESTAMP
                         ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email`    (`email`),
  KEY `idx_role`           (`role`),
  KEY `idx_status`         (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='业务用户表';

CREATE TABLE IF NOT EXISTS `user_achievement` (
  `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id`     bigint unsigned NOT NULL,
  `code`        varchar(64)     NOT NULL
                  COMMENT 'first_submission/sediment_breakthrough/reviewer_century',
  `unlocked_at` timestamp       DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_code` (`user_id`, `code`),
  KEY `idx_user_unlocked_at` (`user_id`, `unlocked_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户成就徽章解锁记录';

-- ---------------------------------------------------------------
-- 2.2 论文表
-- zone: latrine, septic_tank, stone, sediment
-- discipline: science, humanities, information, technology, other
-- status: 0=deleted, 1=active, 2=flagged
-- degradation_level: 0=normal, 1=watched, 2=throttled, 3=sealed
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `paper` (
  `id`                 bigint unsigned  NOT NULL AUTO_INCREMENT,
  `title`              varchar(512)     NOT NULL,
  `title_en`           varchar(512)     DEFAULT '',
  `abstract`           text             NOT NULL,
  `abstract_en`        text             DEFAULT NULL,
  `content`            longtext         NOT NULL,
  `author_id`          bigint unsigned  NOT NULL,
  `author_name`        varchar(64)      NOT NULL,
  `discipline`         varchar(32)      NOT NULL DEFAULT 'other'
                         COMMENT 'science,humanities,information,technology,other',
  `zone`               varchar(32)      NOT NULL DEFAULT 'latrine'
                         COMMENT 'latrine,septic_tank,stone,sediment',
  `shit_score`         decimal(10,4)    DEFAULT 0.0000
                         COMMENT 'S.H.I.T composite score',
  `avg_rating`         decimal(5,2)     DEFAULT 0.00,
  `rating_count`       int              DEFAULT 0,
  `view_count`         int              DEFAULT 0,
  `controversy_index`  decimal(5,4)     DEFAULT 0.0000
                         COMMENT '评分标准差/争议度',
  `weighted_avg_rating` decimal(5,2)   DEFAULT 0.00
                         COMMENT '加权平均评分',
  `reviewer_authority` decimal(5,4)    DEFAULT 0.0000
                         COMMENT '评审者权威度均值',
  `flag_count`         int              DEFAULT 0
                         COMMENT '被举报次数',
  `degradation_level`  tinyint          DEFAULT 0
                         COMMENT '0=normal,1=watched,2=throttled,3=sealed',
  `file_path`          varchar(512)     DEFAULT '',
  `doi`                varchar(128)     DEFAULT '',
  `keywords`           varchar(512)     DEFAULT '',
  `simhash`            bigint unsigned  NOT NULL DEFAULT 0
                         COMMENT '内容 SimHash 指纹',
  `status`             tinyint          NOT NULL DEFAULT 1
                         COMMENT '0=deleted,1=active,2=flagged',
  `promoted_at`        timestamp        NULL DEFAULT NULL,
  `last_accessed_at`   timestamp        NULL DEFAULT NULL
                         COMMENT '最后被访问时间',
  `created_at`         timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`         timestamp        DEFAULT CURRENT_TIMESTAMP
                         ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_zone`            (`zone`),
  KEY `idx_discipline`      (`discipline`),
  KEY `idx_author`          (`author_id`),
  KEY `idx_shit_score`      (`shit_score`),
  KEY `idx_status`          (`status`),
  KEY `idx_simhash`         (`simhash`),
  KEY `idx_zone_discipline` (`zone`, `discipline`),
  KEY `idx_created`         (`created_at`),
  KEY `idx_last_accessed`   (`last_accessed_at`),
  FULLTEXT KEY `ft_title_abstract` (`title`, `abstract`, `keywords`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='论文主表';

-- ---------------------------------------------------------------
-- 2.3 冷数据归档表（热论文 → 冷归档，只追加不更新）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `cold_paper` (
  `id`                 bigint unsigned  NOT NULL
                         COMMENT '与 paper.id 保持一致，不自增',
  `title`              varchar(512)     NOT NULL,
  `title_en`           varchar(512)     DEFAULT '',
  `abstract`           text             NOT NULL,
  `abstract_en`        text             DEFAULT NULL,
  `content`            longtext         NOT NULL,
  `author_id`          bigint unsigned  NOT NULL,
  `author_name`        varchar(64)      NOT NULL,
  `discipline`         varchar(32)      NOT NULL DEFAULT 'other',
  `zone`               varchar(32)      NOT NULL DEFAULT 'latrine',
  `shit_score`         decimal(10,4)    DEFAULT 0.0000,
  `avg_rating`         decimal(5,2)     DEFAULT 0.00,
  `rating_count`       int              DEFAULT 0,
  `view_count`         int              DEFAULT 0,
  `controversy_index`  decimal(5,4)     DEFAULT 0.0000,
  `weighted_avg_rating` decimal(5,2)   DEFAULT 0.00,
  `reviewer_authority` decimal(5,4)    DEFAULT 0.0000,
  `flag_count`         int              DEFAULT 0,
  `degradation_level`  tinyint          DEFAULT 0,
  `file_path`          varchar(512)     DEFAULT '',
  `doi`                varchar(128)     DEFAULT '',
  `keywords`           varchar(512)     DEFAULT '',
  `simhash`            bigint unsigned  NOT NULL DEFAULT 0,
  `status`             tinyint          NOT NULL DEFAULT 1,
  `promoted_at`        timestamp        NULL DEFAULT NULL,
  `last_accessed_at`   timestamp        NULL DEFAULT NULL,
  `created_at`         timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`         timestamp        DEFAULT CURRENT_TIMESTAMP,
  `archived_at`        timestamp        DEFAULT CURRENT_TIMESTAMP
                         COMMENT '归档时间',
  PRIMARY KEY (`id`),
  KEY `idx_author`   (`author_id`),
  KEY `idx_zone`     (`zone`),
  KEY `idx_archived` (`archived_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='冷数据归档表';

-- ---------------------------------------------------------------
-- 2.4 评分表
-- score: 1-10
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `rating` (
  `id`               bigint unsigned  NOT NULL AUTO_INCREMENT,
  `paper_id`         bigint unsigned  NOT NULL,
  `user_id`          bigint unsigned  NOT NULL,
  `score`            tinyint          NOT NULL COMMENT '1-10',
  `comment`          text             DEFAULT NULL,
  `reviewer_weight`  decimal(5,2)     DEFAULT 1.00
                       COMMENT '评审者权重(基于贡献分)',
  `source_ip`        varchar(64)      NOT NULL DEFAULT ''
                       COMMENT '评分请求来源 IP',
  `user_agent`       varchar(512)     NOT NULL DEFAULT ''
                       COMMENT '评分请求 User-Agent',
  `device_fingerprint` char(64)       NOT NULL DEFAULT ''
                       COMMENT 'IP + User-Agent 的 SHA-256 指纹',
  `created_at`       timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       timestamp        DEFAULT CURRENT_TIMESTAMP
                       ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user` (`paper_id`, `user_id`),
  KEY `idx_paper` (`paper_id`),
  KEY `idx_user`  (`user_id`),
  KEY `idx_paper_source_ip` (`paper_id`, `source_ip`, `updated_at`),
  KEY `idx_paper_device_fingerprint` (`paper_id`, `device_fingerprint`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='论文评分表';

-- ---------------------------------------------------------------
-- 2.5 新闻/公告表
-- category: announcement, governance, maintenance, feature
-- status: 0=draft, 1=published
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `news` (
  `id`         bigint unsigned  NOT NULL AUTO_INCREMENT,
  `title`      varchar(256)     NOT NULL,
  `title_en`   varchar(256)     DEFAULT '',
  `content`    longtext         NOT NULL,
  `content_en` longtext         DEFAULT NULL,
  `author_id`  bigint unsigned  NOT NULL,
  `category`   varchar(32)      DEFAULT 'announcement'
                 COMMENT 'announcement,governance,maintenance,feature',
  `is_pinned`  tinyint          DEFAULT 0,
  `status`     tinyint          DEFAULT 1
                 COMMENT '0=draft,1=published',
  `created_at` timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp        DEFAULT CURRENT_TIMESTAMP
                 ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_created` (`status`, `created_at`),
  KEY `idx_category`       (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='站内新闻/公告表';

-- ---------------------------------------------------------------
-- 2.6 举报表
-- target_type: paper, rating, user
-- reason: abuse, spam, plagiarism, sensitive, manipulation
-- status: 0=pending, 1=resolved_degraded, 2=resolved_dismissed
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `flag` (
  `id`                    bigint unsigned  NOT NULL AUTO_INCREMENT,
  `target_type`           varchar(16)      NOT NULL
                            COMMENT 'paper/rating/user',
  `target_id`             bigint unsigned  NOT NULL,
  `reporter_id`           bigint unsigned  NOT NULL,
  `reason`                varchar(32)      NOT NULL
                            COMMENT 'abuse/spam/plagiarism/sensitive/manipulation',
  `detail`                text             DEFAULT NULL,
  `reporter_contribution` decimal(10,2)    DEFAULT 0.00
                            COMMENT '举报时记录举报者贡献分，用于加权',
  `status`                tinyint          DEFAULT 0
                            COMMENT '0=pending,1=resolved_degraded,2=resolved_dismissed',
  `created_at`            timestamp        DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_target_reporter` (`target_type`, `target_id`, `reporter_id`),
  KEY `idx_target`  (`target_type`, `target_id`),
  KEY `idx_status`  (`status`),
  KEY `idx_reporter` (`reporter_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='举报记录表';

-- ---------------------------------------------------------------
-- 2.7 关键词黑名单规则表
-- match_type: keyword, regex, pinyin
-- category: abuse, sensitive, spam, plagiarism, manipulation
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `keyword_rule` (
  `id`               bigint unsigned NOT NULL AUTO_INCREMENT,
  `pattern`          varchar(255)    NOT NULL,
  `match_type`       varchar(16)     NOT NULL DEFAULT 'keyword'
                       COMMENT 'keyword/regex/pinyin',
  `category`         varchar(32)     NOT NULL DEFAULT 'spam'
                       COMMENT 'abuse/sensitive/spam/plagiarism/manipulation',
  `enabled`          tinyint         NOT NULL DEFAULT 1,
  `creator_user_id`  bigint unsigned NOT NULL DEFAULT 0,
  `created_at`       timestamp       DEFAULT CURRENT_TIMESTAMP,
  `updated_at`       timestamp       DEFAULT CURRENT_TIMESTAMP
                       ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_match_type_pattern` (`match_type`, `pattern`),
  KEY `idx_enabled` (`enabled`),
  KEY `idx_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='关键词黑名单规则表';


-- ===============================================================
-- 3. 管理库 journal_admin  —  RBAC
-- ===============================================================
USE `journal_admin`;

-- ---------------------------------------------------------------
-- 3.1 角色表
-- is_super=1 的角色拥有所有权限，无需在 adm_role_permission 分配
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_role` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `code`        varchar(64)      NOT NULL        COMMENT '角色唯一编码',
  `name`        varchar(64)      NOT NULL,
  `description` varchar(255)     DEFAULT '',
  `is_super`    tinyint          NOT NULL DEFAULT 0
                  COMMENT '1=超级管理员，绕过所有权限检查，不可修改/删除',
  `status`      tinyint          NOT NULL DEFAULT 1
                  COMMENT '0=disabled,1=active',
  `created_at`  timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  timestamp        DEFAULT CURRENT_TIMESTAMP
                  ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='管理后台角色表';

-- ---------------------------------------------------------------
-- 3.xxx 管理后台用户表
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_user` (
  `id`            bigint unsigned  NOT NULL AUTO_INCREMENT,
  `username`      varchar(64)      NOT NULL,
  `password_hash` varchar(256)     NOT NULL,
  `nickname`      varchar(64)      DEFAULT '',
  `avatar`        varchar(512)     DEFAULT '',
  `status`        tinyint          NOT NULL DEFAULT 1
                    COMMENT '0=disabled,1=active',
  `created_at`    timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    timestamp        DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='管理后台管理员表';

-- ---------------------------------------------------------------
-- 3.2 权限表
-- code 格式：admin.<module>.<action>
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_permission` (
  `id`          bigint unsigned  NOT NULL AUTO_INCREMENT,
  `code`        varchar(128)     NOT NULL        COMMENT '权限唯一编码',
  `name`        varchar(128)     NOT NULL,
  `module`      varchar(64)      NOT NULL        COMMENT '所属模块',
  `resource`    varchar(64)      NOT NULL        COMMENT '资源对象',
  `action`      varchar(64)      NOT NULL        COMMENT '操作动作',
  `description` varchar(255)     DEFAULT '',
  `status`      tinyint          NOT NULL DEFAULT 1
                  COMMENT '0=disabled,1=active',
  `created_at`  timestamp        DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  timestamp        DEFAULT CURRENT_TIMESTAMP
                  ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permission_code`    (`code`),
  KEY `idx_permission_module`        (`module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='管理后台权限表';

-- ---------------------------------------------------------------
-- 3.3 角色-权限关联表（M:N）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_role_permission` (
  `id`            bigint unsigned  NOT NULL AUTO_INCREMENT,
  `role_id`       bigint unsigned  NOT NULL,
  `permission_id` bigint unsigned  NOT NULL,
  `created_at`    timestamp        DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permission`  (`role_id`, `permission_id`),
  KEY `idx_role_id`                (`role_id`),
  KEY `idx_permission_id`          (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='角色-权限关联表';

-- ---------------------------------------------------------------
-- 3.4 用户-角色关联表（M:N，user_id 来自 journal_admin.adm_user）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_user_role` (
  `id`         bigint unsigned  NOT NULL AUTO_INCREMENT,
  `user_id`    bigint unsigned  NOT NULL  COMMENT '对应 journal_admin.adm_user.id',
  `role_id`    bigint unsigned  NOT NULL,
  `created_at` timestamp        DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role`  (`user_id`, `role_id`),
  KEY `idx_user_id`          (`user_id`),
  KEY `idx_role_id`          (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户-角色关联表';

-- ---------------------------------------------------------------
-- 3.5 操作审计日志表（只追加）
-- ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `adm_audit_log` (
  `id`              bigint unsigned  NOT NULL AUTO_INCREMENT,
  `actor_user_id`   bigint unsigned  NOT NULL    COMMENT '操作人 user_id',
  `permission_code` varchar(128)     NOT NULL    COMMENT '使用的权限码',
  `action`          varchar(128)     NOT NULL    COMMENT '操作描述',
  `target_type`     varchar(64)      DEFAULT ''  COMMENT '操作对象类型',
  `target_id`       bigint unsigned  DEFAULT NULL COMMENT '操作对象 ID',
  `detail`          text             DEFAULT NULL COMMENT '附加详情（JSON）',
  `created_at`      timestamp        DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_actor_created`  (`actor_user_id`, `created_at`),
  KEY `idx_permission`     (`permission_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='管理后台操作审计日志（只追加）';


-- ===============================================================
-- 4. DML — 种子数据
-- ===============================================================

-- ---------------------------------------------------------------
-- 4.1 初始化内置超管角色
--     ON DUPLICATE KEY 保证重复执行幂等
-- ---------------------------------------------------------------
INSERT INTO `journal_admin`.`adm_role`
  (`code`, `name`, `description`, `is_super`, `status`)
VALUES
  ('super_admin', '超级管理员', '内置超管角色，拥有所有权限，不可删除/修改', 1, 1)
ON DUPLICATE KEY UPDATE
  `name`        = VALUES(`name`),
  `description` = VALUES(`description`),
  `is_super`    = VALUES(`is_super`),
  `status`      = VALUES(`status`),
  `updated_at`  = CURRENT_TIMESTAMP;

-- ---------------------------------------------------------------
-- 4.2 初始化 "内容编辑" 角色（普通角色示例）
-- ---------------------------------------------------------------
INSERT INTO `journal_admin`.`adm_role`
  (`code`, `name`, `description`, `is_super`, `status`)
VALUES
  ('content_editor', '内容编辑', '负责新闻和论文区域管理', 0, 1)
ON DUPLICATE KEY UPDATE
  `name`        = VALUES(`name`),
  `description` = VALUES(`description`),
  `status`      = VALUES(`status`),
  `updated_at`  = CURRENT_TIMESTAMP;

-- ---------------------------------------------------------------
-- 4.3 初始化全量权限
-- ---------------------------------------------------------------
INSERT INTO `journal_admin`.`adm_permission`
  (`code`, `name`, `module`, `resource`, `action`, `description`, `status`)
VALUES
  -- Dashboard
  ('admin.dashboard.view',    '查看仪表盘',         'dashboard', 'dashboard', 'view',        '允许查看管理后台仪表盘',           1),
  -- Keyword Rule
  ('admin.keyword.view',      '查看关键词规则',     'keyword',   'keyword',   'view',        '允许查看关键词黑名单规则',         1),
  ('admin.keyword.manage',    '管理关键词规则',     'keyword',   'keyword',   'manage',      '允许新增和删除关键词黑名单规则',   1),
  -- News
  ('admin.news.view',         '查看新闻列表',        'news',      'news',      'view',        '允许查看后台新闻数据',             1),
  ('admin.news.create',       '创建新闻',            'news',      'news',      'create',      '允许在后台发布新闻',               1),
  ('admin.news.update',       '更新新闻',            'news',      'news',      'update',      '允许编辑已发布的新闻',             1),
  ('admin.news.delete',       '删除新闻',            'news',      'news',      'delete',      '允许删除新闻',                     1),
  -- Paper
  ('admin.paper.view',        '查看论文列表',        'paper',     'paper',     'view',        '允许查看论文管理数据',             1),
  ('admin.paper.zone.update', '更新论文分区',        'paper',     'paper',     'zone.update', '允许修改论文所在分区',             1),
  -- User
  ('admin.user.view',         '查看用户列表',        'user',      'user',      'view',        '允许查看用户管理数据',             1),
  ('admin.user.manage',       '管理用户',            'user',      'user',      'manage',      '允许封禁/调整用户角色',            1),
  -- RBAC
  ('admin.role.view',         '查看角色/权限',       'role',      'role',      'view',        '允许查看 RBAC 角色和权限列表',     1),
  ('admin.role.manage',       '管理角色/权限',       'role',      'role',      'manage',      '允许增删改角色、分配权限和用户角色', 1),
  -- Audit
  ('admin.audit.view',        '查看审计日志',        'audit',     'audit',     'view',        '允许查看操作审计日志',             1)
ON DUPLICATE KEY UPDATE
  `name`        = VALUES(`name`),
  `module`      = VALUES(`module`),
  `resource`    = VALUES(`resource`),
  `action`      = VALUES(`action`),
  `description` = VALUES(`description`),
  `status`      = VALUES(`status`),
  `updated_at`  = CURRENT_TIMESTAMP;

-- ---------------------------------------------------------------
-- 4.4 超管角色绑定所有权限
--     INSERT IGNORE 幂等，重复执行安全
-- ---------------------------------------------------------------
INSERT IGNORE INTO `journal_admin`.`adm_role_permission` (`role_id`, `permission_id`)
SELECT r.`id`, p.`id`
FROM   `journal_admin`.`adm_role`       r
CROSS JOIN `journal_admin`.`adm_permission` p
WHERE  r.`code` = 'super_admin';

-- ---------------------------------------------------------------
-- 4.5 内容编辑角色绑定对应权限（news/paper 只读+操作，但无 RBAC/审计权限）
-- ---------------------------------------------------------------
INSERT IGNORE INTO `journal_admin`.`adm_role_permission` (`role_id`, `permission_id`)
SELECT r.`id`, p.`id`
FROM   `journal_admin`.`adm_role`       r
JOIN   `journal_admin`.`adm_permission` p
       ON p.`code` IN (
         'admin.dashboard.view',
         'admin.news.view', 'admin.news.create', 'admin.news.update',
         'admin.paper.view', 'admin.paper.zone.update',
         'admin.user.view'
       )
WHERE  r.`code` = 'content_editor';

-- ---------------------------------------------------------------
-- 4.5.1 初始化默认关键词规则
-- ---------------------------------------------------------------
INSERT INTO `journal_biz`.`keyword_rule`
  (`pattern`, `match_type`, `category`, `enabled`, `creator_user_id`)
VALUES
  ('click here to buy', 'keyword', 'spam', 1, 0),
  ('free download', 'keyword', 'spam', 1, 0),
  ('limited time offer', 'keyword', 'spam', 1, 0)
ON DUPLICATE KEY UPDATE
  `category`   = VALUES(`category`),
  `enabled`    = VALUES(`enabled`),
  `updated_at` = CURRENT_TIMESTAMP;

-- ---------------------------------------------------------------
-- 4.6 测试账号（密码均为 Admin123!，已用 bcrypt 加密）
--     仅用于本地/测试环境，生产环境请删除
-- ---------------------------------------------------------------
INSERT INTO `journal_admin`.`adm_user`
  (`username`, `password_hash`, `nickname`, `status`)
VALUES
  -- 超管账号
  ('admin',
   '$2a$10$7EqJtq98hPqEX7fNZaFWoO5FK4F1p5atAZyHMKb1UNBMOHPbNHfuC',
   'System Admin', 1),
  -- 内容编辑示例账号
  ('editor_demo',
   '$2a$10$7EqJtq98hPqEX7fNZaFWoO5FK4F1p5atAZyHMKb1UNBMOHPbNHfuC',
   'Editor Demo', 1)
ON DUPLICATE KEY UPDATE
  `nickname`  = VALUES(`nickname`),
  `status`    = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;

-- 给 admin 账号绑定超管角色
INSERT IGNORE INTO `journal_admin`.`adm_user_role` (`user_id`, `role_id`)
SELECT u.`id`, r.`id`
FROM   `journal_admin`.`adm_user` u
JOIN   `journal_admin`.`adm_role` r ON r.`code` = 'super_admin'
WHERE  u.`username` = 'admin';

-- 给 editor_demo 账号绑定内容编辑角色
INSERT IGNORE INTO `journal_admin`.`adm_user_role` (`user_id`, `role_id`)
SELECT u.`id`, r.`id`
FROM   `journal_admin`.`adm_user` u
JOIN   `journal_admin`.`adm_role` r ON r.`code` = 'content_editor'
WHERE  u.`username` = 'editor_demo';

-- =============================================================
-- 初始化完成
-- =============================================================
SELECT '✅ DDL + DML 初始化完成' AS `status`;
