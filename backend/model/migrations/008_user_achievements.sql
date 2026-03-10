USE `journal_biz`;

CREATE TABLE IF NOT EXISTS `user_achievement` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `code` varchar(64) NOT NULL COMMENT 'first_submission/sediment_breakthrough/reviewer_century',
  `unlocked_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_code` (`user_id`, `code`),
  KEY `idx_user_unlocked_at` (`user_id`, `unlocked_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='用户成就徽章解锁记录';

INSERT IGNORE INTO `user_achievement` (`user_id`, `code`)
SELECT `author_id`, 'first_submission'
FROM `paper`
WHERE `status` > 0;

INSERT IGNORE INTO `user_achievement` (`user_id`, `code`)
SELECT `author_id`, 'sediment_breakthrough'
FROM `paper`
WHERE `status` > 0 AND `zone` = 'sediment';

INSERT IGNORE INTO `user_achievement` (`user_id`, `code`)
SELECT `user_id`, 'reviewer_century'
FROM `rating`
GROUP BY `user_id`
HAVING COUNT(*) >= 100;
