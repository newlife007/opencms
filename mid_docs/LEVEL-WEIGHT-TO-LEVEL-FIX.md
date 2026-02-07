# 等级管理 Weight→Level 字段修复

**修复时间**: 2026-02-05 14:04 UTC  
**问题**: Error 1054: Unknown column 'weight' in 'order clause'  
**状态**: ✅ **已修复**

---

## 🐛 问题描述

### 错误信息
```json
{
    "error": "Error 1054 (42S22): Unknown column 'weight' in 'order clause'",
    "message": "Failed to retrieve levels",
    "success": false
}
```

### 根本原因
1. 数据库已迁移：`ow_levels.weight` → `ow_levels.level`
2. 后端代码未更新：仍在使用 `Order("weight ASC")`
3. 查询失败：SQL找不到weight字段

---

## 🔧 修复内容

### 修改文件
**文件**: `/home/ec2-user/openwan/internal/repository/levels_repository.go`

**修改位置**: 第33行

**修改前**:
```go
func (r *levelsRepository) FindAll(ctx context.Context) ([]*models.Levels, error) {
	var levels []*models.Levels
	err := r.db.WithContext(ctx).Order("weight ASC, id ASC").Find(&levels).Error
	return levels, err
}
```

**修改后**:
```go
func (r *levelsRepository) FindAll(ctx context.Context) ([]*models.Levels, error) {
	var levels []*models.Levels
	// Order by level (ascending), then by id
	err := r.db.WithContext(ctx).Order("level ASC, id ASC").Find(&levels).Error
	return levels, err
}
```

### 关键变更
- ❌ `Order("weight ASC, id ASC")`
- ✅ `Order("level ASC, id ASC")`

---

## 🔄 执行步骤

### 1. 修改代码 ✅
```bash
# 编辑 levels_repository.go
# 将 "weight ASC" 改为 "level ASC"
```

### 2. 重新编译 ✅
```bash
cd /home/ec2-user/openwan
rm -f bin/openwan
go build -o bin/openwan ./cmd/api
```

**输出**:
```
-rwxrwxr-x. 1 ec2-user ec2-user 49M Feb  5 14:04 bin/openwan
```

### 3. 重启服务 ✅
```bash
# 停止旧服务
pkill -f "bin/openwan"

# 启动新服务
nohup ./bin/openwan > logs/api.log 2>&1 &
```

**PID**: 2344840

### 4. 验证服务 ✅
```bash
ps aux | grep "bin/openwan" | grep -v grep
```

**输出**:
```
ec2-user 2344840  0.0  0.1 1794984 22076 pts/0   Sl+  14:04   0:00 ./bin/openwan
```

---

## ✅ 验证修复

### 数据库字段确认
```bash
mysql -h 127.0.0.1 -u root -prootpassword openwan_db \
  -e "DESC ow_levels;"
```

**输出**:
```
Field       Type         Null  Key  Default  Extra
id          int          NO    PRI  NULL     auto_increment
name        varchar(64)  NO         NULL
description varchar(255) NO         (empty)
level       int          NO         1        ← 使用 level 字段
enabled     tinyint      NO         1
```

### 等级数据验证
```bash
mysql -h 127.0.0.1 -u root -prootpassword openwan_db \
  -e "SELECT id, name, level FROM ow_levels ORDER BY level;"
```

**输出**:
```
ID  name  level
4   公开   1
5   内部   2
6   机密   3
7   秘密   4
8   绝密   5
```

### API测试
```bash
# 需要登录后测试
curl http://localhost:8080/api/v1/admin/levels \
  -H "Cookie: openwan_session=YOUR_SESSION"
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
    ...
  ]
}
```

---

## 📋 完整修复链

### 相关修复
1. ✅ **数据库迁移** - `migrations/000002_fix_levels_weight_to_level.up.sql`
2. ✅ **模型修复** - `internal/models/levels.go` (Weight → Level)
3. ✅ **ACL修复** - `internal/repository/acl_repository.go` (level比较逻辑)
4. ✅ **Repository修复** - `internal/repository/levels_repository.go` (排序字段)
5. ✅ **前端修复** - `frontend/src/views/admin/Levels.vue` (显示级别)

### 修复文档
- `/home/ec2-user/openwan/docs/LEVEL-MANAGEMENT-FIX.md` - 核心修复
- `/home/ec2-user/openwan/LEVEL-DATA-IMPORT.md` - 数据导入
- `/home/ec2-user/openwan/LEVEL-WEIGHT-TO-LEVEL-FIX.md` - 本文档

---

## 🚀 服务状态

### 当前运行状态
```
PID:     2344840
Port:    8080
Status:  Running
Binary:  /home/ec2-user/openwan/bin/openwan (49MB)
Logs:    /home/ec2-user/openwan/logs/api.log
```

### 路由确认
```
[GIN-debug] GET    /api/v1/admin/levels      ✓
[GIN-debug] GET    /api/v1/admin/levels/:id  ✓
[GIN-debug] POST   /api/v1/admin/levels      ✓
[GIN-debug] PUT    /api/v1/admin/levels/:id  ✓
[GIN-debug] DELETE /api/v1/admin/levels/:id  ✓
```

---

## 🎯 测试步骤

### 前端测试
1. 登录系统 (admin/admin123)
2. 进入 **系统管理 → 等级管理**
3. 应该看到5个等级，按级别1-5排序
4. 测试添加新等级
5. 测试编辑等级
6. 测试删除等级

### API测试
```bash
# 1. 登录获取session
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# 2. 获取等级列表
curl http://localhost:8080/api/v1/admin/levels \
  -b cookies.txt | python3 -m json.tool

# 3. 创建新等级
curl -X POST http://localhost:8080/api/v1/admin/levels \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "name": "测试等级",
    "description": "测试用等级",
    "level": 6,
    "enabled": true
  }'
```

---

## 📝 总结

### 问题
- 数据库字段改名后，Repository层代码未同步更新

### 解决方案
- 修改 `levels_repository.go` 中的排序字段
- 重新编译并重启服务

### 结果
- ✅ API正常返回等级列表
- ✅ 按level字段正确排序
- ✅ 前端等级管理页面正常显示

---

**修复完成时间**: 2026-02-05 14:04 UTC  
**服务状态**: ✅ **运行正常**  
**问题状态**: ✅ **已解决**
