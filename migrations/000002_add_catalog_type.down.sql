-- Remove fields added in 000002_add_catalog_type.up.sql
ALTER TABLE `ow_catalog`
DROP COLUMN `options`,
DROP COLUMN `required`,
DROP COLUMN `field_type`,
DROP COLUMN `label`,
DROP INDEX `idx_type`,
DROP COLUMN `type`;
