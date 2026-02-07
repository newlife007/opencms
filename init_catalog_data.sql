-- OpenWan 属性配置初始化数据
-- 基于旧系统 (legacy-php) 的 ow_catalog 表数据
-- 创建时间: 2026-02-05 15:00 UTC

-- 清空现有数据（如果有）
TRUNCATE TABLE ow_catalog;

-- ============================================
-- 根节点：编目信息
-- ============================================
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated) 
VALUES (1, -1, '-1,', '编目信息', '编目信息', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 一级分类：按文件类型
-- ============================================

-- 视频编目 (type=1)
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (2, 1, '-1,1,', '视频编目', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 音频编目 (type=2)
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (3, 1, '-1,1,', '音频编目', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 图片编目 (type=3)
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (4, 1, '-1,1,', '图片编目', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 富媒体编目 (type=4)
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (5, 1, '-1,1,', '富媒体编目', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 视频属性字段 (parent_id=2)
-- ============================================

-- 题名组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (67, 2, '-1,1,2,', '题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (68, 67, '-1,1,2,67,', '正题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (69, 67, '-1,1,2,67,', '并列正题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (70, 67, '-1,1,2,67,', '副题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (71, 67, '-1,1,2,67,', '交替副题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (72, 67, '-1,1,2,67,', '题名说明', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (73, 67, '-1,1,2,67,', '系列题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (75, 67, '-1,1,2,67,', '并列题名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (76, 67, '-1,1,2,67,', '分集总数', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (77, 67, '-1,1,2,67,', '分集次', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 主题组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (74, 2, '-1,1,2,', '主题', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (78, 74, '-1,1,2,74,', '分类名', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (79, 74, '-1,1,2,74,', '主题词', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 音频属性字段 (parent_id=3)
-- ============================================

-- 基本信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (61, 3, '-1,1,3,', '基本信息', '', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (66, 61, '-1,1,3,61,', '标题', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (65, 61, '-1,1,3,61,', '地方', '', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 其他音频字段
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (57, 3, '-1,1,3,', 'sdfasd', '', 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (59, 57, '-1,1,3,57,', 'cce', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 图片属性字段 (parent_id=4)
-- ============================================

-- 示例字段组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (10, 4, '-1,1,4,', '地方放的', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (29, 4, '-1,1,4,', '啊啊是', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (30, 4, '-1,1,4,', '版本', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 子字段
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (31, 10, '-1,1,4,10,', '断点', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (32, 10, '-1,1,4,10,', '搜索', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (33, 29, '-1,1,4,29,', '撒', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (34, 29, '-1,1,4,29,', '公告', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (35, 29, '-1,1,4,29,', '嗯嗯', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (36, 30, '-1,1,4,30,', '高高挂', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (37, 30, '-1,1,4,30,', '呵呵', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 富媒体属性字段 (parent_id=5)
-- ============================================

-- 示例字段组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (11, 5, '-1,1,5,', '爱爱爱', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (49, 5, '-1,1,5,', 'fgfdg', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 子字段
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (14, 11, '-1,1,5,11,', '11', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (15, 11, '-1,1,5,11,', '分割', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (16, 11, '-1,1,5,11,', '过后', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (50, 49, '-1,1,5,49,', 'ddd', '', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 重置自增ID
-- ============================================
ALTER TABLE ow_catalog AUTO_INCREMENT = 80;

-- 查看导入结果
SELECT COUNT(*) as total_records FROM ow_catalog;
SELECT parent_id, COUNT(*) as count FROM ow_catalog GROUP BY parent_id ORDER BY parent_id;
