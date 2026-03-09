-- S.H.I.T Journal Governance Modules Migration
-- 四大自治模块数据库变更

-- 1. 新表：flag（举报记录）
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

-- 2. paper 表新增治理字段
ALTER TABLE `paper`
  ADD COLUMN `flag_count` int DEFAULT 0 COMMENT '被举报次数',
  ADD COLUMN `degradation_level` tinyint DEFAULT 0 COMMENT '0=normal,1=watched,2=throttled,3=sealed',
  ADD COLUMN `weighted_avg_rating` decimal(5,2) DEFAULT 0.00 COMMENT '加权平均评分',
  ADD COLUMN `reviewer_authority` decimal(5,4) DEFAULT 0.0000 COMMENT '评审者权威度';

-- 3. user 表新增治理字段
ALTER TABLE `user`
  ADD COLUMN `last_active_at` timestamp NULL DEFAULT NULL COMMENT '最后活跃时间',
  ADD COLUMN `review_count_30d` int DEFAULT 0 COMMENT '近30天评审数';

-- 4. rating 表新增权重字段
ALTER TABLE `rating`
  ADD COLUMN `reviewer_weight` decimal(5,2) DEFAULT 1.00 COMMENT '评审者权重(基于贡献分)';
