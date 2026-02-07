-- OpenWan Database Schema Migration
-- Creates all tables with ow_ prefix

-- Files table: stores media file information
CREATE TABLE IF NOT EXISTS `ow_files` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `category_id` int(11) NOT NULL COMMENT 'Category ID',
  `category_name` varchar(64) NOT NULL COMMENT 'Category name',
  `type` int(11) NOT NULL DEFAULT '1' COMMENT 'File type (1:video 2:audio 3:image 4:rich_media)',
  `title` varchar(255) NOT NULL COMMENT 'Display title',
  `name` varchar(255) NOT NULL COMMENT 'Filename (MD5)',
  `ext` varchar(16) NOT NULL COMMENT 'File extension',
  `size` bigint(20) NOT NULL DEFAULT '0' COMMENT 'File size in bytes',
  `path` varchar(255) NOT NULL COMMENT 'Storage path',
  `status` int(11) NOT NULL COMMENT 'Status (0:new 1:pending 2:published 3:rejected 4:deleted)',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Browsing level',
  `groups` varchar(255) NOT NULL DEFAULT 'all' COMMENT 'Accessible groups (comma-separated IDs or all)',
  `is_download` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Allow download',
  `catalog_info` text NOT NULL COMMENT 'Catalog metadata (JSON)',
  `upload_username` varchar(64) NOT NULL COMMENT 'Upload user',
  `upload_at` int(11) NOT NULL COMMENT 'Upload timestamp',
  `catalog_username` varchar(64) DEFAULT NULL COMMENT 'Catalog user',
  `catalog_at` int(11) DEFAULT NULL COMMENT 'Catalog timestamp',
  `putout_username` varchar(64) DEFAULT NULL COMMENT 'Publish user',
  `putout_at` int(11) DEFAULT NULL COMMENT 'Publish timestamp',
  PRIMARY KEY (`id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_type` (`type`),
  KEY `idx_title` (`title`),
  KEY `idx_status` (`status`),
  KEY `idx_level` (`level`),
  KEY `idx_upload_at` (`upload_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Media files table';

-- Catalog table: dynamic metadata configuration
CREATE TABLE IF NOT EXISTS `ow_catalog` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `parent_id` int(11) NOT NULL COMMENT 'Parent ID',
  `path` varchar(255) NOT NULL COMMENT 'Hierarchical path',
  `name` varchar(64) NOT NULL COMMENT 'Display name',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT 'Description',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)'for ordering',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  `created` int(11) NOT NULL COMMENT 'Created timestamp',
  `updated` int(11) NOT NULL COMMENT 'Updated timestamp',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Catalog metadata configuration';

-- Category table: hierarchical resource classification
CREATE TABLE IF NOT EXISTS `ow_category` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `parent_id` int(11) NOT NULL COMMENT 'Parent ID',
  `path` varchar(255) NOT NULL COMMENT 'Hierarchical path',
  `name` varchar(64) NOT NULL COMMENT 'Display name',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT 'Description',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)'for ordering',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  `created` int(11) NOT NULL COMMENT 'Created timestamp',
  `updated` int(11) NOT NULL COMMENT 'Updated timestamp',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Resource category table';

-- Users table: user authentication and profile
CREATE TABLE IF NOT EXISTS `ow_users` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `group_id` int(11) NOT NULL COMMENT 'Group ID',
  `level_id` int(11) NOT NULL COMMENT 'Reading level ID',
  `username` varchar(32) NOT NULL COMMENT 'Username',
  `password` varchar(64) NOT NULL COMMENT 'Password (hashed)',
  `nickname` varchar(64) NOT NULL COMMENT 'Nickname',
  `sex` tinyint(2) NOT NULL DEFAULT '0' COMMENT 'Gender (0:secret 1:male 2:female)',
  `birthday` varchar(64) DEFAULT NULL COMMENT 'Birthday',
  `address` varchar(255) DEFAULT NULL COMMENT 'Address',
  `email` varchar(64) DEFAULT NULL COMMENT 'Email',
  `duty` varchar(64) DEFAULT NULL COMMENT 'Duty',
  `office_phone` varchar(64) DEFAULT NULL COMMENT 'Office phone',
  `home_phone` varchar(64) DEFAULT NULL COMMENT 'Home phone',
  `mobile_phone` varchar(64) DEFAULT NULL COMMENT 'Mobile phone',
  `description` varchar(255) DEFAULT NULL COMMENT 'Personal introduction',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  `register_at` int(11) NOT NULL DEFAULT '0' COMMENT 'Register timestamp',
  `register_ip` char(15) NOT NULL DEFAULT '0.0.0.0' COMMENT 'Register IP',
  `login_count` int(11) NOT NULL DEFAULT '0' COMMENT 'Login count',
  `login_at` int(11) NOT NULL DEFAULT '0' COMMENT 'Last login timestamp',
  `login_ip` char(15) NOT NULL DEFAULT '0.0.0.0' COMMENT 'Last login IP',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_password` (`password`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Users table';

-- Groups table: user group management
CREATE TABLE IF NOT EXISTS `ow_groups` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(32) NOT NULL COMMENT 'Name',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT 'Description',
  `quota` int(11) NOT NULL DEFAULT '1000' COMMENT 'User disk quota (MB)',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User groups table';

-- Roles table: role-based access control
CREATE TABLE IF NOT EXISTS `ow_roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(32) NOT NULL COMMENT 'Name',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT 'Description',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Roles table';

-- Permissions table: access control permissions
CREATE TABLE IF NOT EXISTS `ow_permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `namespace` varchar(64) NOT NULL DEFAULT 'default' COMMENT 'Namespace',
  `controller` varchar(64) NOT NULL DEFAULT 'default' COMMENT 'Controller',
  `action` varchar(64) NOT NULL DEFAULT 'index' COMMENT 'Action',
  `aliasname` varchar(64) NOT NULL DEFAULT '' COMMENT 'Alias name',
  `rbac` varchar(32) NOT NULL DEFAULT 'ACL_NULL' COMMENT 'System role',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Permissions table';

-- Levels table: browsing level management
CREATE TABLE IF NOT EXISTS `ow_levels` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(64) NOT NULL COMMENT 'Level name',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT 'Description',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)',
  `enabled` tinyint(2) NOT NULL DEFAULT '1' COMMENT 'Enabled status',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Browsing levels table';

-- Junction table: Groups to Categories (many-to-many)
CREATE TABLE IF NOT EXISTS `ow_groups_has_category` (
  `group_id` int(11) NOT NULL,
  `category_id` int(11) NOT NULL,
  PRIMARY KEY (`group_id`, `category_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Groups to categories mapping';

-- Junction table: Groups to Roles (many-to-many)
CREATE TABLE IF NOT EXISTS `ow_groups_has_roles` (
  `group_id` int(11) NOT NULL,
  `role_id` int(11) NOT NULL,
  PRIMARY KEY (`group_id`, `role_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Groups to roles mapping';

-- Junction table: Roles to Permissions (many-to-many)
CREATE TABLE IF NOT EXISTS `ow_roles_has_permissions` (
  `role_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`role_id`, `permission_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Roles to permissions mapping';

-- Files counter table: for Sphinx search indexing
CREATE TABLE IF NOT EXISTS `ow_files_counter` (
  `id` int(11) NOT NULL COMMENT 'Counter ID',
  `file_id` int(11) NOT NULL COMMENT 'File ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Files counter for search indexing';

-- Sphinx counter table (cs_counter)
CREATE TABLE IF NOT EXISTS `cs_counter` (
  `id` int(11) NOT NULL COMMENT 'Counter ID',
  `maxid` int(11) NOT NULL COMMENT 'Max document ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Sphinx document counter';
