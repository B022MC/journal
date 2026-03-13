-- 009_merge_split_dbs_into_single_journal.sql
-- 将 journal_biz / journal_admin 合并回单库 journal，
-- 真实表统一命名为 biz_* / adm_*，旧业务表名通过兼容视图保留。

CREATE DATABASE IF NOT EXISTS `journal` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'journal'@'%' IDENTIFIED BY 'banishmentB022.';
GRANT ALL PRIVILEGES ON `journal`.* TO 'journal'@'%';
FLUSH PRIVILEGES;

CREATE TABLE IF NOT EXISTS `journal`.`biz_user` LIKE `journal_biz`.`user`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_user_achievement` LIKE `journal_biz`.`user_achievement`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_paper` LIKE `journal_biz`.`paper`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_cold_paper` LIKE `journal_biz`.`cold_paper`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_rating` LIKE `journal_biz`.`rating`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_news` LIKE `journal_biz`.`news`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_flag` LIKE `journal_biz`.`flag`;
CREATE TABLE IF NOT EXISTS `journal`.`biz_keyword_rule` LIKE `journal_biz`.`keyword_rule`;

CREATE TABLE IF NOT EXISTS `journal`.`adm_role` LIKE `journal_admin`.`adm_role`;
CREATE TABLE IF NOT EXISTS `journal`.`adm_permission` LIKE `journal_admin`.`adm_permission`;
CREATE TABLE IF NOT EXISTS `journal`.`adm_role_permission` LIKE `journal_admin`.`adm_role_permission`;
CREATE TABLE IF NOT EXISTS `journal`.`adm_user_role` LIKE `journal_admin`.`adm_user_role`;
CREATE TABLE IF NOT EXISTS `journal`.`adm_audit_log` LIKE `journal_admin`.`adm_audit_log`;

INSERT INTO `journal`.`biz_user`
SELECT * FROM `journal_biz`.`user`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_user_achievement`
SELECT * FROM `journal_biz`.`user_achievement`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_paper`
SELECT * FROM `journal_biz`.`paper`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_cold_paper`
SELECT * FROM `journal_biz`.`cold_paper`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_rating`
SELECT * FROM `journal_biz`.`rating`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_news`
SELECT * FROM `journal_biz`.`news`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_flag`
SELECT * FROM `journal_biz`.`flag`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`biz_keyword_rule`
SELECT * FROM `journal_biz`.`keyword_rule`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`adm_role`
SELECT * FROM `journal_admin`.`adm_role`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`adm_permission`
SELECT * FROM `journal_admin`.`adm_permission`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`adm_role_permission`
SELECT * FROM `journal_admin`.`adm_role_permission`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`adm_user_role`
SELECT * FROM `journal_admin`.`adm_user_role`
ON DUPLICATE KEY UPDATE `id` = `id`;

INSERT INTO `journal`.`adm_audit_log`
SELECT * FROM `journal_admin`.`adm_audit_log`
ON DUPLICATE KEY UPDATE `id` = `id`;

CREATE OR REPLACE VIEW `journal`.`user` AS
SELECT
  `id`,
  `username`,
  `email`,
  `password_hash`,
  `nickname`,
  `avatar`,
  `role`,
  `contribution_score`,
  `last_active_at`,
  `review_count_30d`,
  `status`,
  `created_at`,
  `updated_at`
FROM `journal`.`biz_user`;

CREATE OR REPLACE VIEW `journal`.`user_achievement` AS
SELECT
  `id`,
  `user_id`,
  `code`,
  `unlocked_at`
FROM `journal`.`biz_user_achievement`;

CREATE OR REPLACE VIEW `journal`.`paper` AS
SELECT
  `id`,
  `title`,
  `title_en`,
  `abstract`,
  `abstract_en`,
  `content`,
  `author_id`,
  `author_name`,
  `discipline`,
  `zone`,
  `shit_score`,
  `avg_rating`,
  `rating_count`,
  `view_count`,
  `controversy_index`,
  `weighted_avg_rating`,
  `reviewer_authority`,
  `flag_count`,
  `degradation_level`,
  `file_path`,
  `doi`,
  `keywords`,
  `simhash`,
  `status`,
  `promoted_at`,
  `last_accessed_at`,
  `created_at`,
  `updated_at`
FROM `journal`.`biz_paper`;

CREATE OR REPLACE VIEW `journal`.`cold_paper` AS
SELECT
  `id`,
  `title`,
  `title_en`,
  `abstract`,
  `abstract_en`,
  `content`,
  `author_id`,
  `author_name`,
  `discipline`,
  `zone`,
  `shit_score`,
  `avg_rating`,
  `rating_count`,
  `view_count`,
  `controversy_index`,
  `weighted_avg_rating`,
  `reviewer_authority`,
  `flag_count`,
  `degradation_level`,
  `file_path`,
  `doi`,
  `keywords`,
  `simhash`,
  `status`,
  `promoted_at`,
  `last_accessed_at`,
  `created_at`,
  `updated_at`,
  `archived_at`
FROM `journal`.`biz_cold_paper`;

CREATE OR REPLACE VIEW `journal`.`rating` AS
SELECT
  `id`,
  `paper_id`,
  `user_id`,
  `score`,
  `comment`,
  `reviewer_weight`,
  `source_ip`,
  `user_agent`,
  `device_fingerprint`,
  `created_at`,
  `updated_at`
FROM `journal`.`biz_rating`;

CREATE OR REPLACE VIEW `journal`.`news` AS
SELECT
  `id`,
  `title`,
  `title_en`,
  `content`,
  `content_en`,
  `author_id`,
  `category`,
  `is_pinned`,
  `status`,
  `created_at`,
  `updated_at`
FROM `journal`.`biz_news`;

CREATE OR REPLACE VIEW `journal`.`flag` AS
SELECT
  `id`,
  `target_type`,
  `target_id`,
  `reporter_id`,
  `reason`,
  `detail`,
  `reporter_contribution`,
  `status`,
  `created_at`
FROM `journal`.`biz_flag`;

CREATE OR REPLACE VIEW `journal`.`keyword_rule` AS
SELECT
  `id`,
  `pattern`,
  `match_type`,
  `category`,
  `enabled`,
  `creator_user_id`,
  `created_at`,
  `updated_at`
FROM `journal`.`biz_keyword_rule`;
