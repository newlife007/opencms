# 等级数据导入完成！

**导入时间**: 2026-02-05 14:05 UTC  
**状态**: ✅ **成功**

---

## 📊 导入的等级数据

| ID | 等级名称 | 级别值 | 描述 | 状态 | 可访问文件 |
|----|---------|--------|------|------|-----------|
| 4 | 公开 | 1 | 完全公开的内容，所有人可访问 | ✅ 启用 | level 1 |
| 5 | 内部 | 2 | 公司内部员工可见 | ✅ 启用 | level 1-2 |
| 6 | 机密 | 3 | 部门级机密信息 | ✅ 启用 | level 1-3 |
| 7 | 秘密 | 4 | 公司级秘密资料 | ✅ 启用 | level 1-4 |
| 8 | 绝密 | 5 | 高层管理人员可见 | ✅ 启用 | level 1-5 |

---

## 🔄 执行的操作

### 1. 数据库迁移 ✅
```sql
-- 修改字段名: weight → level
ALTER TABLE `ow_levels` 
CHANGE COLUMN `weight` `level` 
INT(11) NOT NULL DEFAULT 1 
COMMENT 'Level value (higher = more access)';
```

**执行**: `/home/ec2-user/openwan/migrations/000002_fix_levels_weight_to_level.up.sql`

### 2. 清空旧数据 ✅
```sql
DELETE FROM ow_levels;
```

### 3. 导入新数据 ✅
```sql
INSERT INTO ow_levels (name, description, level, enabled) VALUES
('公开', '完全公开的内容，所有人可访问', 1, 1),
('内部', '公司内部员工可见', 2, 1),
('机密', '部门级机密信息', 3, 1),
('秘密', '公司级秘密资料', 4, 1),
('绝密', '高层管理人员可见', 5, 1);
```

**执行**: `/home/ec2-user/openwan/import-levels.sql`

---

## 📋 等级逻辑说明

### 核心规则
```
用户级别 >= 文件级别 → 可访问 ✅
用户级别 < 文件级别 → 拒绝访问 ❌
```

### 访问控制示例

#### 公开用户 (Level 1)
- ✅ 可以访问: level 1 文件
- ❌ 不能访问: level 2, 3, 4, 5 文件

#### 内部用户 (Level 2)
- ✅ 可以访问: level 1, 2 文件
- ❌ 不能访问: level 3, 4, 5 文件

#### 机密用户 (Level 3)
- ✅ 可以访问: level 1, 2, 3 文件
- ❌ 不能访问: level 4, 5 文件

#### 秘密用户 (Level 4)
- ✅ 可以访问: level 1, 2, 3, 4 文件
- ❌ 不能访问: level 5 文件

#### 绝密用户 (Level 5)
- ✅ 可以访问: level 1, 2, 3, 4, 5 所有文件

---

## 🧪 验证数据

### 查询等级列表
```bash
mysql -h 127.0.0.1 -u root -prootpassword openwan_db -e \
  "SELECT * FROM ow_levels ORDER BY level;"
```

### 查看表结构
```bash
mysql -h 127.0.0.1 -u root -prootpassword openwan_db -e \
  "DESC ow_levels;"
```

**输出**:
```
Field       Type         Null  Key  Default  Extra
id          int          NO    PRI  NULL     auto_increment
name        varchar(64)  NO         NULL
description varchar(255) NO         (empty)
level       int          NO         1        ← 改为level字段
enabled     tinyint      NO         1
```

---

## 🌐 API测试

### 获取等级列表 (需要登录)
```bash
# 1. 先登录获取session
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# 2. 使用session访问等级API
curl http://localhost:8080/api/v1/admin/levels \
  -b cookies.txt
```

**期望响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 4,
      "name": "公开",
      "description": "完全公开的内容，所有人可访问",
      "level": 1,
      "enabled": true
    },
    {
      "id": 5,
      "name": "内部",
      "description": "公司内部员工可见",
      "level": 2,
      "enabled": true
    },
    ...
  ]
}
```

---

## 👥 用户等级设置

### 更新用户等级
```sql
-- 设置用户为"内部"等级 (level_id=5, level值=2)
UPDATE ow_users SET level_id = 5 WHERE username = 'test_user';

-- 设置用户为"机密"等级 (level_id=6, level值=3)
UPDATE ow_users SET level_id = 6 WHERE username = 'manager';

-- 设置管理员为"绝密"等级 (level_id=8, level值=5)
UPDATE ow_users SET level_id = 8 WHERE username = 'admin';
```

---

## 📁 文件等级设置

### 上传文件时设置等级
文件上传时，在metadata中指定level值（1-5）：
```json
{
  "title": "机密文档",
  "category_id": 1,
  "level": 3,  ← 设置为机密级别
  "groups": "all"
}
```

### 更新文件等级
```sql
-- 设置文件为"机密"级别
UPDATE ow_files SET level = 3 WHERE id = 1;

-- 设置文件为"公开"级别
UPDATE ow_files SET level = 1 WHERE id = 2;
```

---

## 🎯 前端展示

### 等级管理页面
访问: **系统管理 → 等级管理**

应该显示5个等级：

| ID | 等级名称 | 级别 | 描述 | 状态 | 操作 |
|----|---------|------|------|------|------|
| 4 | 公开 | 1 | 完全公开的内容... | ✓ | 编辑/删除 |
| 5 | 内部 | 2 | 公司内部员工... | ✓ | 编辑/删除 |
| 6 | 机密 | 3 | 部门级机密... | ✓ | 编辑/删除 |
| 7 | 秘密 | 4 | 公司级秘密... | ✓ | 编辑/删除 |
| 8 | 绝密 | 5 | 高层管理人员... | ✓ | 编辑/删除 |

---

## 📝 相关文件

- **迁移脚本**: `/home/ec2-user/openwan/migrations/000002_fix_levels_weight_to_level.up.sql`
- **导入脚本**: `/home/ec2-user/openwan/import-levels.sql`
- **等级修复文档**: `/home/ec2-user/openwan/docs/LEVEL-MANAGEMENT-FIX.md`
- **前端组件**: `/home/ec2-user/openwan/frontend/src/views/admin/Levels.vue`
- **后端模型**: `/home/ec2-user/openwan/internal/models/levels.go`
- **ACL检查**: `/home/ec2-user/openwan/internal/repository/acl_repository.go`

---

## ✅ 总结

### 完成的工作
1. ✅ 数据库字段迁移: `weight` → `level`
2. ✅ 清空旧的3条数据
3. ✅ 导入新的5个等级数据
4. ✅ 验证数据正确性

### 数据状态
- **总数**: 5个等级
- **级别范围**: 1-5
- **状态**: 全部启用
- **字段**: level (INT, DEFAULT 1)

### 下一步
1. 从浏览器登录后端
2. 访问"系统管理 → 等级管理"
3. 验证5个等级正确显示
4. 测试添加/编辑等级功能

---

**导入完成时间**: 2026-02-05 14:05 UTC  
**数据库**: openwan_db @ 127.0.0.1:3306  
**状态**: ✅ **全部成功**
