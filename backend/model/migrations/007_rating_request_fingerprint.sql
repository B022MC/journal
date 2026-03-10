USE `journal_biz`;

ALTER TABLE `rating`
  ADD COLUMN `source_ip` varchar(64) NOT NULL DEFAULT ''
    COMMENT '评分请求来源 IP' AFTER `reviewer_weight`,
  ADD COLUMN `user_agent` varchar(512) NOT NULL DEFAULT ''
    COMMENT '评分请求 User-Agent' AFTER `source_ip`,
  ADD COLUMN `device_fingerprint` char(64) NOT NULL DEFAULT ''
    COMMENT 'IP + User-Agent 的 SHA-256 指纹' AFTER `user_agent`,
  ADD KEY `idx_paper_source_ip` (`paper_id`, `source_ip`, `updated_at`),
  ADD KEY `idx_paper_device_fingerprint` (`paper_id`, `device_fingerprint`, `updated_at`);
