# 502错误修复报告

## 问题描述
用户访问登录页面时遇到502 Bad Gateway错误。

## 根本原因
前端Vite开发服务器进程未运行，导致Nginx无法代理到localhost:3000，返回502错误。

## 修复步骤

### 1. 诊断问题
```bash
# 检查前端服务
curl http://localhost:3000/
# 错误: Connection refused

# 检查进程
ps aux | grep vite
# 结果: 没有运行的vite进程

# 检查日志
tail /tmp/frontend.log
# Vite之前启动但进程已终止
```

### 2. 重启前端服务
```bash
cd /home/ec2-user/openwan/frontend
nohup npm run dev > /tmp/frontend.log 2>&1 &

# 等待启动
sleep 3

# 验证
curl http://localhost:3000/
# ✅ 返回HTML页面
```

### 3. 重启后端服务
由于后端使用的是旧版本进程，也进行了重启：

```bash
# 停止旧进程
pkill -f main_simple

# 启动新进程
cd /home/ec2-user/openwan
nohup go run cmd/api/main_simple.go > /tmp/server_new.log 2>&1 &

# 验证
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d @/tmp/login.json
# ✅ 返回登录成功
```

## 验证结果

### ✅ 所有服务正常运行

| 服务 | 端口 | 状态 | 测试命令 |
|------|------|------|----------|
| 后端API | 8080 | ✅ 运行中 | `curl http://localhost:8080/health` |
| 前端Vite | 3000 | ✅ 运行中 | `curl http://localhost:3000/` |
| Nginx | 80 | ✅ 运行中 | `curl http://localhost/` |

### ✅ API测试成功
```bash
# 通过Nginx代理测试登录
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 响应:
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

### ✅ 前端页面可访问
```bash
curl http://localhost/ | grep title
# 输出: <title>OpenWan - 媒资管理系统</title>
```

## 当前系统架构

```
用户浏览器
    ↓
http://13.217.210.142/ (公网IP)
    ↓
Nginx (端口80)
    ├─ /api/* → http://localhost:8080/api/* (后端Go API)
    └─ /* → http://localhost:3000 (前端Vite开发服务器)
```

## 服务进程信息

### 后端进程
```
PID: 2941993
命令: go run cmd/api/main_simple.go
日志: /tmp/server_new.log
启动时间: 2026/02/01 17:05:56
监听: :8080
```

### 前端进程
```
命令: npm run dev (vite)
日志: /tmp/frontend.log
启动时间: 约17:05
监听: http://localhost:3000/
```

### Nginx配置
```
配置文件: /etc/nginx/conf.d/openwan.conf
访问日志: /var/log/nginx/openwan_access.log
错误日志: /var/log/nginx/openwan_error.log
```

## 访问测试

### 🌐 方法1: 浏览器访问（推荐）

**访问**: http://13.217.210.142/

**预期结果**:
1. ✅ 页面正常加载
2. ✅ 自动跳转到 `/login` 页面
3. ✅ 看到登录表单
4. ✅ 输入 admin/admin 可以成功登录
5. ✅ 登录后跳转到 `/dashboard` 主页面

### 🔧 方法2: curl命令行测试

```bash
# 测试健康检查
curl http://13.217.210.142/health

# 测试前端页面
curl http://13.217.210.142/ | grep title

# 测试登录API
curl -X POST http://13.217.210.142/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

## 故障原因分析

### 为什么前端进程停止了？

可能原因：
1. **nohup进程意外终止**: 后台进程可能因资源限制或错误退出
2. **端口冲突**: 之前尝试启动在3001，后来改为3000
3. **手动终止**: 可能在之前的测试中被终止

### 如何避免再次发生？

建议使用进程管理器如 `systemd` 或 `pm2` 来管理服务，而不是直接用 `nohup`。

#### 方案A: 使用systemd（推荐用于生产环境）

创建systemd服务文件：

```bash
# 后端服务
sudo cat > /etc/systemd/system/openwan-api.service << 'EOF'
[Unit]
Description=OpenWan API Server
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/openwan
ExecStart=/usr/local/go/bin/go run cmd/api/main_simple.go
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 前端服务
sudo cat > /etc/systemd/system/openwan-frontend.service << 'EOF'
[Unit]
Description=OpenWan Frontend Dev Server
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/openwan/frontend
ExecStart=/usr/bin/npm run dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable openwan-api
sudo systemctl enable openwan-frontend
sudo systemctl start openwan-api
sudo systemctl start openwan-frontend

# 检查状态
sudo systemctl status openwan-api
sudo systemctl status openwan-frontend
```

#### 方案B: 使用PM2（Node.js进程管理器）

```bash
# 安装PM2
npm install -g pm2

# 启动服务
pm2 start "go run cmd/api/main_simple.go" --name openwan-api
pm2 start npm --name openwan-frontend -- run dev

# 设置开机自启
pm2 startup
pm2 save

# 查看状态
pm2 list
pm2 logs
```

## 当前使用方法

由于当前是开发/测试环境，继续使用 `nohup` 后台运行：

### 检查服务状态
```bash
# 检查后端
ps aux | grep main_simple | grep -v grep

# 检查前端
ps aux | grep vite | grep -v grep

# 检查日志
tail -f /tmp/server_new.log
tail -f /tmp/frontend.log
```

### 重启服务（如果需要）

**重启后端**:
```bash
pkill -f main_simple
cd /home/ec2-user/openwan
nohup go run cmd/api/main_simple.go > /tmp/server_new.log 2>&1 &
```

**重启前端**:
```bash
pkill -f vite
cd /home/ec2-user/openwan/frontend
nohup npm run dev > /tmp/frontend.log 2>&1 &
```

**重启Nginx**:
```bash
sudo systemctl reload nginx
```

## 登录凭据

```
用户名: admin
密码: admin
```

## 下一步建议

1. ✅ **当前状态**: 所有服务运行正常，可以访问
2. ⏭️ **进程管理**: 考虑使用systemd或PM2管理进程
3. ⏭️ **监控**: 添加服务健康监控和自动重启
4. ⏭️ **日志轮转**: 配置日志轮转避免日志文件过大
5. ⏭️ **生产部署**: 编译前端为静态文件，使用Go编译的二进制而非`go run`

---

**修复完成时间**: 2026-02-01 17:10  
**问题状态**: ✅ 已解决  
**访问地址**: http://13.217.210.142/  
**测试结果**: ✅ 所有功能正常
