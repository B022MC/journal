-- 004_split_biz_admin_rbac.sql
-- 将旧的单库 journal 迁移为同实例双库：
--   - journal_biz   业务库
--   - journal_admin 管理后台/RBAC 库

CREATE DATABASE IF NOT EXISTS `journal_biz` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS `journal_admin` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'journal'@'%' IDENTIFIED BY 'journal123';
GRANT ALL PRIVILEGES ON `journal_biz`.* TO 'journal'@'%';
GRANT ALL PRIVILEGES ON `journal_admin`.* TO 'journal'@'%';
FLUSH PRIVILEGES;

SET @move_user = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'user')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'user'),
    'RENAME TABLE `journal`.`user` TO `journal_biz`.`user`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_user;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @move_paper = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'paper')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'paper'),
    'RENAME TABLE `journal`.`paper` TO `journal_biz`.`paper`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_paper;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @move_cold_paper = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'cold_paper')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'cold_paper'),
    'RENAME TABLE `journal`.`cold_paper` TO `journal_biz`.`cold_paper`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_cold_paper;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @move_rating = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'rating')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'rating'),
    'RENAME TABLE `journal`.`rating` TO `journal_biz`.`rating`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_rating;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @move_news = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'news')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'news'),
    'RENAME TABLE `journal`.`news` TO `journal_biz`.`news`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_news;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @move_flag = (
  SELECT IF(
    EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal' AND table_name = 'flag')
    AND NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'journal_biz' AND table_name = 'flag'),
    'RENAME TABLE `journal`.`flag` TO `journal_biz`.`flag`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @move_flag;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

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
