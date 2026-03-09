-- 003_cold_hot_separation.sql
-- 冷热数据分离：论文表新增 last_accessed_at + 冷数据归档表

-- 论文表新增最后访问时间字段
ALTER TABLE `paper`
  ADD COLUMN `last_accessed_at` timestamp NULL DEFAULT NULL COMMENT '最后被访问时间（浏览/评分/搜索命中）' AFTER `promoted_at`,
  ADD KEY `idx_last_accessed` (`last_accessed_at`);

-- 冷数据归档表（结构与 paper 表一致）
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
  COMMENT='冷数据归档表：>90天未访问且zone=sediment或status=0的论文';
