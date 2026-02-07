# OpenWan 数据库设计文档

**数据库类型**: MySQL 5.7+  
**字符集**: UTF-8MB4  
**存储引擎**: InnoDB  
**表前缀**: `ow_`  
**版本**: v2.0  
**更新日期**: 2026-02-01

---

## 📋 目录

1. [数据库概述](#数据库概述)
2. [ER 图](#er-图)
3. [表结构详解](#表结构详解)
4. [索引设计](#索引设计)
5. [数据字典](#数据字典)
6. [迁移脚本](#迁移脚本)

---

## 数据库概述

### 设计原则

1. **规范化**: 遵循第三范式(3NF)，减少数据冗余
2. **性能优化**: 合理使用索引，避免过度索引
3. **扩展性**: 预留扩展字段，支持未来功能
4. **兼容性**: 兼容 MySQL 5.7+ 和 MariaDB 10.3+
5. **安全性**: 密码字段加密存储，敏感数据脱敏

### 数据库信息

```
数据库名: openwan_db
字符集: utf8mb4
排序规则: utf8mb4_unicode_ci
存储引擎: InnoDB
表数量: 13
```

### 表分类

**核心业务表** (3):
- `ow_files` - 媒体文件表
- `ow_catalog` - 编目元数据配置表
- `ow_category` - 资源分类表

**用户权限表** (5):
- `ow_users` - 用户表
- `ow_groups` - 用户组表
- `ow_roles` - 角色表
- `ow_permissions` - 权限表
- `ow_levels` - 浏览级别表

**关系映射表** (3):
- `ow_groups_has_category` - 组-分类关联表
- `ow_groups_has_roles` - 组-角色关联表
- `ow_roles_has_permissions` - 角色-权限关联表

**辅助表** (2):
- `ow_files_counter` - 文件计数表（用于 Sphinx）
- `cs_counter` - Sphinx 文档计数表

---

## ER 图

### 核心实体关系

```
┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   ow_users   │────┐    │  ow_groups   │────┐    │   ow_roles   │
│              │    │    │              │    │    │              │
│ - id         │    │    │ - id         │    │    │ - id         │
│ - username   │    │    │ - name       │    │    │ - name       │
│ - password   │    │    │ - level      │    │    │ - level      │
│ - group_id ──┼────┘    └──────────────┘    └────┼─ id          │
│ - level_id   │                  │                └──────────────┘
└──────────────┘                  │                       │
       │                          │                       │
       │                    ow_groups_has_roles          │
       │                    ┌────────────┐               │
       │                    │ group_id ──┼───────────────┘
       │                    │ role_id    │
       │                    └────────────┘
       │                          │
       ▼                          ▼
┌──────────────┐         ┌──────────────────┐
│  ow_levels   │         │ ow_roles_has_    │
│              │         │   permissions    │
│ - id         │         │                  │
│ - name       │         │ role_id          │
│ - level      │         │ permission_id ───┼───┐
└──────────────┘         └──────────────────┘   │
                                                 │
                         ┌──────────────────┐   │
                         │ ow_permissions   │◄──┘
                         │                  │
                         │ - id             │
                         │ - namespace      │
                         │ - controller     │
                         │ - action         │
                         └──────────────────┘

┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│   ow_files   │         │ ow_category  │         │  ow_catalog  │
│              │         │              │         │              │
│ - id         │         │ - id         │         │ - id         │
│ - title      │         │ - name       │         │ - name       │
│ - type       │         │ - parent_id  │         │ - parent_id  │
│ - category_id┼────────►│ - path       │         │ - path       │
│ - level      │         │ - level      │         │ - enabled    │
│ - groups     │         └──────────────┘         └──────────────┘
│ - catalog_info│                │
│ - status     │                 │
└──────────────┘                 │
                                 │
                        ow_groups_has_category
                        ┌────────────┐
                        │ group_id   │
                        │ category_id│
                        └────────────┘
```

### RBAC 权限模型

```
用户 (Users)
   │
   ├─► 所属组 (Groups)
   │      │
   │      ├─► 分配角色 (Roles)
   │      │      │
   │      │      └─► 拥有权限 (Permissions)
   │      │
   │      └─► 访问分类 (Categories)
   │
   └─► 浏览级别 (Levels)
```

---

## 表结构详解

### 1. ow_files - 媒体文件表

**用途**: 存储所有上传的媒体文件信息和元数据

**表结构**:

```sql
CREATE TABLE `ow_files` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `category_id` int(11) NOT NULL COMMENT 'Category ID',
  `category_name` varchar(64) NOT NULL COMMENT 'Category name',
  `type` int(11) NOT NULL DEFAULT '1' COMMENT 'File type',
  `title` varchar(255) NOT NULL COMMENT 'Display title',
  `name` varchar(255) NOT NULL COMMENT 'Filename (MD5)',
  `ext` varchar(16) NOT NULL COMMENT 'File extension',
  `size` bigint(20) NOT NULL DEFAULT '0' COMMENT 'File size (bytes)',
  `path` varchar(255) NOT NULL COMMENT 'Storage path',
  `status` int(11) NOT NULL COMMENT 'Status',
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Browsing level',
  `groups` varchar(255) NOT NULL DEFAULT 'all' COMMENT 'Accessible groups',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**字段说明**:

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | bigint(20) | 主键ID，自增 | 1000001 |
| category_id | int(11) | 所属分类ID | 10 |
| category_name | varchar(64) | 分类名称（冗余，性能优化） | "新闻视频" |
| type | int(11) | 文件类型 | 1=视频, 2=音频, 3=图片, 4=富媒体 |
| title | varchar(255) | 文件显示标题 | "2026春节晚会" |
| name | varchar(255) | 存储文件名(MD5) | "a3b5c7d9e1f2..." |
| ext | varchar(16) | 文件扩展名 | "mp4", "jpg" |
| size | bigint(20) | 文件大小(字节) | 1073741824 (1GB) |
| path | varchar(255) | 存储路径 | "/data1/a3/b5/file.mp4" |
| status | int(11) | 文件状态 | 0-4 (见下表) |
| level | int(11) | 浏览级别 | 1-10 |
| groups | varchar(255) | 可访问组 | "1,2,3" 或 "all" |
| is_download | tinyint(1) | 允许下载 | 0=否, 1=是 |
| catalog_info | text | 编目元数据(JSON) | `{"director":"张三",...}` |
| upload_username | varchar(64) | 上传用户 | "zhangsan" |
| upload_at | int(11) | 上传时间戳 | 1704038400 |
| catalog_username | varchar(64) | 编目用户 | "lisi" |
| catalog_at | int(11) | 编目时间戳 | 1704124800 |
| putout_username | varchar(64) | 发布用户 | "admin" |
| putout_at | int(11) | 发布时间戳 | 1704211200 |

**文件状态枚举**:

| 值 | 常量 | 说明 | 业务含义 |
|----|------|------|---------|
| 0 | STATUS_NEW | 新上传 | 刚上传，未编目 |
| 1 | STATUS_PENDING | 待审核 | 已编目，等待审核 |
| 2 | STATUS_PUBLISHED | 已发布 | 审核通过，对用户可见 |
| 3 | STATUS_REJECTED | 已拒绝 | 审核未通过 |
| 4 | STATUS_DELETED | 已删除 | 进入回收站 |

**文件类型枚举**:

| 值 | 常量 | 说明 | 支持格式 |
|----|------|------|---------|
| 1 | TYPE_VIDEO | 视频 | MP4, AVI, MOV, FLV, MKV |
| 2 | TYPE_AUDIO | 音频 | MP3, WAV, AAC, FLAC |
| 3 | TYPE_IMAGE | 图片 | JPG, PNG, GIF, BMP |
| 4 | TYPE_DOCUMENT | 富媒体 | PDF, DOC, XLS, PPT |

**索引说明**:

- `PRIMARY KEY (id)`: 主键索引，快速定位单条记录
- `idx_category_id`: 按分类查询
- `idx_type`: 按类型筛选
- `idx_title`: 按标题搜索（前缀匹配）
- `idx_status`: 按状态筛选（待审核、已发布等）
- `idx_level`: 权限控制查询
- `idx_upload_at`: 按时间排序

**数据示例**:

```sql
INSERT INTO ow_files VALUES (
  1,                          -- id
  10,                         -- category_id
  '新闻视频',                 -- category_name
  1,                          -- type (视频)
  '2026年春节联欢晚会',        -- title
  'a3b5c7d9e1f2...',          -- name (MD5)
  'mp4',                      -- ext
  1073741824,                 -- size (1GB)
  '/data1/a3/b5/video.mp4',   -- path
  2,                          -- status (已发布)
  1,                          -- level
  'all',                      -- groups
  1,                          -- is_download
  '{"director":"张艺谋","duration":"180分钟"}', -- catalog_info
  'zhangsan',                 -- upload_username
  1704038400,                 -- upload_at
  'lisi',                     -- catalog_username
  1704124800,                 -- catalog_at
  'admin',                    -- putout_username
  1704211200                  -- putout_at
);
```

---

### 2. ow_catalog - 编目元数据配置表

**用途**: 定义动态编目字段结构，支持按文件类型自定义元数据

**表结构**:

```sql
CREATE TABLE `ow_catalog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `parent_id` int(11) NOT NULL,
  `path` varchar(255) NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT '1',
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  `created` int(11) NOT NULL,
  `updated` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**字段说明**:

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | int(11) | 主键ID | 1 |
| parent_id | int(11) | 父字段ID，0表示根 | 0 |
| path | varchar(255) | 层级路径（逗号分隔） | "0,1,5" |
| name | varchar(64) | 字段名称 | "导演", "演员" |
| description | varchar(255) | 字段描述 | "影片导演姓名" |
| level | int(11) | 排序权重 | 1-999 |
| enabled | tinyint(2) | 启用状态 | 0=禁用, 1=启用 |
| created | int(11) | 创建时间戳 | 1704038400 |
| updated | int(11) | 更新时间戳 | 1704038400 |

**层级结构示例**:

```
视频元数据 (id=1, parent_id=0, path="0")
├─ 基本信息 (id=2, parent_id=1, path="0,1")
│  ├─ 导演 (id=3, parent_id=2, path="0,1,2")
│  ├─ 演员 (id=4, parent_id=2, path="0,1,2")
│  └─ 时长 (id=5, parent_id=2, path="0,1,2")
└─ 版权信息 (id=6, parent_id=1, path="0,1")
   ├─ 版权方 (id=7, parent_id=6, path="0,1,6")
   └─ 授权期限 (id=8, parent_id=6, path="0,1,6")
```

---

### 3. ow_category - 资源分类表

**用途**: 定义媒体文件的分类层级结构

**表结构**: 与 `ow_catalog` 结构相同，但用途不同

```sql
CREATE TABLE `ow_category` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `parent_id` int(11) NOT NULL,
  `path` varchar(255) NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT '1',
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  `created` int(11) NOT NULL,
  `updated` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_path` (`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**分类示例**:

```
全部 (id=1, parent_id=0)
├─ 新闻 (id=2, parent_id=1)
│  ├─ 时政新闻 (id=3, parent_id=2)
│  └─ 社会新闻 (id=4, parent_id=2)
├─ 娱乐 (id=5, parent_id=1)
│  ├─ 电影 (id=6, parent_id=5)
│  └─ 电视剧 (id=7, parent_id=5)
└─ 体育 (id=8, parent_id=1)
```

---

### 4. ow_users - 用户表

**用途**: 存储用户账号和个人信息

**表结构**:

```sql
CREATE TABLE `ow_users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `group_id` int(11) NOT NULL,
  `level_id` int(11) NOT NULL,
  `username` varchar(32) NOT NULL,
  `password` varchar(64) NOT NULL,
  `nickname` varchar(64) NOT NULL,
  `sex` tinyint(2) NOT NULL DEFAULT '0',
  `birthday` varchar(64) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(64) DEFAULT NULL,
  `duty` varchar(64) DEFAULT NULL,
  `office_phone` varchar(64) DEFAULT NULL,
  `home_phone` varchar(64) DEFAULT NULL,
  `mobile_phone` varchar(64) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  `register_at` int(11) NOT NULL DEFAULT '0',
  `register_ip` char(15) NOT NULL DEFAULT '0.0.0.0',
  `login_count` int(11) NOT NULL DEFAULT '0',
  `login_at` int(11) NOT NULL DEFAULT '0',
  `login_ip` char(15) NOT NULL DEFAULT '0.0.0.0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_password` (`password`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**字段说明**:

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| id | int(11) | 主键ID | 1 |
| group_id | int(11) | 所属组ID | 2 |
| level_id | int(11) | 浏览级别ID | 3 |
| username | varchar(32) | 用户名（唯一） | "zhangsan" |
| password | varchar(64) | 密码（bcrypt哈希） | "$2a$10$..." |
| nickname | varchar(64) | 昵称 | "张三" |
| sex | tinyint(2) | 性别 | 0=保密, 1=男, 2=女 |
| email | varchar(64) | 邮箱 | "zhangsan@example.com" |
| enabled | tinyint(2) | 启用状态 | 0=禁用, 1=启用 |
| register_at | int(11) | 注册时间戳 | 1704038400 |
| register_ip | char(15) | 注册IP | "192.168.1.100" |
| login_count | int(11) | 登录次数 | 58 |
| login_at | int(11) | 最后登录时间 | 1704124800 |
| login_ip | char(15) | 最后登录IP | "192.168.1.100" |

**密码加密**: 使用 bcrypt 算法，salt轮数=10

---

### 5. ow_groups - 用户组表

**用途**: 用户组管理，用于权限批量分配

```sql
CREATE TABLE `ow_groups` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `quota` int(11) NOT NULL DEFAULT '1000',
  `level` int(11) NOT NULL DEFAULT '1',
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**字段说明**:

| 字段 | 说明 | 示例 |
|------|------|------|
| id | 主键ID | 1 |
| name | 组名 | "编辑部" |
| description | 描述 | "负责内容编辑和编目" |
| quota | 磁盘配额(MB) | 10000 (10GB) |
| level | 级别值 | 5 |
| enabled | 启用状态 | 1 |

---

### 6. ow_roles - 角色表

**用途**: 角色定义，连接组和权限

```sql
CREATE TABLE `ow_roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT '1',
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**预定义角色**:

| ID | 角色名 | 描述 | 权限范围 |
|----|--------|------|---------|
| 1 | 管理员 | System Administrator | 全部权限 |
| 2 | 编目员 | Content Cataloger | 上传、编目 |
| 3 | 审核员 | Content Reviewer | 审核、发布 |
| 4 | 查看者 | Viewer | 查看、下载 |

---

### 7. ow_permissions - 权限表

**用途**: 定义系统所有权限点

```sql
CREATE TABLE `ow_permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `namespace` varchar(64) NOT NULL DEFAULT 'default',
  `controller` varchar(64) NOT NULL DEFAULT 'default',
  `action` varchar(64) NOT NULL DEFAULT 'index',
  `aliasname` varchar(64) NOT NULL DEFAULT '',
  `rbac` varchar(32) NOT NULL DEFAULT 'ACL_NULL',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**权限示例**:

| ID | Namespace | Controller | Action | Alias | RBAC |
|----|-----------|------------|--------|-------|------|
| 1 | default | files | list | 文件列表 | ACL_NONE |
| 2 | default | files | upload | 文件上传 | ACL_LOGGED |
| 3 | default | files | delete | 文件删除 | ACL_ADMIN |
| 4 | admin | users | list | 用户管理 | ACL_ADMIN |

---

### 8. ow_levels - 浏览级别表

**用途**: 定义文件和用户的浏览级别

```sql
CREATE TABLE `ow_levels` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `level` int(11) NOT NULL DEFAULT '1',
  `enabled` tinyint(2) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**级别定义**:

| ID | 名称 | Level值 | 说明 |
|----|------|---------|------|
| 1 | 公开 | 1 | 所有用户可见 |
| 2 | 内部 | 3 | 内部员工可见 |
| 3 | 受限 | 5 | 高级用户可见 |
| 4 | 机密 | 7 | 管理层可见 |
| 5 | 绝密 | 10 | 仅管理员可见 |

**权限控制逻辑**:
```
用户可见文件 = 用户.level_id >= 文件.level
```

---

### 9. ow_groups_has_category - 组-分类关联表

**用途**: 控制用户组对分类的访问权限

```sql
CREATE TABLE `ow_groups_has_category` (
  `group_id` int(11) NOT NULL,
  `category_id` int(11) NOT NULL,
  PRIMARY KEY (`group_id`, `category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**数据示例**:

| group_id | category_id | 说明 |
|----------|-------------|------|
| 1 | 1 | 管理员组可访问全部分类 |
| 2 | 2 | 新闻组可访问新闻分类 |
| 2 | 3 | 新闻组可访问时政新闻 |

---

### 10. ow_groups_has_roles - 组-角色关联表

**用途**: 为用户组分配角色

```sql
CREATE TABLE `ow_groups_has_roles` (
  `group_id` int(11) NOT NULL,
  `role_id` int(11) NOT NULL,
  PRIMARY KEY (`group_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

### 11. ow_roles_has_permissions - 角色-权限关联表

**用途**: 为角色分配权限

```sql
CREATE TABLE `ow_roles_has_permissions` (
  `role_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

### 12. ow_files_counter - 文件计数表

**用途**: Sphinx 搜索引擎使用的文档计数器

```sql
CREATE TABLE `ow_files_counter` (
  `id` int(11) NOT NULL,
  `file_id` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

### 13. cs_counter - Sphinx计数表

**用途**: Sphinx 增量索引使用的最大文档ID记录

```sql
CREATE TABLE `cs_counter` (
  `id` int(11) NOT NULL,
  `maxid` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 索引设计

### 索引策略

1. **主键索引**: 所有表都有自增主键
2. **唯一索引**: username (用户名唯一)
3. **外键索引**: 所有外键字段都建立索引
4. **查询索引**: 常用查询字段建立索引
5. **组合索引**: 高频组合查询建立复合索引

### 性能优化建议

**避免全表扫描**:
- status, type, level 等枚举字段已建索引
- upload_at 时间字段已建索引

**索引维护**:
```sql
-- 分析表
ANALYZE TABLE ow_files;

-- 优化表
OPTIMIZE TABLE ow_files;

-- 查看索引使用情况
SHOW INDEX FROM ow_files;
```

---

## 数据字典

### 完整数据字典

详细字段说明请参考各表结构部分。

**数据字典导出**:
```bash
# 导出数据字典到 Excel
mysqldump -d openwan_db > schema.sql
```

---

## 迁移脚本

### 初始化脚本

**位置**: `/migrations/000001_init_schema.up.sql`

**执行方式**:

```bash
# 方式1: 使用 migrate 工具
migrate -path ./migrations -database "mysql://user:pass@localhost:3306/openwan_db" up

# 方式2: 直接执行 SQL
mysql -u root -p openwan_db < migrations/000001_init_schema.up.sql
```

### 回滚脚本

**位置**: `/migrations/000001_init_schema.down.sql`

```bash
migrate -path ./migrations -database "mysql://user:pass@localhost:3306/openwan_db" down
```

### 数据迁移

**从 PHP 版本迁移**:

参考 `/docs/migration-guide.md`

---

## 附录

### A. 常用SQL查询

**查询用户权限**:
```sql
SELECT p.*
FROM ow_permissions p
JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id
JOIN ow_roles r ON rhp.role_id = r.id
JOIN ow_groups_has_roles ghr ON r.id = ghr.role_id
JOIN ow_users u ON ghr.group_id = u.group_id
WHERE u.username = 'zhangsan';
```

**查询用户可访问的分类**:
```sql
SELECT c.*
FROM ow_category c
JOIN ow_groups_has_category ghc ON c.id = ghc.category_id
JOIN ow_users u ON ghc.group_id = u.group_id
WHERE u.username = 'zhangsan';
```

**查询用户可见的文件**:
```sql
SELECT f.*
FROM ow_files f
JOIN ow_users u ON (
  u.level_id >= f.level
  AND (f.groups = 'all' OR FIND_IN_SET(u.group_id, f.groups) > 0)
)
WHERE u.username = 'zhangsan'
  AND f.status = 2;
```

### B. 数据备份

**备份命令**:
```bash
# 完整备份
mysqldump -u root -p --single-transaction openwan_db > backup_$(date +%Y%m%d).sql

# 仅结构
mysqldump -u root -p -d openwan_db > schema_only.sql

# 仅数据
mysqldump -u root -p -t openwan_db > data_only.sql
```

**恢复命令**:
```bash
mysql -u root -p openwan_db < backup_20260201.sql
```

### C. 性能监控

```sql
-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query%';

-- 查看表大小
SELECT 
  table_name,
  ROUND(((data_length + index_length) / 1024 / 1024), 2) AS `Size (MB)`
FROM information_schema.TABLES
WHERE table_schema = 'openwan_db'
ORDER BY (data_length + index_length) DESC;
```

---

**文档版本**: v2.0  
**最后更新**: 2026-02-01  
**维护者**: OpenWan 开发团队
