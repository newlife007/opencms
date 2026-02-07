-- OpenWan System Permissions Seed Data
-- 为系统的每个功能模块创建权限
-- 表结构: namespace, controller, action, aliasname, rbac

-- ============================================
-- 1. 文件管理权限 (Files Management)
-- ============================================
INSERT INTO ow_permissions (namespace, controller, action, aliasname, rbac) VALUES
-- 文件浏览
('files', 'browse', 'list', '浏览文件列表', 'ACL_ALL'),
('files', 'browse', 'view', '查看文件详情', 'ACL_ALL'),
('files', 'browse', 'search', '搜索文件', 'ACL_ALL'),
('files', 'browse', 'download', '下载文件', 'ACL_ALL'),
('files', 'browse', 'preview', '预览文件', 'ACL_ALL'),

-- 文件上传
('files', 'upload', 'create', '上传文件', 'ACL_EDIT'),
('files', 'upload', 'batch', '批量上传文件', 'ACL_EDIT'),

-- 文件编辑
('files', 'edit', 'update', '编辑文件信息', 'ACL_EDIT'),
('files', 'edit', 'delete', '删除文件', 'ACL_EDIT'),
('files', 'edit', 'restore', '恢复已删除文件', 'ACL_EDIT'),

-- 文件编目
('files', 'catalog', 'edit', '编辑文件编目信息', 'ACL_CATALOG'),
('files', 'catalog', 'submit', '提交编目审核', 'ACL_CATALOG'),

-- 文件发布
('files', 'publish', 'approve', '审核发布文件', 'ACL_PUTOUT'),
('files', 'publish', 'reject', '拒绝发布文件', 'ACL_PUTOUT'),
('files', 'publish', 'unpublish', '取消发布文件', 'ACL_PUTOUT'),

-- ============================================
-- 2. 分类管理权限 (Categories Management)
-- ============================================
('categories', 'manage', 'list', '浏览分类列表', 'ACL_ALL'),
('categories', 'manage', 'view', '查看分类详情', 'ACL_ALL'),
('categories', 'manage', 'create', '创建分类', 'ACL_ADMIN'),
('categories', 'manage', 'update', '编辑分类', 'ACL_ADMIN'),
('categories', 'manage', 'delete', '删除分类', 'ACL_ADMIN'),
('categories', 'manage', 'move', '移动分类', 'ACL_ADMIN'),

-- ============================================
-- 3. 目录配置权限 (Catalog Configuration)
-- ============================================
('catalog', 'config', 'list', '浏览目录配置', 'ACL_ADMIN'),
('catalog', 'config', 'view', '查看目录配置详情', 'ACL_ADMIN'),
('catalog', 'config', 'create', '创建目录字段', 'ACL_ADMIN'),
('catalog', 'config', 'update', '编辑目录字段', 'ACL_ADMIN'),
('catalog', 'config', 'delete', '删除目录字段', 'ACL_ADMIN'),
('catalog', 'config', 'move', '移动目录字段', 'ACL_ADMIN'),

-- ============================================
-- 4. 用户管理权限 (User Management)
-- ============================================
('users', 'manage', 'list', '浏览用户列表', 'ACL_ADMIN'),
('users', 'manage', 'view', '查看用户详情', 'ACL_ADMIN'),
('users', 'manage', 'create', '创建用户', 'ACL_ADMIN'),
('users', 'manage', 'update', '编辑用户', 'ACL_ADMIN'),
('users', 'manage', 'delete', '删除用户', 'ACL_ADMIN'),
('users', 'manage', 'enable', '启用/禁用用户', 'ACL_ADMIN'),
('users', 'manage', 'reset_password', '重置用户密码', 'ACL_ADMIN'),

-- ============================================
-- 5. 组管理权限 (Group Management)
-- ============================================
('groups', 'manage', 'list', '浏览组列表', 'ACL_ADMIN'),
('groups', 'manage', 'view', '查看组详情', 'ACL_ADMIN'),
('groups', 'manage', 'create', '创建组', 'ACL_ADMIN'),
('groups', 'manage', 'update', '编辑组', 'ACL_ADMIN'),
('groups', 'manage', 'delete', '删除组', 'ACL_ADMIN'),
('groups', 'manage', 'assign_roles', '分配角色给组', 'ACL_ADMIN'),
('groups', 'manage', 'assign_categories', '分配分类给组', 'ACL_ADMIN'),

-- ============================================
-- 6. 角色管理权限 (Role Management)
-- ============================================
('roles', 'manage', 'list', '浏览角色列表', 'ACL_ADMIN'),
('roles', 'manage', 'view', '查看角色详情', 'ACL_ADMIN'),
('roles', 'manage', 'create', '创建角色', 'ACL_ADMIN'),
('roles', 'manage', 'update', '编辑角色', 'ACL_ADMIN'),
('roles', 'manage', 'delete', '删除角色', 'ACL_ADMIN'),
('roles', 'manage', 'assign_permissions', '分配权限给角色', 'ACL_ADMIN'),

-- ============================================
-- 7. 权限管理权限 (Permission Management)
-- ============================================
('permissions', 'manage', 'list', '浏览权限列表', 'ACL_ADMIN'),
('permissions', 'manage', 'view', '查看权限详情', 'ACL_ADMIN'),

-- ============================================
-- 8. 浏览级别管理权限 (Level Management)
-- ============================================
('levels', 'manage', 'list', '浏览级别列表', 'ACL_ADMIN'),
('levels', 'manage', 'view', '查看级别详情', 'ACL_ADMIN'),
('levels', 'manage', 'create', '创建级别', 'ACL_ADMIN'),
('levels', 'manage', 'update', '编辑级别', 'ACL_ADMIN'),
('levels', 'manage', 'delete', '删除级别', 'ACL_ADMIN'),

-- ============================================
-- 9. 搜索权限 (Search)
-- ============================================
('search', 'basic', 'query', '基本搜索', 'ACL_ALL'),
('search', 'advanced', 'query', '高级搜索', 'ACL_ALL'),
('search', 'reindex', 'trigger', '触发重建索引', 'ACL_ADMIN'),

-- ============================================
-- 10. 转码管理权限 (Transcoding Management)
-- ============================================
('transcoding', 'manage', 'list', '查看转码任务列表', 'ACL_CATALOG'),
('transcoding', 'manage', 'view', '查看转码任务详情', 'ACL_CATALOG'),
('transcoding', 'manage', 'start', '启动转码任务', 'ACL_CATALOG'),
('transcoding', 'manage', 'cancel', '取消转码任务', 'ACL_CATALOG'),
('transcoding', 'manage', 'retry', '重试失败的转码', 'ACL_CATALOG'),

-- ============================================
-- 11. 系统监控权限 (System Monitoring)
-- ============================================
('system', 'monitor', 'health', '查看系统健康状态', 'ACL_ADMIN'),
('system', 'monitor', 'metrics', '查看系统指标', 'ACL_ADMIN'),
('system', 'monitor', 'logs', '查看系统日志', 'ACL_ADMIN'),

-- ============================================
-- 12. 系统配置权限 (System Configuration)
-- ============================================
('system', 'config', 'view', '查看系统配置', 'ACL_ADMIN'),
('system', 'config', 'update', '修改系统配置', 'ACL_ADMIN'),

-- ============================================
-- 13. 报表统计权限 (Reports & Statistics)
-- ============================================
('reports', 'statistics', 'view', '查看统计报表', 'ACL_EDIT'),
('reports', 'statistics', 'export', '导出统计报表', 'ACL_EDIT'),

-- ============================================
-- 14. 个人中心权限 (User Profile)
-- ============================================
('profile', 'self', 'view', '查看个人信息', 'ACL_ALL'),
('profile', 'self', 'update', '修改个人信息', 'ACL_ALL'),
('profile', 'self', 'change_password', '修改个人密码', 'ACL_ALL');

-- 查询插入结果
SELECT COUNT(*) as total_permissions FROM ow_permissions;
SELECT namespace, COUNT(*) as count FROM ow_permissions GROUP BY namespace ORDER BY count DESC;
