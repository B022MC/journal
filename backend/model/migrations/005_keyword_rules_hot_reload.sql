USE `journal_biz`;

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

INSERT INTO `keyword_rule` (`pattern`, `match_type`, `category`, `enabled`, `creator_user_id`)
VALUES
  ('click here to buy', 'keyword', 'spam', 1, 0),
  ('free download', 'keyword', 'spam', 1, 0),
  ('limited time offer', 'keyword', 'spam', 1, 0)
ON DUPLICATE KEY UPDATE
  `category` = VALUES(`category`),
  `enabled` = VALUES(`enabled`),
  `updated_at` = CURRENT_TIMESTAMP;

USE `journal_admin`;

INSERT INTO `adm_permission` (`code`, `name`, `module`, `resource`, `action`, `description`, `status`)
VALUES
  ('admin.keyword.view', 'View keyword rules', 'keyword', 'keyword', 'view', 'Allow viewing keyword blacklist rules', 1),
  ('admin.keyword.manage', 'Manage keyword rules', 'keyword', 'keyword', 'manage', 'Allow managing keyword blacklist rules', 1)
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
JOIN `adm_permission` p ON p.`code` IN ('admin.keyword.view', 'admin.keyword.manage')
WHERE r.`code` = 'super_admin';
