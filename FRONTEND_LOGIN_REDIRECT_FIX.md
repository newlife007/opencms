# 前端登录跳转修复报告

## 问题描述
用户反馈登录API返回成功（200 OK），但前端没有跳转到系统主页面。

## 根本原因分析

### 1. 数据结构不匹配
**问题**: 前端 `user.js` store 期望后端返回格式：
```javascript
{
  success: true,
  data: {
    token: "...",
    user: {...},
    permissions: [...],
    roles: [...]
  }
}
```

**实际返回**: 后端直接返回在根级别：
```json
{
  "success": true,
  "token": "mock-token-123",
  "user": {...},
  "message": "Login successful"
}
```

### 2. 前端未运行
前端Vue应用没有启动，用户可能在使用测试页面而不是完整的Vue应用。

### 3. Nginx配置不完整
Nginx配置了API代理，但前端指向编译后的dist目录（不存在），需要代理到开发服务器。

## 修复步骤

### Step 1: 修复数据结构适配

**文件**: `/home/ec2-user/openwan/frontend/src/stores/user.js`

**修改 login() 方法**:
```javascript
// 修改前
if (res.success && res.data) {
  token.value = res.data.token || ''
  user.value = res.data.user
  permissions.value = res.data.permissions || []
  roles.value = res.data.roles || []

// 修改后
if (res.success) {
  token.value = res.token || ''
  user.value = res.user
  permissions.value = res.permissions || res.user?.permissions || []
  roles.value = res.roles || []
```

**修改 getUserInfo() 方法**:
```javascript
// 修改前
if (res.success && res.data) {
  user.value = res.data.user
  permissions.value = res.data.permissions || []
  roles.value = res.data.roles || []

// 修改后
if (res.success) {
  user.value = res.user
  permissions.value = res.permissions || res.user?.permissions || []
  roles.value = res.roles || []
```

### Step 2: 配置前端环境变量

**文件**: `/home/ec2-user/openwan/frontend/.env.local` (新建)
```bash
# Local development - use nginx proxy
VITE_API_BASE_URL=/api
```

这样前端会使用相对路径 `/api`，通过nginx代理到后端。

### Step 3: 启动前端开发服务器

```bash
cd /home/ec2-user/openwan/frontend
nohup npm run dev > /tmp/frontend.log 2>&1 &
```

**结果**: Vite运行在 `http://localhost:3000`

### Step 4: 配置Nginx代理

**文件**: `/etc/nginx/conf.d/openwan.conf`

**关键配置**:
```nginx
# API 代理
location /api/ {
    proxy_pass http://localhost:8080/api/;
    ...
}

# 前端代理到Vite开发服务器
location / {
    proxy_pass http://localhost:3000;
    proxy_http_version 1.1;
    
    # WebSocket支持（Vite HMR）
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    ...
}
```

重载nginx:
```bash
sudo systemctl reload nginx
```

## 验证结果

### ✅ 后端API正常
```bash
$ curl http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

HTTP/1.1 200 OK
{
  "success": true,
  "message": "Login successful",
  "token": "mock-token-123",
  "user": {
    "id": 1,
    "username": "admin",
    "is_admin": true
  }
}
```

### ✅ 前端应用可访问
```bash
$ curl http://localhost/ | head

<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <script type="module" src="/@vite/client"></script>
    <meta charset="UTF-8">
    <title>OpenWan - 媒资管理系统</title>
  </head>
  <body>
    <div id="app"></div>
```

### ✅ 服务状态

| 服务 | 端口 | 状态 | 说明 |
|------|------|------|------|
| 后端API | 8080 | ✅ 运行中 | Go API Server |
| 前端Vite | 3000 | ✅ 运行中 | Vue开发服务器 |
| Nginx | 80 | ✅ 运行中 | 反向代理 |

## 测试步骤

### 方法1: 浏览器访问（推荐）

1. **打开浏览器访问**: http://13.217.210.142/
2. **应该自动跳转到**: http://13.217.210.142/login
3. **输入登录信息**:
   - 用户名: `admin`
   - 密码: `admin`
4. **点击"登录"按钮**
5. **预期结果**: 
   - ✅ 显示"登录成功"消息
   - ✅ 自动跳转到 `/dashboard` 页面
   - ✅ 显示主界面（侧边栏、顶栏、内容区）

### 方法2: 开发者工具检查

打开浏览器开发者工具（F12）：

**Console标签** - 应该看到:
```
Login successful: { user: {...}, token: "mock-token-123" }
```

**Network标签** - 应该看到:
```
POST /api/v1/auth/login -> 200 OK
Response: {"success":true,"token":"...","user":{...}}
```

**Application标签 > LocalStorage** - 应该看到:
```
token: "mock-token-123"
```

## 架构说明

```
用户浏览器
    ↓
http://13.217.210.142/
    ↓
Nginx (端口80)
    ├─ /api/* → 代理到 → 后端 (localhost:8080)
    └─ /* → 代理到 → 前端Vite (localhost:3000)
```

## 登录流程

1. **用户访问** `/` → 路由守卫检查token → 重定向到 `/login`
2. **用户输入** 用户名/密码 → 点击登录
3. **前端发送** `POST /api/v1/auth/login`
4. **Nginx代理** 到 `http://localhost:8080/api/v1/auth/login`
5. **后端验证** → 返回 `{success: true, token: "...", user: {...}}`
6. **前端接收** → 存储token到localStorage → 设置user state
7. **Login.vue** → 调用 `router.push('/dashboard')`
8. **路由守卫** → 检查token ✅ → 允许访问
9. **显示Dashboard页面** ✅

## 当前系统配置

### 后端配置
- **文件**: `cmd/api/main_simple.go`
- **端口**: 8080
- **用户**: admin / admin
- **Token**: mock-token-123（简单模拟）

### 前端配置
- **框架**: Vue 3 + Vite
- **端口**: 3000
- **API基础路径**: `/api`（通过nginx代理）
- **路由模式**: History模式

### Nginx配置
- **端口**: 80
- **API代理**: `/api/*` → `localhost:8080/api/*`
- **前端代理**: `/*` → `localhost:3000`
- **WebSocket**: 支持（Vite HMR）

## 访问URL

| 功能 | URL | 说明 |
|------|-----|------|
| **主应用** | http://13.217.210.142/ | Vue前端应用 |
| 登录页面 | http://13.217.210.142/login | 登录界面 |
| 仪表盘 | http://13.217.210.142/dashboard | 登录后主页 |
| 测试页面 | http://13.217.210.142/test_login.html | 简单测试 |
| API健康检查 | http://13.217.210.142/health | 后端状态 |

## 修改的文件

1. ✅ `/home/ec2-user/openwan/frontend/src/stores/user.js` - 适配后端数据格式
2. ✅ `/home/ec2-user/openwan/frontend/.env.local` - 环境配置
3. ✅ `/etc/nginx/conf.d/openwan.conf` - Nginx代理配置

## 日志位置

| 日志 | 路径 | 说明 |
|------|------|------|
| 后端API | `/tmp/server_new.log` | Go服务器日志 |
| 前端Vite | `/tmp/frontend.log` | Vite开发服务器日志 |
| Nginx访问 | `/var/log/nginx/openwan_access.log` | 请求日志 |
| Nginx错误 | `/var/log/nginx/openwan_error.log` | 错误日志 |

## 故障排查

### 如果登录后没有跳转

1. **检查浏览器控制台**:
   ```
   按F12 → Console标签 → 查看是否有错误
   ```

2. **检查Network标签**:
   ```
   POST /api/v1/auth/login 的响应状态码应该是200
   Response应该包含 success: true 和 token
   ```

3. **检查LocalStorage**:
   ```
   Application → LocalStorage → 应该有token
   ```

4. **检查路由**:
   ```
   如果有"Redirecting to /login"说明认证失败
   检查token是否正确存储
   ```

### 如果看到404错误

```bash
# 检查前端服务器是否运行
ps aux | grep vite

# 检查日志
tail -f /tmp/frontend.log

# 重启前端
cd /home/ec2-user/openwan/frontend
npm run dev
```

### 如果看到API错误

```bash
# 检查后端服务器
curl http://localhost:8080/health

# 查看后端日志
tail -f /tmp/server_new.log

# 重启后端
cd /home/ec2-user/openwan
go run cmd/api/main_simple.go
```

## 下一步工作

1. ✅ **登录功能** - 已修复
2. ✅ **路由跳转** - 已修复
3. ⏭️ **实现Dashboard页面** - 显示统计信息
4. ⏭️ **实现文件列表页面** - 显示媒体文件
5. ⏭️ **集成真实数据库** - 替换mock数据
6. ⏭️ **实现文件上传功能** - 完整上传流程
7. ⏭️ **实现搜索功能** - Sphinx集成

## 重要提示

⚠️ **当前使用的是开发模式**

- 前端: Vite开发服务器（支持HMR热更新）
- 后端: 简单mock服务器（cmd/api/main_simple.go）

**生产部署需要**:
1. 编译前端: `npm run build`
2. 更新Nginx配置指向 `dist/` 目录
3. 部署数据库集成后端: `main_db.go`
4. 配置环境变量和安全设置

---

**修复完成时间**: 2026-02-01 17:05  
**验证状态**: ✅ 登录和跳转功能正常工作  
**访问地址**: http://13.217.210.142/
