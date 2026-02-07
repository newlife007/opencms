# OpenWan系统快速参考

## ✅ 问题已解决 - 502错误修复完成

### 访问信息

**主应用**: http://13.217.210.142/

**登录凭据**:
- 用户名: `admin`
- 密码: `admin`

---

## 服务状态

| 服务 | 端口 | 状态 | 进程检查 |
|------|------|------|----------|
| 后端API | 8080 | ✅ 运行 | `ps aux \| grep main_simple` |
| 前端Vite | 3000 | ✅ 运行 | `ps aux \| grep vite` |
| Nginx | 80 | ✅ 运行 | `sudo systemctl status nginx` |

---

## 快速测试命令

### 测试后端
```bash
curl http://localhost:8080/health
# 预期: {"service":"openwan-api","status":"healthy","version":"1.0.0"}
```

### 测试前端
```bash
curl http://localhost:3000/ | grep title
# 预期: <title>OpenWan - 媒资管理系统</title>
```

### 测试Nginx代理
```bash
curl http://localhost/ | grep title
# 预期: <title>OpenWan - 媒资管理系统</title>
```

### 测试登录API
```bash
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
# 预期: {"success":true,"message":"Login successful",...}
```

---

## 服务管理命令

### 查看日志
```bash
# 后端日志
tail -f /tmp/server_new.log

# 前端日志
tail -f /tmp/frontend.log

# Nginx访问日志
sudo tail -f /var/log/nginx/openwan_access.log

# Nginx错误日志
sudo tail -f /var/log/nginx/openwan_error.log
```

### 重启服务

#### 重启后端
```bash
pkill -f main_simple
cd /home/ec2-user/openwan
nohup go run cmd/api/main_simple.go > /tmp/server_new.log 2>&1 &
```

#### 重启前端
```bash
pkill -f vite
cd /home/ec2-user/openwan/frontend
nohup npm run dev > /tmp/frontend.log 2>&1 &
```

#### 重启Nginx
```bash
sudo systemctl reload nginx
```

---

## 故障排查

### 如果看到502错误

1. **检查前端服务**:
   ```bash
   curl http://localhost:3000/
   ```
   - 如果失败，重启前端服务

2. **检查Nginx配置**:
   ```bash
   sudo nginx -t
   ```

3. **查看Nginx错误日志**:
   ```bash
   sudo tail -20 /var/log/nginx/openwan_error.log
   ```

### 如果登录失败

1. **检查后端服务**:
   ```bash
   curl http://localhost:8080/health
   ```
   - 如果失败，重启后端服务

2. **检查后端日志**:
   ```bash
   tail -20 /tmp/server_new.log
   ```

3. **测试直接登录**:
   ```bash
   echo '{"username":"admin","password":"admin"}' > /tmp/test.json
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d @/tmp/test.json
   ```

### 如果前端页面不显示

1. **检查浏览器控制台** (F12):
   - 查看是否有JavaScript错误
   - 查看Network标签是否有请求失败

2. **检查Vite服务**:
   ```bash
   ps aux | grep vite
   tail -20 /tmp/frontend.log
   ```

3. **清除浏览器缓存**:
   - 按 Ctrl+Shift+R (Windows/Linux)
   - 按 Cmd+Shift+R (Mac)

---

## 系统架构

```
用户浏览器
    ↓
http://13.217.210.142/ (公网)
    ↓
Nginx (端口80)
    ├─ /api/* → http://localhost:8080/api/* (Go API)
    └─ /* → http://localhost:3000 (Vite Dev Server)
```

---

## 文件位置

### 应用代码
```
/home/ec2-user/openwan/
├── cmd/api/main_simple.go    # 后端入口
├── frontend/                  # 前端代码
│   ├── src/
│   ├── package.json
│   └── .env.local            # 环境配置
└── legacy-php/               # 原PHP代码
```

### 配置文件
```
/etc/nginx/conf.d/openwan.conf    # Nginx配置
```

### 日志文件
```
/tmp/server_new.log               # 后端日志
/tmp/frontend.log                 # 前端日志
/var/log/nginx/openwan_access.log # Nginx访问日志
/var/log/nginx/openwan_error.log  # Nginx错误日志
```

### 文档
```
/home/ec2-user/openwan/
├── LOGIN_401_FIX.md                      # 401错误修复
├── FRONTEND_LOGIN_REDIRECT_FIX.md        # 登录跳转修复
├── 502_ERROR_FIX.md                      # 502错误修复
└── QUICK_REFERENCE.md                    # 本文档
```

---

## 预期行为

### 正常登录流程

1. **访问主页**: http://13.217.210.142/
2. **自动跳转**: → `/login` 页面
3. **输入凭据**: admin / admin
4. **点击登录**: 表单提交
5. **显示消息**: "登录成功"
6. **自动跳转**: → `/dashboard` 主页
7. **显示界面**: 侧边栏 + 顶栏 + 内容区

### 预期响应时间

- 页面加载: < 1秒
- API响应: < 100ms
- 登录跳转: < 500ms

---

## 已知限制

### 当前为开发模式

- ⚠️ 使用 `go run` 而非编译的二进制
- ⚠️ 使用 Vite 开发服务器 (HMR)
- ⚠️ 使用 nohup 而非进程管理器
- ⚠️ Mock数据，未连接真实数据库

### 生产部署需要

1. 编译前端: `npm run build`
2. 编译后端: `go build -o openwan-api cmd/api/main.go`
3. 使用进程管理器: systemd 或 PM2
4. 连接真实数据库
5. 配置HTTPS (TLS/SSL)
6. 配置环境变量和密钥

---

## 支持

### 相关文档
- `LOGIN_401_FIX.md` - 401认证错误修复
- `FRONTEND_LOGIN_REDIRECT_FIX.md` - 登录跳转问题修复
- `502_ERROR_FIX.md` - 502网关错误修复

### 测试账户
- 用户名: admin
- 密码: admin
- 权限: 管理员

---

**文档更新时间**: 2026-02-01 17:12  
**系统状态**: ✅ 所有服务正常运行  
**访问地址**: http://13.217.210.142/
