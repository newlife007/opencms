# 默认权限分析：无角色用户为什么能访问系统

## 用户问题

**问题**: 创建了一个user用户，设置为查看组，查看组并没有分配角色，为什么该用户可以访问系统呢？是有默认权限吗？默认权限是什么呢？

## 问题分析

### 实际情况

**用户**: user (user_id=16)  
**所属组**: 查看组 (group_id=4)  
**组的角色**: 0个（未分配任何角色）  
**用户权限**: 0个（没有任何权限）

```bash
# 登录响应
{
    "success": true,
    "token": "e504294a-fca7-470b-92ed-59da52a559ea",
    "user": {
        "id": 16,
        "username": "user",
        "permissions": [],  # ← 没有权限
        "roles": []         # ← 没有角色
    }
}
```

### 测试结果

| 操作 | 需要权限 | 结果 | 原因 |
|------|---------|------|------|
| 登录 | 无 | ✅ 成功 | 登录不需要权限 |
| 浏览文件列表 | 无 | ✅ 成功 | 接口没有权限检查 |
| 查看文件详情 | 无 | ✅ 成功 | 接口没有权限检查 |
| 下载文件 | 无 | ✅ 成功 | 接口没有权限检查 |
| 预览文件 | 无 | ✅ 成功 | 接口没有权限检查 |
| **上传文件** | 无 | ✅ 成功 | 只有RequireAuth，无权限检查 |
| 编辑文件 | 无 | ✅ 成功 | 只有RequireAuth，无权限检查 |
| **删除文件** | files.edit.delete | ❌ 403 | 有权限检查，拒绝访问 |
| 审核发布 | files.publish.approve | ❌ 403 | 有权限检查，拒绝访问 |

---

## 核心发现：OpenWan的权限设计策略

### 策略1: 基础读取操作 - 无权限限制（默认开放）

**设计理念**: 已登录用户都可以浏览和查看内容

```go
// router.go - files路由配置
files := v1.Group("/files")
{
    // ✅ 无权限检查 - 所有已登录用户都可以访问
    files.GET("", fileHandler.ListFiles())              // 浏览列表
    files.GET("/:id", fileHandler.GetFile())            // 查看详情
    files.GET("/:id/download", fileHandler.DownloadFile()) // 下载
    files.GET("/:id/preview", fileHandler.PreviewFile())   // 预览
    files.GET("/stats", fileHandler.GetStats())         // 统计信息
    files.GET("/recent", fileHandler.GetRecentFiles())  // 最近文件
}
```

**影响**:
- ✅ 任何已登录用户都可以浏览文件列表
- ✅ 任何已登录用户都可以查看文件详情
- ✅ 任何已登录用户都可以下载文件
- ✅ 任何已登录用户都可以预览文件

**为什么这样设计？**
1. **OpenWan是媒体资产管理系统** - 主要功能是让用户查看和使用媒体资源
2. **内容发现** - 用户需要能够浏览和搜索才能找到需要的资源
3. **降低使用门槛** - 不需要为每个用户分配"查看"权限
4. **细粒度控制在文件级别** - 通过文件的 `level` 和 `groups` 字段控制访问

### 策略2: 写入和管理操作 - 只需登录（部分开放）

```go
// ⚠️ 只有RequireAuth中间件，没有RequirePermission
files.POST("", middleware.RequireAuth(), fileHandler.Upload())
files.PUT("/:id", middleware.RequireAuth(), fileHandler.UpdateFile())
files.POST("/:id/submit", middleware.RequireAuth(), workflowHandler.SubmitForReview())
```

**影响**:
- ✅ 任何已登录用户都可以上传文件
- ✅ 任何已登录用户都可以编辑文件
- ✅ 任何已登录用户都可以提交编目审核

**这是设计问题还是有意为之？**

这看起来是**权限配置不完整**的问题！

**应该添加权限检查**:
```go
// 推荐配置
files.POST("", 
    middleware.RequireAuth(), 
    middleware.RequirePermission("file.upload"),  // ← 应该添加
    fileHandler.Upload()
)

files.PUT("/:id", 
    middleware.RequireAuth(),
    middleware.RequirePermission("file.edit"),    // ← 应该添加
    fileHandler.UpdateFile()
)
```

### 策略3: 危险操作 - 需要特定权限（严格控制）

```go
// ✅ 有RequirePermission中间件 - 必须有权限才能访问
files.DELETE("/:id", 
    middleware.RequireAuth(), 
    middleware.RequirePermission("file.delete"),  // ← 权限检查
    fileHandler.DeleteFile()
)

files.POST("/:id/publish", 
    middleware.RequireAuth(), 
    middleware.RequirePermission("file.publish"), // ← 权限检查
    workflowHandler.PublishFile()
)

files.POST("/:id/reject", 
    middleware.RequireAuth(), 
    middleware.RequirePermission("file.reject"),  // ← 权限检查
    workflowHandler.RejectFile()
)
```

**影响**:
- ❌ 删除文件 - 需要 `files.edit.delete` 权限
- ❌ 审核发布 - 需要 `files.publish.approve` 权限
- ❌ 拒绝发布 - 需要 `files.publish.reject` 权限
- ❌ 修改状态 - 需要 `file.manage` 权限

---

## 完整路由权限配置分析

### Files 模块

| 路由 | 方法 | 权限要求 | 实际配置 | 问题 |
|------|------|---------|---------|------|
| /files | GET | 无 | 无 | ⚠️ 应该考虑文件级别控制 |
| /files/:id | GET | 无 | 无 | ⚠️ 应该考虑文件级别控制 |
| /files/:id/download | GET | 无 | 无 | ⚠️ 应该考虑文件级别控制 |
| /files/:id/preview | GET | 无 | 无 | ⚠️ 应该考虑文件级别控制 |
| /files | POST | files.upload.create | **RequireAuth** | ❌ **缺少权限检查** |
| /files/:id | PUT | files.edit.update | **RequireAuth** | ❌ **缺少权限检查** |
| /files/:id | DELETE | files.edit.delete | ✅ RequirePermission | ✅ 正确 |
| /files/:id/submit | POST | files.catalog.submit | **RequireAuth** | ❌ **缺少权限检查** |
| /files/:id/publish | POST | files.publish.approve | ✅ RequirePermission | ✅ 正确 |
| /files/:id/reject | POST | files.publish.reject | ✅ RequirePermission | ✅ 正确 |

### Categories 模块

```go
categories := v1.Group("/categories")
{
    categories.GET("", categoryHandler.ListCategories())                    // ⚠️ 无权限检查
    categories.GET("/tree", categoryHandler.GetCategoryTree())              // ⚠️ 无权限检查
    categories.GET("/:id", categoryHandler.GetCategory())                   // ⚠️ 无权限检查
    categories.POST("", middleware.RequireAuth(), 
        middleware.RequirePermission("category.create"),                   // ✅ 有权限检查
        categoryHandler.CreateCategory())
    categories.PUT("/:id", middleware.RequireAuth(), 
        middleware.RequirePermission("category.update"),                   // ✅ 有权限检查
        categoryHandler.UpdateCategory())
    categories.DELETE("/:id", middleware.RequireAuth(), 
        middleware.RequirePermission("category.delete"),                   // ✅ 有权限检查
        categoryHandler.DeleteCategory())
}
```

**分析**: Categories模块配置较好，读操作开放，写操作有权限控制

### Admin 模块

```go
adminGroup := v1.Group("/admin")
adminGroup.Use(middleware.RequireAuth())  // ✅ 所有admin路由都需要登录
{
    // ⚠️ 但是没有RequireAdmin或RequirePermission中间件！
    adminGroup.GET("/groups", groupHandler.ListGroups())
    adminGroup.POST("/groups", groupHandler.CreateGroup())
    // ...
}
```

**严重问题**: Admin路由只检查登录，不检查管理员权限！

**应该改为**:
```go
adminGroup := v1.Group("/admin")
adminGroup.Use(middleware.RequireAuth())
adminGroup.Use(middleware.RequireAdmin())  // ← 应该添加
```

---

## 文件级别的权限控制（补充机制）

虽然路由层面没有完全限制，但OpenWan还有**文件级别的权限控制**：

### Files表的权限字段

```sql
CREATE TABLE ow_files (
    level INT,           -- 浏览级别（数字越大权限越高）
    groups VARCHAR(255), -- 允许访问的组ID（逗号分隔，或"all"）
    is_download TINYINT, -- 是否允许下载
    ...
);
```

### 文件访问控制逻辑

```go
// internal/repository/acl_repository.go
func (r *aclRepository) CanAccessFile(ctx context.Context, userID int, fileID uint64) (bool, error) {
    // 1. 获取用户的级别
    var user models.Users
    r.db.First(&user, userID)
    
    // 2. 获取文件信息
    var file models.Files
    r.db.First(&file, fileID)
    
    // 3. 检查级别限制
    if file.Level > user.LevelID {
        return false, nil  // 用户级别不够
    }
    
    // 4. 检查组限制
    if file.Groups == "all" {
        return true, nil   // 所有人都可以访问
    }
    
    // 5. 检查用户是否在允许的组中
    allowedGroups := strings.Split(file.Groups, ",")
    userGroupStr := fmt.Sprintf("%d", user.GroupID)
    for _, g := range allowedGroups {
        if strings.TrimSpace(g) == userGroupStr {
            return true, nil
        }
    }
    
    return false, nil  // 不在允许的组中
}
```

**但是问题**: 这个 `CanAccessFile` 函数定义了，但在文件下载、预览等接口中**没有被调用**！

---

## 总结：OpenWan的"默认权限"

### 不是真正的"默认权限"

OpenWan **没有**为用户分配默认权限！

user用户的 `permissions: []` 和 `roles: []` 证明了这一点。

### 实际情况：权限配置策略

OpenWan采用了**分层的权限控制策略**：

#### Layer 1: 路由层权限（API访问控制）

**完全开放的操作**:
- ✅ 浏览文件列表
- ✅ 查看文件详情
- ✅ 下载文件
- ✅ 预览文件
- ✅ 查看分类
- ✅ 搜索

**只需登录的操作** (⚠️ 缺少权限检查):
- ⚠️ 上传文件
- ⚠️ 编辑文件
- ⚠️ 提交编目审核
- ⚠️ 管理后台所有操作

**需要权限的操作** (✅ 正确配置):
- ✅ 删除文件
- ✅ 审核发布
- ✅ 拒绝发布
- ✅ 创建/编辑/删除分类

#### Layer 2: 文件级别权限（数据访问控制）

通过文件的 `level` 和 `groups` 字段控制：
- 用户级别 >= 文件级别
- 用户组在文件允许的组列表中

**但是**: 这个机制在代码中定义了，但在实际的文件访问接口中**没有被调用**！

---

## 存在的问题

### 问题1: 上传和编辑缺少权限检查

**当前**:
```go
files.POST("", middleware.RequireAuth(), fileHandler.Upload())
files.PUT("/:id", middleware.RequireAuth(), fileHandler.UpdateFile())
```

**应该**:
```go
files.POST("", middleware.RequireAuth(), 
    middleware.RequirePermission("file.upload"), 
    fileHandler.Upload())
    
files.PUT("/:id", middleware.RequireAuth(), 
    middleware.RequirePermission("file.edit"), 
    fileHandler.UpdateFile())
```

### 问题2: Admin路由缺少管理员检查

**当前**:
```go
adminGroup := v1.Group("/admin")
adminGroup.Use(middleware.RequireAuth())  // 只检查登录
```

**应该**:
```go
adminGroup := v1.Group("/admin")
adminGroup.Use(middleware.RequireAuth())
adminGroup.Use(middleware.RequireAdmin())  // 添加管理员检查
```

### 问题3: 文件级别权限未实施

`CanAccessFile()` 函数已定义但未在以下接口中调用：
- GET /files/:id
- GET /files/:id/download
- GET /files/:id/preview

---

## 推荐的修复方案

### 方案1: 最小权限原则（推荐）

**为所有写操作添加权限检查**:

```go
// Files模块
files.POST("", middleware.RequireAuth(), 
    middleware.RequirePermission("file.upload"), 
    fileHandler.Upload())
    
files.PUT("/:id", middleware.RequireAuth(), 
    middleware.RequirePermission("file.edit"), 
    fileHandler.UpdateFile())
    
files.POST("/:id/submit", middleware.RequireAuth(), 
    middleware.RequirePermission("file.catalog"), 
    workflowHandler.SubmitForReview())

// Admin模块
adminGroup.Use(middleware.RequireAdmin())
```

**配置默认角色权限**:
- 查看组(group 4) → 查看者角色(role 5) → 浏览权限(files.browse.*)
- 编辑组(group 3) → 编辑角色(role 4) → 浏览+上传+编目权限
- 审核组(group 2) → 审核员角色(role 3) → 浏览+审核权限

### 方案2: 文件级别权限实施

**在文件访问接口中调用CanAccessFile**:

```go
func (h *FileHandler) GetFile() gin.HandlerFunc {
    return func(c *gin.Context) {
        fileID := c.Param("id")
        userID := c.Get("user_id").(uint)
        
        // 检查文件级别权限
        canAccess, err := h.aclService.CanAccessFile(ctx, int(userID), fileID)
        if !canAccess {
            c.JSON(403, gin.H{"error": "Access denied to this file"})
            return
        }
        
        // 继续处理...
    }
}
```

### 方案3: 混合方案（最佳）

1. **路由层**: 添加权限检查到写操作
2. **文件层**: 在读操作中调用CanAccessFile检查文件级别权限
3. **默认角色**: 为查看组分配查看者角色和基本浏览权限

---

## 回答用户的问题

### Q: 为什么user用户可以访问系统？

**A**: 因为很多接口**没有**设置权限要求，只要登录就可以访问：
- ✅ 浏览文件列表 - 无权限检查
- ✅ 查看文件详情 - 无权限检查
- ✅ 下载文件 - 无权限检查
- ✅ 上传文件 - 只检查登录，不检查权限
- ❌ 删除文件 - 有权限检查，user用户被拒绝

### Q: 是有默认权限吗？

**A**: **没有真正的默认权限**！

user用户的权限列表是空的:
```json
{
    "permissions": [],
    "roles": []
}
```

但是由于很多接口配置为"只需登录"或"完全开放"，所以user用户可以访问。

### Q: 默认权限是什么？

**A**: OpenWan采用了**分层开放策略**，不是默认权限：

**完全开放** (不需要登录):
- 无（所有接口都需要登录）

**只需登录** (不检查权限):
- 浏览文件列表、查看详情、下载、预览
- 上传文件（⚠️ 应该添加权限检查）
- 编辑文件（⚠️ 应该添加权限检查）
- 提交编目审核（⚠️ 应该添加权限检查）
- 管理后台操作（⚠️ 应该添加管理员检查）

**需要特定权限**:
- 删除文件
- 审核发布
- 创建/编辑/删除分类和目录

---

## 建议

### 立即修复

1. **为查看组分配查看者角色**:
```sql
-- 为group 4分配role 5
INSERT INTO ow_groups_has_roles (group_id, role_id) VALUES (4, 5);

-- 为role 5分配浏览权限
INSERT INTO ow_roles_has_permissions (role_id, permission_id)
SELECT 5, id FROM ow_permissions WHERE namespace = 'files' AND controller = 'browse';
```

2. **添加上传和编辑的权限检查**（修改router.go）

3. **添加Admin路由的管理员检查**（修改router.go）

### 长期改进

1. 实施文件级别权限检查
2. 完善权限粒度配置
3. 添加审计日志记录所有操作
4. 定期审查权限配置

---

## 相关文件

- 路由配置: `/home/ec2-user/openwan/internal/api/router.go`
- ACL Repository: `/home/ec2-user/openwan/internal/repository/acl_repository.go`
- RBAC Middleware: `/home/ec2-user/openwan/internal/api/middleware/rbac.go`
- 文件Handler: `/home/ec2-user/openwan/internal/api/handlers/files.go`
