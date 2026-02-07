-- 插入测试分类数据
INSERT INTO ow_category (id, parent_id, path, name, description, weight, enabled, created, updated) VALUES
(1, 0, '1', '视频资源', '视频文件分类', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, 0, '2', '音频资源', '音频文件分类', 2, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 0, '3', '图片资源', '图片文件分类', 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4, 0, '4', '文档资源', '文档文件分类', 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(5, 1, '1,5', '教学视频', '教学相关视频', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(6, 1, '1,6', '宣传视频', '宣传推广视频', 2, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(7, 2, '2,7', '背景音乐', '背景音乐素材', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(8, 3, '3,8', '产品图片', '产品展示图片', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入测试文件数据
INSERT INTO ow_files (id, category_id, category_name, upload_username, catalog_username, putout_username, type, title, name, ext, size, path, status, level, `groups`, catalog_info, is_download, upload_at, catalog_at, putout_at, created, updated) VALUES
(1, 5, '教学视频', 'admin', 'admin', NULL, 1, '系统功能介绍视频', 'sample_video_01', 'mp4', 15728640, '/data/sample/video01.mp4', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, 5, '教学视频', 'admin', 'admin', NULL, 1, '操作指南视频', 'sample_video_02', 'mp4', 20971520, '/data/sample/video02.mp4', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 7, '背景音乐', 'admin', 'admin', NULL, 2, '轻音乐-春天', 'sample_audio_01', 'mp3', 5242880, '/data/sample/audio01.mp3', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4, 8, '产品图片', 'admin', NULL, NULL, 3, '产品展示图1', 'sample_image_01', 'jpg', 1048576, '/data/sample/image01.jpg', 1, 1, '1', '{}', 1, UNIX_TIMESTAMP(), NULL, NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(5, 8, '产品图片', 'admin', NULL, NULL, 3, '产品展示图2', 'sample_image_02', 'png', 2097152, '/data/sample/image02.png', 0, 1, '1', '{}', 1, UNIX_TIMESTAMP(), NULL, NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(6, 4, '文档资源', 'admin', 'admin', NULL, 4, '用户手册', 'user_manual', 'pdf', 3145728, '/data/sample/manual.pdf', 2, 1, 'all', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 授予管理员组访问所有分类的权限
INSERT INTO ow_groups_has_category (group_id, category_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8);
