INSERT INTO ow_files (id, category_id, category_name, upload_username, catalog_username, putout_username, type, title, name, ext, size, path, status, level, `groups`, catalog_info, is_download, upload_at, catalog_at, putout_at) VALUES
(1, 5, '教学视频', 'admin', 'admin', NULL, 1, '系统功能介绍视频', 'sample_video_01', 'mp4', 15728640, '/data/sample/video01.mp4', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL),
(2, 5, '教学视频', 'admin', 'admin', NULL, 1, '操作指南视频', 'sample_video_02', 'mp4', 20971520, '/data/sample/video02.mp4', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL),
(3, 7, '背景音乐', 'admin', 'admin', NULL, 2, '轻音乐-春天', 'sample_audio_01', 'mp3', 5242880, '/data/sample/audio01.mp3', 2, 1, '1', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL),
(4, 8, '产品图片', 'admin', NULL, NULL, 3, '产品展示图1', 'sample_image_01', 'jpg', 1048576, '/data/sample/image01.jpg', 1, 1, '1', '{}', 1, UNIX_TIMESTAMP(), NULL, NULL),
(5, 8, '产品图片', 'admin', NULL, NULL, 3, '产品展示图2', 'sample_image_02', 'png', 2097152, '/data/sample/image02.png', 0, 1, '1', '{}', 1, UNIX_TIMESTAMP(), NULL, NULL),
(6, 4, '文档资源', 'admin', 'admin', NULL, 4, '用户手册', 'user_manual', 'pdf', 3145728, '/data/sample/manual.pdf', 2, 1, 'all', '{}', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), NULL);
