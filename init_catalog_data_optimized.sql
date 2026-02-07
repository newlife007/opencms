-- OpenWan 属性配置优化版初始化数据
-- 基于旧系统结构，优化为实用的元数据字段
-- 创建时间: 2026-02-05 15:05 UTC

-- 清空现有数据
TRUNCATE TABLE ow_catalog;

-- ============================================
-- 根节点（保留但不显示）
-- ============================================
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated) 
VALUES (1, 0, '1,', '编目信息', '文件元数据根节点', 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 视频属性 (用于 type=1)
-- ============================================

-- 基本信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (10, 1, '1,10,', '基本信息', '视频基本信息', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (11, 10, '1,10,11,', '标题', '视频标题', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (12, 10, '1,10,12,', '副标题', '视频副标题', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (13, 10, '1,10,13,', '描述', '视频描述说明', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (14, 10, '1,10,14,', '关键词', '搜索关键词，逗号分隔', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 内容信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (20, 1, '1,20,', '内容信息', '视频内容详细信息', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (21, 20, '1,20,21,', '导演', '导演姓名', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (22, 20, '1,20,22,', '主演', '主要演员，逗号分隔', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (23, 20, '1,20,23,', '制作单位', '制作方/出品方', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (24, 20, '1,20,24,', '制作日期', '制作完成日期', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (25, 20, '1,20,25,', '系列名称', '所属系列', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (26, 20, '1,20,26,', '集次', '第几集/第几期', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 技术参数组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (30, 1, '1,30,', '技术参数', '视频技术参数', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (31, 30, '1,30,31,', '时长', '视频时长（分钟）', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (32, 30, '1,30,32,', '分辨率', '视频分辨率，如1920x1080', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (33, 30, '1,30,33,', '视频格式', '原始视频格式，如MP4/AVI', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (34, 30, '1,30,34,', '视频编码', '视频编码格式，如H.264', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (35, 30, '1,30,35,', '音频编码', '音频编码格式，如AAC', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (36, 30, '1,30,36,', '码率', '视频码率，如5000kbps', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (37, 30, '1,30,37,', '帧率', '视频帧率，如25fps', 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 版权信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (40, 1, '1,40,', '版权信息', '版权相关信息', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (41, 40, '1,40,41,', '版权方', '版权所有方', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (42, 40, '1,40,42,', '授权类型', '授权方式：独家/非独家', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (43, 40, '1,40,43,', '授权日期', '授权开始日期', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (44, 40, '1,40,44,', '授权期限', '授权有效期（年）', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (45, 40, '1,40,45,', '使用范围', '使用场景限制', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 音频属性 (用于 type=2)
-- ============================================

-- 基本信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (50, 1, '1,50,', '音频基本信息', '音频文件基本信息', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (51, 50, '1,50,51,', '曲名', '歌曲或音频名称', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (52, 50, '1,50,52,', '艺术家', '演唱者/演奏者', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (53, 50, '1,50,53,', '专辑', '所属专辑', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (54, 50, '1,50,54,', '作词', '作词人', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (55, 50, '1,50,55,', '作曲', '作曲人', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (56, 50, '1,50,56,', '发行年份', '发行年份', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (57, 50, '1,50,57,', '语言', '歌曲语言', 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (58, 50, '1,50,58,', '风格', '音乐风格/流派', 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 技术参数组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (60, 1, '1,60,', '音频参数', '音频技术参数', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (61, 60, '1,60,61,', '时长', '音频时长（分:秒）', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (62, 60, '1,60,62,', '比特率', '音频比特率，如320kbps', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (63, 60, '1,60,63,', '采样率', '采样率，如44100Hz', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (64, 60, '1,60,64,', '音频格式', '文件格式，如MP3/WAV/FLAC', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (65, 60, '1,60,65,', '声道', '声道数，如立体声/单声道', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 图片属性 (用于 type=3)
-- ============================================

-- 基本信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (70, 1, '1,70,', '图片基本信息', '图片文件基本信息', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (71, 70, '1,70,71,', '图片名称', '图片标题', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (72, 70, '1,70,72,', '图片描述', '图片说明', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (73, 70, '1,70,73,', '摄影师', '拍摄者/摄影师', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (74, 70, '1,70,74,', '拍摄日期', '拍摄时间', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (75, 70, '1,70,75,', '拍摄地点', '拍摄位置', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (76, 70, '1,70,76,', '主题标签', '图片标签，逗号分隔', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 技术参数组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (80, 1, '1,80,', '图片参数', '图片技术参数', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (81, 80, '1,80,81,', '分辨率', '图片尺寸，如1920x1080', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (82, 80, '1,80,82,', '图片格式', '文件格式，如JPG/PNG/GIF', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (83, 80, '1,80,83,', '色彩模式', '颜色模式，如RGB/CMYK', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (84, 80, '1,80,84,', 'DPI', '图片DPI', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (85, 80, '1,80,85,', '文件大小', '原始文件大小（KB）', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 富媒体属性 (用于 type=4)
-- ============================================

-- 基本信息组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (90, 1, '1,90,', '富媒体基本信息', '富媒体文件基本信息', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (91, 90, '1,90,91,', '文档标题', '文档名称', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (92, 90, '1,90,92,', '文档描述', '文档说明', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (93, 90, '1,90,93,', '作者', '文档作者', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (94, 90, '1,90,94,', '创建日期', '文档创建日期', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (95, 90, '1,90,95,', '版本号', '文档版本', 6, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (96, 90, '1,90,96,', '关键词', '搜索关键词', 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 技术参数组
INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (100, 1, '1,100,', '文档参数', '文档技术参数', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (101, 100, '1,100,101,', '文件格式', '文档格式，如PDF/DOC/PPT', 10, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (102, 100, '1,100,102,', '页数', '文档总页数', 9, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (103, 100, '1,100,103,', '文件大小', '文件大小（MB）', 8, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO ow_catalog (id, parent_id, path, name, description, weight, enabled, created, updated)
VALUES (104, 100, '1,100,104,', '语言', '文档语言', 7, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ============================================
-- 重置自增ID
-- ============================================
ALTER TABLE ow_catalog AUTO_INCREMENT = 105;

-- ============================================
-- 查看导入结果
-- ============================================
SELECT '============================================' as '';
SELECT '属性配置初始化完成' as 'Status';
SELECT '============================================' as '';
SELECT '' as '';

SELECT '总记录数统计：' as '';
SELECT COUNT(*) as '总记录数' FROM ow_catalog;
SELECT '' as '';

SELECT '按文件类型分组统计：' as '';
SELECT 
    CASE parent_id
        WHEN 0 THEN '根节点'
        WHEN 1 THEN '一级分类'
        ELSE '子字段'
    END as '节点类型',
    COUNT(*) as '数量'
FROM ow_catalog 
GROUP BY 
    CASE parent_id
        WHEN 0 THEN '根节点'
        WHEN 1 THEN '一级分类'
        ELSE '子字段'
    END;
SELECT '' as '';

SELECT '一级分类列表：' as '';
SELECT id, name, description, weight 
FROM ow_catalog 
WHERE parent_id = 1 
ORDER BY weight DESC, id;
