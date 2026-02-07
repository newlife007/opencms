-- Add type field to ow_catalog table
ALTER TABLE `ow_catalog` 
ADD COLUMN `type` int(11) NOT NULL DEFAULT 0 COMMENT 'File type (1=video, 2=audio, 3=image, 4=rich)' AFTER `id`,
ADD INDEX `idx_type` (`type`);

-- Add label field for field label (display name)
ALTER TABLE `ow_catalog`
ADD COLUMN `label` varchar(64) NOT NULL DEFAULT '' COMMENT 'Display label' AFTER `name`;

-- Add field_type for input type (text, number, date, select, textarea)
ALTER TABLE `ow_catalog`
ADD COLUMN `field_type` varchar(32) NOT NULL DEFAULT 'text' COMMENT 'Field input type' AFTER `description`;

-- Add required flag
ALTER TABLE `ow_catalog`
ADD COLUMN `required` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'Is field required' AFTER `field_type`;

-- Add options field for select type
ALTER TABLE `ow_catalog`
ADD COLUMN `options` text COMMENT 'JSON options for select field' AFTER `required`;
