CREATE DATABASE IF NOT EXISTS `journal_biz` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `journal_admin` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'journal'@'%' IDENTIFIED BY 'journal123';
GRANT ALL PRIVILEGES ON `journal_biz`.* TO 'journal'@'%';
GRANT ALL PRIVILEGES ON `journal_admin`.* TO 'journal'@'%';
FLUSH PRIVILEGES;

USE `journal_biz`;

-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `email` varchar(128) NOT NULL,
  `password_hash` varchar(256) NOT NULL,
  `nickname` varchar(64) DEFAULT '',
  `avatar` varchar(512) DEFAULT '',
  `role` tinyint NOT NULL DEFAULT 0 COMMENT '0=member, 1=scooper, 2=editor, 3=admin',
  `contribution_score` decimal(10,2) DEFAULT 0.00,
  `last_active_at` timestamp NULL DEFAULT NULL COMMENT '最后活跃时间',
  `review_count_30d` int DEFAULT 0 COMMENT '近30天评审数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=banned, 1=active',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 论文表
CREATE TABLE IF NOT EXISTS `paper` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(512) NOT NULL,
  `title_en` varchar(512) DEFAULT '',
  `abstract` text NOT NULL,
  `abstract_en` text DEFAULT NULL,
  `content` longtext NOT NULL,
  `author_id` bigint unsigned NOT NULL,
  `author_name` varchar(64) NOT NULL,
  `discipline` varchar(32) NOT NULL DEFAULT 'other' COMMENT 'science,humanities,information,technology,other',
  `zone` varchar(32) NOT NULL DEFAULT 'latrine' COMMENT 'latrine,septic_tank,stone,sediment',
  `shit_score` decimal(10,4) DEFAULT 0.0000 COMMENT 'S.H.I.T composite score',
  `avg_rating` decimal(5,2) DEFAULT 0.00,
  `rating_count` int DEFAULT 0,
  `view_count` int DEFAULT 0,
  `controversy_index` decimal(5,4) DEFAULT 0.0000 COMMENT '评分标准差/争议度',
  `weighted_avg_rating` decimal(5,2) DEFAULT 0.00 COMMENT '加权平均评分',
  `reviewer_authority` decimal(5,4) DEFAULT 0.0000 COMMENT '评审者权威度',
  `flag_count` int DEFAULT 0 COMMENT '被举报次数',
  `degradation_level` tinyint DEFAULT 0 COMMENT '0=normal,1=watched,2=throttled,3=sealed',
  `file_path` varchar(512) DEFAULT '',
  `doi` varchar(128) DEFAULT '',
  `keywords` varchar(512) DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=deleted, 1=active, 2=flagged',
  `promoted_at` timestamp NULL DEFAULT NULL,
  `last_accessed_at` timestamp NULL DEFAULT NULL COMMENT '最后被访问时间',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_zone` (`zone`),
  KEY `idx_discipline` (`discipline`),
  KEY `idx_author` (`author_id`),
  KEY `idx_shit_score` (`shit_score`),
  KEY `idx_zone_discipline` (`zone`, `discipline`),
  KEY `idx_created` (`created_at`),
  KEY `idx_last_accessed` (`last_accessed_at`),
  FULLTEXT KEY `ft_title_abstract` (`title`, `abstract`, `keywords`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 冷数据归档表
CREATE TABLE IF NOT EXISTS `cold_paper` (
  `id` bigint unsigned NOT NULL,
  `title` varchar(512) NOT NULL,
  `title_en` varchar(512) DEFAULT '',
  `abstract` text NOT NULL,
  `abstract_en` text DEFAULT NULL,
  `content` longtext NOT NULL,
  `author_id` bigint unsigned NOT NULL,
  `author_name` varchar(64) NOT NULL,
  `discipline` varchar(32) NOT NULL DEFAULT 'other',
  `zone` varchar(32) NOT NULL DEFAULT 'latrine',
  `shit_score` decimal(10,4) DEFAULT 0.0000,
  `avg_rating` decimal(5,2) DEFAULT 0.00,
  `rating_count` int DEFAULT 0,
  `view_count` int DEFAULT 0,
  `controversy_index` decimal(5,4) DEFAULT 0.0000,
  `weighted_avg_rating` decimal(5,2) DEFAULT 0.00,
  `reviewer_authority` decimal(5,4) DEFAULT 0.0000,
  `flag_count` int DEFAULT 0,
  `degradation_level` tinyint DEFAULT 0,
  `file_path` varchar(512) DEFAULT '',
  `doi` varchar(128) DEFAULT '',
  `keywords` varchar(512) DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1,
  `promoted_at` timestamp NULL DEFAULT NULL,
  `last_accessed_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `archived_at` timestamp DEFAULT CURRENT_TIMESTAMP COMMENT '归档时间',
  PRIMARY KEY (`id`),
  KEY `idx_author` (`author_id`),
  KEY `idx_zone` (`zone`),
  KEY `idx_archived` (`archived_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='冷数据归档表';

-- 评分表
CREATE TABLE IF NOT EXISTS `rating` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `paper_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `score` tinyint NOT NULL COMMENT '1-10',
  `comment` text DEFAULT NULL,
  `reviewer_weight` decimal(5,2) DEFAULT 1.00 COMMENT '评审者权重(基于贡献分)',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_paper_user` (`paper_id`, `user_id`),
  KEY `idx_paper` (`paper_id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 新闻表
CREATE TABLE IF NOT EXISTS `news` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(256) NOT NULL,
  `title_en` varchar(256) DEFAULT '',
  `content` longtext NOT NULL,
  `content_en` longtext DEFAULT NULL,
  `author_id` bigint unsigned NOT NULL,
  `category` varchar(32) DEFAULT 'announcement' COMMENT 'announcement,governance,maintenance,feature',
  `is_pinned` tinyint DEFAULT 0,
  `status` tinyint DEFAULT 1 COMMENT '0=draft, 1=published',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_created` (`status`, `created_at`),
  KEY `idx_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 举报表
CREATE TABLE IF NOT EXISTS `flag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `target_type` varchar(16) NOT NULL COMMENT 'paper/rating/user',
  `target_id` bigint unsigned NOT NULL,
  `reporter_id` bigint unsigned NOT NULL,
  `reason` varchar(32) NOT NULL COMMENT 'abuse/spam/plagiarism/sensitive/manipulation',
  `detail` text,
  `reporter_contribution` decimal(10,2) DEFAULT 0.00,
  `status` tinyint DEFAULT 0 COMMENT '0=pending, 1=resolved_degraded, 2=resolved_dismissed',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_target_reporter` (`target_type`, `target_id`, `reporter_id`),
  KEY `idx_target` (`target_type`, `target_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

USE `journal_admin`;

CREATE TABLE IF NOT EXISTS `adm_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) DEFAULT '',
  `is_super` tinyint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=disabled,1=active',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `adm_permission` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(128) NOT NULL,
  `name` varchar(128) NOT NULL,
  `module` varchar(64) NOT NULL,
  `resource` varchar(64) NOT NULL,
  `action` varchar(64) NOT NULL,
  `description` varchar(255) DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=disabled,1=active',
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permission_code` (`code`),
  KEY `idx_permission_module` (`module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `adm_user_role` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `adm_role_permission` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL,
  `permission_id` bigint unsigned NOT NULL,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `adm_audit_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `actor_user_id` bigint unsigned NOT NULL,
  `permission_code` varchar(128) NOT NULL,
  `action` varchar(128) NOT NULL,
  `target_type` varchar(64) DEFAULT '',
  `target_id` bigint unsigned DEFAULT NULL,
  `detail` text,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_actor_created` (`actor_user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `adm_role` (`code`, `name`, `description`, `is_super`, `status`)
VALUES ('super_admin', 'Super Admin', 'Built-in super administrator role', 1, 1)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `is_super` = VALUES(`is_super`),
  `status` = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;

INSERT INTO `adm_permission` (`code`, `name`, `module`, `resource`, `action`, `description`, `status`)
VALUES
  ('admin.dashboard.view', 'View dashboard', 'dashboard', 'dashboard', 'view', 'Allow viewing the admin dashboard', 1),
  ('admin.news.view', 'View news', 'news', 'news', 'view', 'Allow viewing admin news data', 1),
  ('admin.news.create', 'Create news', 'news', 'news', 'create', 'Allow creating news', 1),
  ('admin.news.update', 'Update news', 'news', 'news', 'update', 'Allow updating news', 1),
  ('admin.news.delete', 'Delete news', 'news', 'news', 'delete', 'Allow deleting news', 1),
  ('admin.paper.view', 'View papers', 'paper', 'paper', 'view', 'Allow viewing paper management data', 1),
  ('admin.paper.zone.update', 'Update paper zone', 'paper', 'paper', 'zone.update', 'Allow updating paper zone', 1),
  ('admin.user.view', 'View users', 'user', 'user', 'view', 'Allow viewing user management data', 1),
  ('admin.user.manage', 'Manage users', 'user', 'user', 'manage', 'Allow managing users', 1),
  ('admin.role.view', 'View roles', 'role', 'role', 'view', 'Allow viewing RBAC roles and permissions', 1),
  ('admin.role.manage', 'Manage roles', 'role', 'role', 'manage', 'Allow managing RBAC roles and permissions', 1),
  ('admin.audit.view', 'View audit logs', 'audit', 'audit', 'view', 'Allow viewing admin audit logs', 1)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `module` = VALUES(`module`),
  `resource` = VALUES(`resource`),
  `action` = VALUES(`action`),
  `description` = VALUES(`description`),
  `status` = VALUES(`status`),
  `updated_at` = CURRENT_TIMESTAMP;

INSERT IGNORE INTO `adm_role_permission` (`role_id`, `permission_id`)
SELECT r.`id`, p.`id`
FROM `adm_role` r
CROSS JOIN `adm_permission` p
WHERE r.`code` = 'super_admin';

INSERT IGNORE INTO `adm_user_role` (`user_id`, `role_id`)
SELECT u.`id`, r.`id`
FROM `journal_biz`.`user` u
JOIN `adm_role` r ON r.`code` = 'super_admin'
WHERE u.`role` = 3;
