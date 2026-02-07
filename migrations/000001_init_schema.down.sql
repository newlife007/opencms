-- Rollback migration for OpenWan Database Schema
-- Drops all tables in reverse order of dependencies

DROP TABLE IF EXISTS `cs_counter`;
DROP TABLE IF EXISTS `ow_files_counter`;
DROP TABLE IF EXISTS `ow_roles_has_permissions`;
DROP TABLE IF EXISTS `ow_groups_has_roles`;
DROP TABLE IF EXISTS `ow_groups_has_category`;
DROP TABLE IF EXISTS `ow_levels`;
DROP TABLE IF EXISTS `ow_permissions`;
DROP TABLE IF EXISTS `ow_roles`;
DROP TABLE IF EXISTS `ow_groups`;
DROP TABLE IF EXISTS `ow_users`;
DROP TABLE IF EXISTS `ow_category`;
DROP TABLE IF EXISTS `ow_catalog`;
DROP TABLE IF EXISTS `ow_files`;
