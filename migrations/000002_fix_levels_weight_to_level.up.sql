-- Migration: Change ow_levels.weight to ow_levels.level
-- This fixes the level management logic

-- Rename column from weight to level
ALTER TABLE `ow_levels` CHANGE COLUMN `weight` `level` INT(11) NOT NULL DEFAULT 1 COMMENT 'Level value (higher = more access)';

-- Update existing data if needed (assuming weight was 0-based, convert to 1-based)
-- UPDATE `ow_levels` SET `level` = `level` + 1 WHERE `level` = 0;

-- Update sample levels data
UPDATE `ow_levels` SET `level` = 1 WHERE `name` = 'Public';
UPDATE `ow_levels` SET `level` = 2 WHERE `name` = 'Internal';
UPDATE `ow_levels` SET `level` = 3 WHERE `name` = 'Confidential';
UPDATE `ow_levels` SET `level` = 4 WHERE `name` = 'Secret';
UPDATE `ow_levels` SET `level` = 5 WHERE `name` = 'Top Secret';
