-- Rollback: Change ow_levels.level back to ow_levels.weight

ALTER TABLE `ow_levels` CHANGE COLUMN `level` `weight` INT(11) NOT NULL DEFAULT 0 COMMENT 'Weight';
