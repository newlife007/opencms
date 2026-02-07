# Auth.go 逐步修复指南

由于工具限制，我已经准备好完整的修复说明。您只需要按照以下步骤操作：

## 方案A：使用VI编辑器手动修改（推荐！最可靠）

### 1. 打开文件
```bash
vi /home/ec2-user/openwan/internal/api/handlers/auth.go
```

### 2. 找到Login函数中的权限获取部分（约第76-85行）

找到这段代码：
```go
// Get user permissions
permissions, err := h.aclService.GetUserPermissions(c.Request.Context(), user.ID)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "success": false,
        "message": "Failed to load user permissions",
    })
    return
}
```

**在这段代码的后面添加（按'o'进入插入模式）：**
```go

// Get user roles
roles, err := h.aclService.GetUserRoles(c.Request.Context(), user.ID)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "success": false,
        "message": "Failed to load user roles",
    })
    return
}
```

### 3. 找到构建权限列表的部分（约第110-115行）

找到这段代码：
```go
// Build permission list first
permList := make([]string, len(permissions))
for i, p := range permissions {
    permList[i] = fmt.Sprintf("%s.%s.%s", p.Namespace, p.Controller, p.Action)
}
sess.Permissions = permList
```

**在 `sess.Permissions = permList` 这行后面添加：**
```go

// Build role list
roleList := make([]string, len(roles))
for i, r := range roles {
    roleList[i] = r.Name
}

// Check if user is admin based on roles
isAdmin := false
for _, roleName := range roleList {
    if roleName == "ADMIN" || roleName == "SYSTEM" {
        isAdmin = true
        break
    }
}

// Update session IsAdmin field
sess.IsAdmin = isAdmin
```

### 4. 修改SessionData创建部分

找到这行（约第104行）：
```go
IsAdmin:     user.Username == "admin",
```

**改为：**
```go
IsAdmin:     false,  // Will be set after role check
```

### 5. 找到Login响应部分（约第170行）

找到这段代码：
```go
User: &UserInfo{
    ID:          user.ID,
    Username:    user.Username,
    Email:       user.Email,
    GroupID:     user.GroupID,
    LevelID:     user.LevelID,
    Permissions: permList,
},
```

**在 `Permissions: permList,` 后面添加一行：**
```go
    Roles:       roleList,
```

### 6. 修改GetCurrentUser函数（约第240-270行）

找到GetCurrentUser中的权限获取后面，添加角色获取：

在这段之后：
```go
permList := make([]string, len(permissions))
for i, p := range permissions {
    permList[i] = fmt.Sprintf("%s.%s.%s", p.Namespace, p.Controller, p.Action)
}
```

**添加：**
```go

// Get user roles
roles, err := h.aclService.GetUserRoles(c.Request.Context(), user.ID)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "success": false,
        "message": "Failed to load user roles",
    })
    return
}

roleList := make([]string, len(roles))
for i, r := range roles {
    roleList[i] = r.Name
}
```

然后找到GetCurrentUser的JSON响应：
```go
"user": UserInfo{
    ID:          user.ID,
    Username:    user.Username,
    Email:       user.Email,
    GroupID:     user.GroupID,
    LevelID:     user.LevelID,
    Permissions: permList,
},
```

**在 `Permissions: permList,` 后面添加：**
```go
    Roles:       roleList,
```

### 7. 保存并退出
按 `ESC`，然后输入 `:wq` 并回车

---

## 方案B：使用sed命令批量替换（快速但需要小心）

```bash
cd /home/ec2-user/openwan

# 备份当前文件
cp internal/api/handlers/auth.go internal/api/handlers/auth.go.before_manual_fix

# 方案B的具体命令我会在下一步提供
```

---

## 验证修改

完成修改后，运行以下命令验证：

```bash
cd /home/ec2-user/openwan

# 1. 检查语法
go fmt ./internal/api/handlers/auth.go

# 2. 编译
go build -o bin/openwan .

# 3. 如果编译成功，启动服务器测试
# （根据您的启动方式）
```

---

## 如果遇到问题

1. **编译失败：** 检查括号是否匹配，缩进是否正确
2. **语法错误：** 运行 `go fmt` 会给出具体行号
3. **变量未定义：** 确保 `roleList` 和 `isAdmin` 在正确的作用域内

---

## 快速替换文件方案（最简单！）

如果您觉得手动编辑太复杂，我可以为您生成一个完整的正确文件。

**请告诉我：**
"请生成完整的auth.go文件"

然后我会创建一个完整文件，您只需：
```bash
cp /path/to/generated/auth.go /home/ec2-user/openwan/internal/api/handlers/auth.go
```

一步完成！

---

**您希望使用哪种方案？**
- 方案A（VI手动编辑）- 最可靠
- 方案B（提供替换脚本）- 快速
- 生成完整文件 - 最简单
