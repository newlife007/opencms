# 角色与权限对应关系详解

## 用户提出的三个问题

1. **默认角色与权限的对应关系是如何对应的？**
2. **只是对应了大类还是与细粒度的权限对应了？**
3. **后端是通过角色识别权限？还是通过角色对应的实际权限来识别权限？**

---

## 问题1: 默认角色与权限的对应关系

### 当前状态（数据库实际情况）

```sql
SELECT 
    r.id, 
    r.name, 
    COUNT(rp.permission_id) as permission_count
FROM ow_roles r
LEFT JOIN ow_roles_has_permissions rp ON r.id = rp.role_id
GROUP BY r.id;
```

**结果**:
```
role_id  role_name      permission_count
1        超级管理员     72    ✅ 所有权限
2        内容管理员     5     ⚠️  只有5个浏览权限
3        审核员         0     ❌ 未分配权限
4        编辑           0     ❌ 未分配权限
5        查看者         0     ❌ 未分配权限
```

### 超级管理员(role 1)的权限分布

```sql
SELECT namespace, COUNT(*) as count
FROM ow_roles_has_permissions rp
JOIN ow_permissions p ON rp.permission_id = p.id
WHERE rp.role_id = 1
GROUP BY namespace;
```

**72个权限的模块分布**:
```
namespace      count   说明
----------------------------------------
files          15      文件管理（浏览、上传、编辑、编目、发布）
users          7       用户管理（CRUD、密码重置、状态切换）
groups         7       组管理（CRUD、分配分类、分配角色）
roles          6       角色管理（CRUD、分配权限）
categories     6       分类管理（CRUD、移动）
catalog        6       目录配置（CRUD、移动）
levels         5       级别管理（CRUD）
system         5       系统管理（配置、监控、日志）
transcoding    5       转码管理（查看、启动、取消、重试）
search         3       搜索（基本搜索、高级搜索、重建索引）
profile        3       个人中心（查看、修改、改密码）
permissions    2       权限管理（浏览、查看详情）
reports        2       报表统计（查看、导出）
----------------------------------------
总计           72      完整的细粒度权限
```

### 内容管理员(role 2)的当前权限

**仅分配了5个权限（测试数据）**:
```
id   namespace  controller  action    aliasname
1    files      browse      list      浏览文件列表
2    files      browse      view      查看文件详情
3    files      browse      search    搜索文件
4    files      browse      download  下载文件
5    files      browse      preview   预览文件
```

**问题**: 这不符合"内容管理员"的职责！应该有更多权限。

---

## 问题2: 是大类还是细粒度权限？

### 答案: **完全是细粒度权限！**

OpenWan的权限系统采用**三级细粒度结构**：

```
namespace.controller.action
   ↓          ↓          ↓
 模块       控制器      操作
```

### 示例：文件管理模块(files)的细粒度权限

```
namespace: files (15个细粒度权限)

├─ controller: browse (浏览控制器) - 5个权限
│  ├─ action: list      (浏览列表)
│  ├─ action: view      (查看详情)
│  ├─ action: search    (搜索)
│  ├─ action: download  (下载)
│  └─ action: preview   (预览)
│
├─ controller: upload (上传控制器) - 2个权限
│  ├─ action: create    (上传文件)
│  └─ action: batch     (批量上传)
│
├─ controller: edit (编辑控制器) - 3个权限
│  ├─ action: update    (编辑信息)
│  ├─ action: delete    (删除文件)
│  └─ action: restore   (恢复已删除)
│
├─ controller: catalog (编目控制器) - 2个权限
│  ├─ action: edit      (编辑编目信息)
│  └─ action: submit    (提交编目审核)
│
└─ controller: publish (发布控制器) - 3个权限
   ├─ action: approve   (审核发布)
   ├─ action: reject    (拒绝发布)
   └─ action: unpublish (取消发布)
```

### 权限粒度对比

**如果是大类权限（不推荐）**:
```
✗ files.all          (文件模块所有权限)
✗ users.all          (用户模块所有权限)
```

**OpenWan实际采用的细粒度权限（推荐）**:
```
✓ files.browse.list      (只能浏览文件列表)
✓ files.upload.create    (只能上传文件)
✓ files.edit.delete      (只能删除文件)
✓ users.manage.create    (只能创建用户)
✓ users.manage.delete    (只能删除用户)
```

**优势**:
- ✅ 精确控制：可以只给用户"上传"权限，不给"删除"权限
- ✅ 安全性高：最小权限原则
- ✅ 灵活性强：可以任意组合权限
- ✅ 审计清晰：知道用户具体能做什么

---

## 问题3: 后端是通过角色识别权限？还是通过实际权限？

### 答案: **通过角色对应的实际权限识别！**

### 权限验证流程（关键）

```
用户请求 → 中间件检查 → ACL服务 → 数据库查询 → 权限验证
```

#### Step 1: 用户发起请求

```bash
DELETE /api/v1/files/1
Authorization: Bearer {token}
```

#### Step 2: RequirePermission中间件拦截

```go
// internal/api/middleware/rbac.go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := int(c.Get("user_id").(uint))
        
        // 解析权限: "file.delete" → "files.edit.delete"
        namespace, controller, action := mapSimplifiedPermission(permission)
        
        // ✅ 关键：调用ACL服务检查实际权限
        hasPermission, _ := aclService.HasPermission(
            ctx, userID, namespace, controller, action
        )
        
        if !hasPermission {
            c.JSON(403, gin.H{"error": "Permission denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

**注意**: 这里传入的是**具体的权限标识符** (namespace, controller, action)，不是角色名！

#### Step 3: ACL Service处理

```go
// internal/service/acl_service.go
func (s *ACLService) HasPermission(ctx context.Context, userID int, 
    namespace, controller, action string) (bool, error) {
    
    // 直接调用repository查询实际权限
    return s.repo.ACL().HasPermission(ctx, userID, namespace, controller, action)
}
```

#### Step 4: ACL Repository查询数据库

```go
// internal/repository/acl_repository.go
func (r *aclRepository) HasPermission(ctx context.Context, userID int, 
    namespace, controller, action string) (bool, error) {
    
    // 1. 获取用户所属的组
    var user models.Users
    r.db.Preload("Group").First(&user, userID)
    
    // 2. 查询该组通过角色拥有的具体权限
    var count int64
    r.db.Table("ow_permissions p").
        Joins("JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id").
        Joins("JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id").
        Where("ghr.group_id = ? AND p.namespace = ? AND p.controller = ? AND p.action = ?",
            user.GroupID, namespace, controller, action).
        Count(&count)
    
    // 3. 有权限返回true，无权限返回false
    return count > 0, nil
}
```

### 数据库查询分析（4表JOIN）

```sql
SELECT COUNT(*) 
FROM ow_permissions p
JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id
JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id
WHERE ghr.group_id = ?              -- 用户的组
  AND p.namespace = 'files'         -- 权限的namespace
  AND p.controller = 'edit'         -- 权限的controller
  AND p.action = 'delete';          -- 权限的action
```

**查询逻辑解释**:

1. **ow_permissions**: 权限表（72个细粒度权限）
2. **ow_roles_has_permissions**: 角色-权限关联表
3. **ow_groups_has_roles**: 组-角色关联表
4. **用户通过GroupID关联到组**

**查询路径**:
```
用户 → 组 → 角色 → 权限 → 匹配 namespace.controller.action
```

**关键点**:
- ❌ 不是检查角色名称（如"超级管理员"、"编辑"）
- ✅ 是检查角色关联的具体权限记录
- ✅ 必须精确匹配 (namespace='files' AND controller='edit' AND action='delete')

---

## 完整示例：删除文件权限验证

### 场景1: 超级管理员（有权限）

**用户**: admin (user_id=1)  
**组**: 管理员组 (group_id=1)  
**角色**: 超级管理员 (role_id=1)  
**角色权限**: 72个权限（包含 files.edit.delete）

**请求**:
```bash
DELETE /api/v1/files/1
Authorization: Bearer {admin_token}
```

**验证流程**:
```
1. 中间件: RequirePermission("file.delete")
   ↓
2. 映射: "file.delete" → "files.edit.delete"
   ↓
3. ACL Service: HasPermission(1, "files", "edit", "delete")
   ↓
4. 数据库查询:
   SELECT COUNT(*) 
   FROM ow_permissions p
   JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id
   JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id
   WHERE ghr.group_id = 1                -- admin的组
     AND p.namespace = 'files'
     AND p.controller = 'edit'
     AND p.action = 'delete'
   
   结果: COUNT(*) = 1  ✅ (role 1有这个权限)
   ↓
5. 返回: hasPermission = true
   ↓
6. 中间件: c.Next() → 执行删除
   ↓
7. 响应: 200 OK ✅
```

### 场景2: 查看者（无权限）

**用户**: viewer (user_id=5)  
**组**: 普通用户组 (group_id=3)  
**角色**: 查看者 (role_id=5)  
**角色权限**: 假设只有 files.browse.* 权限（没有 files.edit.delete）

**请求**:
```bash
DELETE /api/v1/files/1
Authorization: Bearer {viewer_token}
```

**验证流程**:
```
1. 中间件: RequirePermission("file.delete")
   ↓
2. 映射: "file.delete" → "files.edit.delete"
   ↓
3. ACL Service: HasPermission(5, "files", "edit", "delete")
   ↓
4. 数据库查询:
   SELECT COUNT(*) 
   FROM ow_permissions p
   JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id
   JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id
   WHERE ghr.group_id = 3                -- viewer的组
     AND p.namespace = 'files'
     AND p.controller = 'edit'
     AND p.action = 'delete'
   
   结果: COUNT(*) = 0  ❌ (role 5没有这个权限)
   ↓
5. 返回: hasPermission = false
   ↓
6. 中间件: 403 Forbidden
   {
     "success": false,
     "message": "Permission denied",
     "error": "You don't have permission: file.delete"
   }
   ↓
7. 请求被拒绝 ❌
```

---

## 关键结论

### Q1: 默认角色与权限的对应关系？

**A**: 通过 `ow_roles_has_permissions` 表建立关联

```
角色表 (ow_roles)
  ↓ role_id
角色-权限关联表 (ow_roles_has_permissions)
  ↓ permission_id
权限表 (ow_permissions)
```

**当前状态**:
- ✅ 超级管理员: 72个权限（完整）
- ⚠️  内容管理员: 5个权限（不完整）
- ❌ 其他角色: 0个权限（未分配）

### Q2: 大类还是细粒度权限？

**A**: 完全是细粒度权限！

- ✅ 三级结构: `namespace.controller.action`
- ✅ 72个独立的细粒度权限
- ✅ 每个操作都有对应的权限记录
- ✅ 可以精确控制到具体功能

### Q3: 通过角色还是实际权限识别？

**A**: 通过角色对应的实际权限识别！

**流程**:
```
1. 获取用户的组ID
2. 查询该组分配的角色
3. 查询这些角色拥有的权限
4. 检查是否有匹配的具体权限
5. 根据查询结果允许/拒绝访问
```

**不是检查**:
- ❌ 角色名称（"超级管理员"、"编辑"）
- ❌ 角色级别
- ❌ 用户属性

**而是检查**:
- ✅ 角色关联的具体权限记录
- ✅ 精确的 namespace.controller.action 匹配
- ✅ 数据库中的实际权限数据

---

## 推荐的权限配置

### 1. 超级管理员 (role_id=1) - ✅ 已配置

**权限**: 所有72个权限  
**职责**: 系统管理员，拥有所有权限

### 2. 内容管理员 (role_id=2) - ⚠️ 需要补充

**推荐权限** (应该有 ~30个权限):
```sql
-- 文件管理（全部15个）
files.browse.*      (5个)
files.upload.*      (2个)
files.edit.*        (3个)
files.catalog.*     (2个)
files.publish.*     (3个)

-- 分类管理（全部6个）
categories.manage.*

-- 目录配置（全部6个）
catalog.config.*

-- 搜索（3个）
search.*
```

**当前只有5个**: files.browse.* (不符合职责)

### 3. 审核员 (role_id=3) - ❌ 需要配置

**推荐权限** (~10个):
```sql
-- 浏览文件
files.browse.*      (5个)

-- 审核发布
files.publish.*     (3个)

-- 搜索
search.*            (3个)
```

### 4. 编辑 (role_id=4) - ❌ 需要配置

**推荐权限** (~12个):
```sql
-- 浏览文件
files.browse.*      (5个)

-- 上传文件
files.upload.*      (2个)

-- 编目
files.catalog.*     (2个)

-- 搜索
search.*            (3个)

-- 个人中心
profile.*           (3个)
```

### 5. 查看者 (role_id=5) - ❌ 需要配置

**推荐权限** (~8个):
```sql
-- 浏览文件
files.browse.*      (5个)

-- 搜索
search.*            (3个)
```

---

## 如何为角色分配权限

### 方法1: 通过前端UI

1. 登录管理员账号
2. 进入"角色管理"页面
3. 点击角色的"分配权限"按钮
4. 勾选需要的权限（按模块分组显示）
5. 点击"确定"保存

### 方法2: 通过API

```bash
# 为role 2（内容管理员）分配权限
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
     -H "Content-Type: application/json" \
     "http://localhost:8080/api/v1/admin/roles/2/permissions" \
     -d '{
       "permission_ids": [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27]
     }'
```

### 方法3: 直接操作数据库

```sql
-- 为role 2分配文件管理的所有权限（15个）
INSERT INTO ow_roles_has_permissions (role_id, permission_id)
SELECT 2, id FROM ow_permissions WHERE namespace = 'files';

-- 为role 2分配分类管理权限（6个）
INSERT INTO ow_roles_has_permissions (role_id, permission_id)
SELECT 2, id FROM ow_permissions WHERE namespace = 'categories';

-- 为role 2分配目录配置权限（6个）
INSERT INTO ow_roles_has_permissions (role_id, permission_id)
SELECT 2, id FROM ow_permissions WHERE namespace = 'catalog';
```

---

## 验证权限是否生效

### 测试步骤

1. **创建测试用户**:
```sql
-- 创建一个编辑用户（属于组2，假设组2分配了role 4）
INSERT INTO ow_users (username, password, group_id, level_id, enabled)
VALUES ('editor', '$2a$10$...', 2, 1, 1);
```

2. **为role 4分配权限**:
```bash
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
     "http://localhost:8080/api/v1/admin/roles/4/permissions" \
     -d '{"permission_ids": [1,2,3,4,5,6,7,11,12]}'
```

3. **用editor登录并测试**:
```bash
# 登录
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"username":"editor","password":"pass123"}' | jq -r .token)

# 测试上传（应该成功 - 有权限）
curl -X POST -H "Authorization: Bearer $TOKEN" \
     -F "file=@test.jpg" \
     "http://localhost:8080/api/v1/files"
# 期望: 200 OK ✅

# 测试删除（应该失败 - 无权限）
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8080/api/v1/files/1"
# 期望: 403 Forbidden ✅
{
  "success": false,
  "message": "Permission denied",
  "error": "You don't have permission: file.delete"
}
```

---

## 总结

### 三个关键点

1. **细粒度权限**: OpenWan采用 `namespace.controller.action` 三级细粒度权限，共72个独立权限

2. **角色-权限映射**: 通过 `ow_roles_has_permissions` 表建立角色与具体权限的关联关系

3. **实际权限验证**: 后端通过查询数据库中用户的实际权限记录进行验证，不是通过角色名称判断

### 当前状态

- ✅ 权限系统架构完整且正确
- ✅ 超级管理员权限配置完整（72个）
- ⚠️  其他角色权限需要根据职责配置
- ✅ 权限验证逻辑完全生效

### 建议

为2-5号角色配置合适的权限，使其符合职责定义：
- 内容管理员: ~30个权限（文件+分类+目录）
- 审核员: ~10个权限（浏览+审核）
- 编辑: ~12个权限（浏览+上传+编目）
- 查看者: ~8个权限（浏览+搜索）

---

## 相关文件

- 权限表: `ow_permissions` (72条记录)
- 角色表: `ow_roles` (5条记录)
- 角色-权限关联: `ow_roles_has_permissions` (73条记录，role 1有72个)
- ACL Repository: `/home/ec2-user/openwan/internal/repository/acl_repository.go`
- ACL Service: `/home/ec2-user/openwan/internal/service/acl_service.go`
- RBAC Middleware: `/home/ec2-user/openwan/internal/api/middleware/rbac.go`
