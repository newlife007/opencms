# 管理员403权限拒绝问题修复

## 修复时间：2026-02-05 13:10

## ✅ 问题已修复！

---

## 问题描述

**用户反馈**:
admin用户登录后，访问 `/api/v1/files/stats` 返回 403 Forbidden

**错误信息**:
```
GET http://13.217.210.142/api/v1/files/stats 403 (Forbidden)
Response error: AxiosError: Request failed with status code 403
```

---

## 根本原因

### 后端isAdmin检查逻辑问题

**问题代码** (auth.go第108-115行) ❌:
```go
// Check if user is admin based on roles
isAdmin := false
for _, roleName := range roleList {
    if roleName == "ADMIN" || roleName == "SYSTEM" {  // ❌ 检查大写英文
        isAdmin = true
        break
    }
}
```

**数据库中的角色名** ✅:
```
admin用户的角色: ["超级管理员", "内容管理员", "审核员"]
```

**后果**:
- 检查 `roleName == "ADMIN"` → false（实际是"超级管理员"）
- `isAdmin` 设置为 false
- 存储到session: `IsAdmin: false`
- 中间件检查权限时，管理员没有绕过
- 返回 403 Permission Denied

---

## 解决方案

### 修复isAdmin检查逻辑

**修改文件**: `/home/ec2-user/openwan/internal/api/handlers/auth.go`

**修改前** ❌:
```go
// Check if user is admin based on roles
isAdmin := false
for _, roleName := range roleList {
    if roleName == "ADMIN" || roleName == "SYSTEM" {
        isAdmin = true
        break
    }
}
```

**修改后** ✅:
```go
// Check if user is admin based on roles (case-insensitive, support Chinese)
isAdmin := false
for _, roleName := range roleList {
    roleNameUpper := strings.ToUpper(roleName)
    if roleNameUpper == "ADMIN" || roleNameUpper == "SYSTEM" || 
       roleNameUpper == "ADMINISTRATOR" || roleName == "超级管理员" {
        isAdmin = true
        break
    }
}
```

**改进点**:
1. ✅ 支持大小写不敏感（`strings.ToUpper`）
2. ✅ 支持中文角色名（"超级管理员"）
3. ✅ 支持多种英文变体（ADMIN/SYSTEM/ADMINISTRATOR）

---

## 验证测试

### 测试1: admin用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

**响应**:
```json
{
  "success": true,
  "token": "9b4ad77a-5d4d-4507-a955-481f7e7991aa",
  "user": {
    "username": "admin",
    "roles": ["超级管理员", "内容管理员", "审核员"],  ✅
    "permissions": [... 70+ permissions ...]  ✅
  }
}
```

**Cookie设置**:
```
Set-Cookie: openwan_session=9b4ad77a-5d4d-4507-a955-481f7e7991aa; 
           Path=/; Max-Age=86400; HttpOnly
```

### 测试2: 访问受保护的API

```bash
curl -H "Cookie: openwan_session=9b4ad77a-5d4d-4507-a955-481f7e7991aa" \
  http://localhost:8080/api/v1/files/stats
```

**响应** ✅:
```json
{
  "success": true,
  "data": {
    "total": 11,
    "video": 2,
    "audio": 1,
    "image": 6,
    "rich_media": 2,
    "new_count": 7,
    "published": 3,
    "pending": 1
  }
}
```

**状态码**: 200 OK ✅（不再是403！）

---

## 权限检查流程

### 完整流程图

```
1. 用户登录
   ↓
2. 获取用户角色
   roles = ["超级管理员", "内容管理员", "审核员"]
   ↓
3. 检查isAdmin
   for role in roles:
       if role.upper() == "ADMIN" or role == "超级管理员":  ✅
           isAdmin = true
   ↓
4. 存储到session
   SessionData {
       UserID: 1,
       Username: "admin",
       IsAdmin: true,  ✅ 正确设置
       ...
   }
   ↓
5. API请求
   GET /api/v1/files/stats
   ↓
6. Session中间件加载session
   c.Set("is_admin", sess.IsAdmin)  // true
   ↓
7. 权限中间件检查
   isAdmin := c.Get("is_admin")
   if isAdmin == true:
       c.Next()  ✅ 绕过权限检查
       return
   ↓
8. 返回200 OK ✅
```

---

## 编译和部署

### 1. 编译后端

```bash
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api
```

**状态**: ✅ 编译成功（已添加strings包）

### 2. 重启后端服务

```bash
pkill -f "bin/openwan"
cd /home/ec2-user/openwan
./bin/openwan > /tmp/openwan.log 2>&1 &
```

**状态**: ✅ 服务已重启

### 3. 验证服务

```bash
curl http://localhost:8080/health
```

**响应**:
```json
{
  "status": "healthy",
  "database": "ok",
  "redis": "ok",
  "storage": "ok"
}
```

---

## 前端对接说明

### Cookie名称

**后端设置的cookie**: `openwan_session`

前端需要确保请求中包含此cookie。

### 检查前端配置

#### axios配置

**文件**: `frontend/src/api/request.js`

**应该包含**:
```javascript
// 允许携带cookie
withCredentials: true
```

**或者手动设置cookie**:
```javascript
axios.defaults.headers.common['Cookie'] = `openwan_session=${token}`;
```

### 测试步骤

1. **清除浏览器缓存和Cookie**
   - 按 `Ctrl+Shift+Delete`
   - 清除所有缓存和Cookie

2. **使用admin登录**
   - 用户名: admin
   - 密码: admin123

3. **检查浏览器开发者工具**
   - Application → Cookies
   - 应该看到 `openwan_session` cookie ✅

4. **查看网络请求**
   - Network → Headers
   - Request Headers应该包含:
     ```
     Cookie: openwan_session=xxx
     ```

5. **验证API响应**
   - GET /api/v1/files/stats 应该返回 200 ✅
   - 不再是 403 ❌

---

## 支持的管理员角色名

### 英文（不区分大小写）

| 原始名 | 匹配后 |
|--------|--------|
| Admin | ADMIN ✅ |
| admin | ADMIN ✅ |
| ADMIN | ADMIN ✅ |
| System | SYSTEM ✅ |
| SYSTEM | SYSTEM ✅ |
| Administrator | ADMINISTRATOR ✅ |
| administrator | ADMINISTRATOR ✅ |

### 中文（精确匹配）

| 角色名 | 匹配 |
|--------|------|
| 超级管理员 | ✅ |
| 系统管理员 | ❌ (需要添加) |
| 管理员 | ❌ (需要添加) |

**如需添加更多中文角色名**:
```go
if roleNameUpper == "ADMIN" || roleNameUpper == "SYSTEM" || 
   roleNameUpper == "ADMINISTRATOR" || 
   roleName == "超级管理员" || roleName == "系统管理员" {  // 添加这里
    isAdmin = true
}
```

---

## 相关文件

### 后端修改

| 文件 | 改动 | 行数 |
|------|------|------|
| `internal/api/handlers/auth.go` | 导入strings包 | 第4行 |
| `internal/api/handlers/auth.go` | 修改isAdmin检查 | 第108-116行 |

### 权限检查中间件

| 文件 | 功能 |
|------|------|
| `internal/api/middleware/rbac.go` | RequirePermission中间件 |
| `internal/api/middleware/rbac.go` | RequireAdmin中间件 |

**管理员绕过逻辑** (rbac.go第71-74行):
```go
// Admin bypass - admins have all permissions
isAdmin, _ := c.Get("is_admin")
if isAdmin != nil && isAdmin.(bool) {
    c.Next()  // ✅ 绕过权限检查
    return
}
```

---

## 故障排查

### 问题1: 仍然返回403

**可能原因**:
1. Session cookie未发送
2. Session已过期
3. 后端服务未重启

**解决**:
```bash
# 检查后端是否运行
ps aux | grep "bin/openwan"

# 重启后端
pkill -f "bin/openwan"
cd /home/ec2-user/openwan
./bin/openwan &

# 清除前端缓存
浏览器: Ctrl+Shift+Delete
```

### 问题2: Cookie未设置

**可能原因**:
- axios配置未启用 withCredentials
- 跨域问题

**解决**:
```javascript
// axios配置
axios.defaults.withCredentials = true;

// 或每个请求
axios.get(url, { withCredentials: true })
```

### 问题3: Session加载失败

**检查Redis**:
```bash
redis-cli
> KEYS openwan:session:*
> GET "openwan:session:9b4ad77a-5d4d-4507-a955-481f7e7991aa"
```

**应该看到session数据**:
```json
{
  "user_id": 1,
  "username": "admin",
  "is_admin": true,  ← 应该是true
  ...
}
```

---

## 相关文档

- **权限格式修复**: `/home/ec2-user/openwan/docs/PERMISSION-FORMAT-FIX.md`
- **前端权限控制**: `/home/ec2-user/openwan/docs/FRONTEND-PERMISSION-CONTROL.md`
- **API权限加固**: `/home/ec2-user/openwan/docs/API-PERMISSION-HARDENING.md`

---

## 总结

### 问题

❌ admin角色名不匹配：检查"ADMIN"但实际是"超级管理员"

### 原因

后端只检查大写英文角色名，不支持中文

### 解决

✅ 添加中文角色名支持 + 大小写不敏感

### 验证

✅ admin登录成功

✅ API返回200（不再403）

✅ Session中is_admin=true

---

**修复完成时间**: 2026-02-05 13:10  
**修复人员**: AWS Transform CLI  
**版本**: 5.3 Admin Role Fix  
**状态**: ✅ **已修复并测试通过**

**下一步**: 清除浏览器缓存后重新测试前端！
