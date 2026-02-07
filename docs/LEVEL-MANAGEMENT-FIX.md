# 等级管理逻辑修复 - Weight改为Level

## 修复时间：2026-02-05 14:00

## ✅ 问题已修复！

---

## 问题描述

**用户反馈**:
等级设计逻辑有问题：
- ❌ 等级没有权重，只有级别
- ✅ 上传的文件内容有级别
- ✅ 用户也有级别
- ✅ 用户只能查看小于等于用户级别的文件

**用户说得对**！设计应该是基于级别（Level），不是权重（Weight）。

---

## 错误的设计 ❌

### 原来的表结构

**ow_levels表**:
```sql
CREATE TABLE `ow_levels` (
  `id` int(11) PRIMARY KEY,
  `name` varchar(64),
  `description` varchar(255),
  `weight` int(11) DEFAULT 0,  -- ❌ 错误：应该是level
  `enabled` tinyint(2)
);
```

### 原来的逻辑

**文件访问检查** (错误):
```go
// ❌ 比较user.LevelID（是ID）和file.Level（是数值）
if file.Level > user.LevelID {
    return false, nil
}
```

**问题**:
1. 比较的是ID和Level数值，类型不一致
2. Weight字段没有明确的级别语义
3. 用户级别逻辑不清晰

---

## 正确的设计 ✅

### 级别逻辑

**核心规则**:
- 用户有级别（通过level_id关联到ow_levels）
- 文件有级别（直接存储level数值）
- **用户只能查看 file.level <= user's level value 的文件**
- **级别越高，可访问的文件越多**

**示例**:
```
Level 1 (公开) - 可以看: level 1 文件
Level 2 (内部) - 可以看: level 1, 2 文件
Level 3 (机密) - 可以看: level 1, 2, 3 文件
Level 4 (秘密) - 可以看: level 1, 2, 3, 4 文件
Level 5 (绝密) - 可以看: level 1, 2, 3, 4, 5 文件
```

---

## 修复内容

### 1. 修复Levels模型

**文件**: `internal/models/levels.go`

**修改前** ❌:
```go
type Levels struct {
    ID          int    `gorm:"column:id;primaryKey" json:"id"`
    Name        string `gorm:"column:name" json:"name"`
    Description string `gorm:"column:description" json:"description"`
    Weight      int    `gorm:"column:weight" json:"weight"`  // ❌
    Enabled     bool   `gorm:"column:enabled" json:"enabled"`
}
```

**修改后** ✅:
```go
// Levels represents the ow_levels table for browsing level (access level) management
// Level logic: User can only view files with level <= user's level
// Higher level = More access (e.g., level 5 user can see level 1,2,3,4,5 files)
type Levels struct {
    ID          int    `gorm:"column:id;primaryKey" json:"id"`
    Name        string `gorm:"column:name" json:"name"`
    Description string `gorm:"column:description" json:"description"`
    Level       int    `gorm:"column:level" json:"level"` // ✅ Browsing level value (1-10)
    Enabled     bool   `gorm:"column:enabled" json:"enabled"`
}
```

---

### 2. 修复数据库迁移

**文件**: `migrations/000001_init_schema.up.sql`

**修改前** ❌:
```sql
CREATE TABLE `ow_levels` (
  `weight` int(11) NOT NULL DEFAULT '0' COMMENT 'Weight',
  ...
);
```

**修改后** ✅:
```sql
CREATE TABLE `ow_levels` (
  `level` int(11) NOT NULL DEFAULT '1' COMMENT 'Level value (higher = more access)',
  ...
);
```

**新建迁移** `migrations/000002_fix_levels_weight_to_level.up.sql`:
```sql
-- Change column from weight to level
ALTER TABLE `ow_levels` 
CHANGE COLUMN `weight` `level` 
INT(11) NOT NULL DEFAULT 1 
COMMENT 'Level value (higher = more access)';

-- Update sample data
UPDATE `ow_levels` SET `level` = 1 WHERE `name` = 'Public';
UPDATE `ow_levels` SET `level` = 2 WHERE `name` = 'Internal';
UPDATE `ow_levels` SET `level` = 3 WHERE `name` = 'Confidential';
UPDATE `ow_levels` SET `level` = 4 WHERE `name` = 'Secret';
UPDATE `ow_levels` SET `level` = 5 WHERE `name` = 'Top Secret';
```

---

### 3. 修复文件访问检查逻辑

**文件**: `internal/repository/acl_repository.go`

**修改前** ❌:
```go
func (r *aclRepository) CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error) {
    var user models.Users
    if err := r.db.WithContext(ctx).Preload("Level").First(&user, userID).Error; err != nil {
        return false, err
    }
    
    var file models.Files
    if err := r.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
        return false, err
    }
    
    // ❌ 错误：比较file.Level（数值）和user.LevelID（ID）
    if file.Level > user.LevelID {
        return false, nil
    }
    
    // ... 其他检查
}
```

**修改后** ✅:
```go
func (r *aclRepository) CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error) {
    var user models.Users
    if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
        return false, err
    }
    
    // ✅ 获取用户级别详情
    var userLevel models.Levels
    if err := r.db.WithContext(ctx).First(&userLevel, user.LevelID).Error; err != nil {
        return false, err
    }
    
    var file models.Files
    if err := r.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
        return false, err
    }
    
    // ✅ 正确：比较file.Level（数值）和userLevel.Level（数值）
    // Logic: User can only view files with level <= user's level
    // Higher level = More access (e.g., level 5 user can see level 1,2,3,4,5 files)
    if file.Level > userLevel.Level {
        return false, nil
    }
    
    // ... 其他检查（组权限）
}
```

**改进点**:
1. ✅ 先查询用户的Levels记录，获取实际level值
2. ✅ 比较 `file.Level` (数值) 和 `userLevel.Level` (数值)
3. ✅ 逻辑正确：file.Level <= userLevel.Level 可访问

---

### 4. 修复前端等级管理页面

**文件**: `frontend/src/views/admin/Levels.vue`

**修改内容**:

#### 表格列
```vue
<!-- 修改前 ❌ -->
<el-table-column prop="weight" label="权重" width="100">
  <template #default="{ row }">
    <el-tag>{{ row.weight }}</el-tag>
  </template>
</el-table-column>

<!-- 修改后 ✅ -->
<el-table-column prop="level" label="级别" width="100">
  <template #default="{ row }">
    <el-tag type="warning">{{ row.level }}</el-tag>
  </template>
</el-table-column>
```

#### 表单字段
```vue
<!-- 修改前 ❌ -->
<el-form-item label="权重" prop="weight">
  <el-input-number 
    v-model="form.weight" 
    :min="0" 
    :max="999"
    placeholder="数值越大权限越高"
  />
  <span>数值越大，权限越高</span>
</el-form-item>

<!-- 修改后 ✅ -->
<el-form-item label="级别" prop="level">
  <el-input-number 
    v-model="form.level" 
    :min="1" 
    :max="10"
    placeholder="级别数值（1-10）"
  />
  <span>级别越高，可访问的文件越多</span>
</el-form-item>
```

#### 数据模型
```javascript
// 修改前 ❌
const form = ref({
  id: null,
  name: '',
  description: '',
  weight: 0,  // ❌
  enabled: true
})

// 修改后 ✅
const form = ref({
  id: null,
  name: '',
  description: '',
  level: 1,  // ✅ 默认级别1
  enabled: true
})
```

#### 验证规则
```javascript
// 修改前 ❌
weight: [
  { required: true, message: '请输入权重', trigger: 'blur' },
  { type: 'number', message: '权重必须为数字', trigger: 'blur' }
]

// 修改后 ✅
level: [
  { required: true, message: '请输入级别', trigger: 'blur' },
  { type: 'number', min: 1, max: 10, message: '级别范围为 1-10', trigger: 'blur' }
]
```

---

## 数据库表结构

### ow_users表
```sql
CREATE TABLE `ow_users` (
  `id` int(11) PRIMARY KEY,
  `username` varchar(64),
  `level_id` int(11),  -- ✅ 关联到ow_levels.id
  ...
);
```

### ow_levels表
```sql
CREATE TABLE `ow_levels` (
  `id` int(11) PRIMARY KEY,
  `name` varchar(64),
  `description` varchar(255),
  `level` int(11) DEFAULT 1,  -- ✅ 级别数值（1-10）
  `enabled` tinyint(2)
);
```

### ow_files表
```sql
CREATE TABLE `ow_files` (
  `id` bigint(20) PRIMARY KEY,
  `title` varchar(255),
  `level` int(11) DEFAULT 1,  -- ✅ 文件级别（直接存储数值）
  `groups` varchar(255),      -- 组访问控制
  ...
);
```

---

## 访问控制逻辑

### 完整流程

```
1. 用户请求访问文件
   userID = 5, fileID = 100
   ↓
2. 查询用户
   user.level_id = 3  (关联到ow_levels.id=3)
   ↓
3. 查询用户级别
   userLevel = ow_levels.find(3)
   userLevel.level = 5  (实际级别值)
   ↓
4. 查询文件
   file.level = 3  (文件级别)
   ↓
5. 级别检查
   file.level (3) <= userLevel.level (5) ?
   3 <= 5 → true ✅
   ↓
6. 组权限检查
   file.groups = "1,2,5" 或 "all"
   user.group_id = 2
   groups.includes(2) → true ✅
   ↓
7. 允许访问 ✅
```

### 示例场景

**场景1: 高级别用户访问低级别文件** ✅

```
用户级别: 5 (绝密)
文件级别: 2 (内部)
检查: 2 <= 5 → true ✅
结果: 允许访问
```

**场景2: 低级别用户访问高级别文件** ❌

```
用户级别: 2 (内部)
文件级别: 4 (秘密)
检查: 4 <= 2 → false ❌
结果: 拒绝访问
```

**场景3: 同级别访问** ✅

```
用户级别: 3 (机密)
文件级别: 3 (机密)
检查: 3 <= 3 → true ✅
结果: 允许访问
```

---

## 级别定义建议

| ID | 名称 | Level | 描述 | 可访问文件 |
|----|------|-------|------|-----------|
| 1 | 公开 | 1 | 完全公开的内容 | level 1 |
| 2 | 内部 | 2 | 公司内部员工可见 | level 1-2 |
| 3 | 机密 | 3 | 部门级机密 | level 1-3 |
| 4 | 秘密 | 4 | 公司级秘密 | level 1-4 |
| 5 | 绝密 | 5 | 高层管理可见 | level 1-5 |

**创建命令**:
```sql
INSERT INTO ow_levels (name, description, level, enabled) VALUES
('公开', '完全公开的内容，所有人可访问', 1, 1),
('内部', '公司内部员工可见', 2, 1),
('机密', '部门级机密信息', 3, 1),
('秘密', '公司级秘密资料', 4, 1),
('绝密', '高层管理人员可见', 5, 1);
```

---

## 部署步骤

### 1. 后端编译

```bash
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api
```

**状态**: ✅ 编译成功

### 2. 前端构建

```bash
cd /home/ec2-user/openwan/frontend
npm run build
```

**状态**: ✅ 构建成功

### 3. 数据库迁移

**如果数据库正在运行**:
```bash
mysql -u root -p openwan_db < migrations/000002_fix_levels_weight_to_level.up.sql
```

**或在应用启动时自动执行** (如果配置了迁移工具)

### 4. 重启服务

```bash
# 停止旧服务
pkill -f "bin/openwan"

# 启动新服务
cd /home/ec2-user/openwan
./bin/openwan &
```

---

## 测试验证

### 1. 创建测试级别

**登录admin**，进入"系统管理 → 等级管理"

**添加级别**:
- 名称: 公开
- 描述: 完全公开的内容
- 级别: 1
- 状态: 启用

重复创建级别2-5。

### 2. 创建测试用户

**用户A**:
- 用户名: user_level2
- 级别: 2 (内部)

**用户B**:
- 用户名: user_level4
- 级别: 4 (秘密)

### 3. 上传测试文件

**文件1**:
- 标题: 公开文档
- 级别: 1

**文件2**:
- 标题: 机密资料
- 级别: 3

### 4. 验证访问控制

**user_level2登录**:
- ✅ 可以看到"公开文档" (level 1 <= 2)
- ❌ 不能看到"机密资料" (level 3 > 2)

**user_level4登录**:
- ✅ 可以看到"公开文档" (level 1 <= 4)
- ✅ 可以看到"机密资料" (level 3 <= 4)

---

## API更新

### 等级管理API

**列表** `GET /api/v1/admin/levels`:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "公开",
      "description": "完全公开的内容",
      "level": 1,  ← 返回level字段
      "enabled": true
    }
  ]
}
```

**创建/更新** `POST/PUT /api/v1/admin/levels`:
```json
{
  "name": "机密",
  "description": "部门级机密信息",
  "level": 3,  ← 提交level字段（1-10）
  "enabled": true
}
```

---

## 相关文档

- **数据库迁移**: `/home/ec2-user/openwan/migrations/000002_fix_levels_weight_to_level.up.sql`
- **模型定义**: `/home/ec2-user/openwan/internal/models/levels.go`
- **访问控制**: `/home/ec2-user/openwan/internal/repository/acl_repository.go`
- **前端管理**: `/home/ec2-user/openwan/frontend/src/views/admin/Levels.vue`

---

## 总结

### 问题

❌ Weight（权重）概念不明确，级别逻辑错误

❌ 比较user.LevelID（ID）和file.Level（数值），类型不一致

❌ 前端使用"权重"标签，语义模糊

### 解决

✅ 改为Level（级别），语义清晰

✅ 正确比较userLevel.Level和file.Level

✅ 前端显示"级别"，范围1-10

### 逻辑

✅ 用户级别 >= 文件级别 → 可访问

✅ 级别越高，可访问的文件越多

✅ 符合保密等级的直觉理解

---

**修复完成时间**: 2026-02-05 14:00  
**修复人员**: AWS Transform CLI  
**版本**: 5.5 Level Management Logic Fix  
**状态**: ✅ **已修复并编译成功**

**下一步**: 部署后测试等级访问控制功能！
