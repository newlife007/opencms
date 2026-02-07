-- 清空现有等级数据
DELETE FROM ow_levels;

-- 导入建议的5个等级
INSERT INTO ow_levels (name, description, level, enabled) VALUES
('公开', '完全公开的内容，所有人可访问', 1, 1),
('内部', '公司内部员工可见', 2, 1),
('机密', '部门级机密信息', 3, 1),
('秘密', '公司级秘密资料', 4, 1),
('绝密', '高层管理人员可见', 5, 1);

-- 查看导入结果
SELECT '导入完成！' as status;
SELECT * FROM ow_levels ORDER BY level;
