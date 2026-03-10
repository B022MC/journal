ALTER TABLE `paper`
  ADD COLUMN `simhash` bigint unsigned NOT NULL DEFAULT 0 COMMENT '内容 SimHash 指纹' AFTER `keywords`,
  ADD KEY `idx_simhash` (`simhash`);

ALTER TABLE `cold_paper`
  ADD COLUMN `simhash` bigint unsigned NOT NULL DEFAULT 0 AFTER `keywords`,
  ADD KEY `idx_simhash` (`simhash`);
