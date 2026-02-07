# OpenWan Backend Deployment Status

**部署时间**: 2026-02-05 13:50 UTC  
**状态**: ✅ **运行中**

---

## 服务信息

### 进程信息
- **PID**: 2333289
- **端口**: 8080
- **启动时间**: 13:50
- **CPU使用**: 0.0%
- **内存使用**: 0.1%

### 监听端口
```
tcp6  0  0  :::8080  :::*  LISTEN  PID:2333289
```

### 服务端点
- **健康检查**: http://localhost:8080/health
- **就绪检查**: http://localhost:8080/ready
- **存活检查**: http://localhost:8080/alive
- **API基础**: http://localhost:8080/api/v1/
- **登录接口**: http://localhost:8080/api/v1/auth/login
- **文件列表**: http://localhost:8080/api/v1/files

---

## 服务初始化状态

### ✅ 已初始化组件
1. ✓ Database connected (MySQL 127.0.0.1:3306/openwan_db)
2. ✓ Redis session store connected (localhost:6379)
3. ✓ Storage service initialized (local mode)
4. ✓ Repositories initialized
5. ✓ Services initialized
6. ✓ Router configured

### ⚠️ 未初始化组件
- Queue (RabbitMQ) - 转码队列，非关键功能
- FFmpeg - 视频转码，需要配置路径

---

## API路由 (已注册)

### 认证模块
- POST   /api/v1/auth/login
- POST   /api/v1/auth/logout
- GET    /api/v1/auth/me
- POST   /api/v1/auth/refresh
- PUT    /api/v1/auth/profile
- POST   /api/v1/auth/change-password

### 文件模块
- GET    /api/v1/files
- GET    /api/v1/files/stats
- GET    /api/v1/files/recent
- GET    /api/v1/files/:id
- POST   /api/v1/files (upload)
- PUT    /api/v1/files/:id
- DELETE /api/v1/files/:id
- GET    /api/v1/files/:id/download
- GET    /api/v1/files/:id/preview

### 工作流模块
- POST   /api/v1/files/:id/submit
- POST   /api/v1/files/:id/publish
- POST   /api/v1/files/:id/reject
- PUT    /api/v1/files/:id/status

### 分类模块
- GET    /api/v1/categories
- GET    /api/v1/categories/tree
- GET    /api/v1/categories/:id
- POST   /api/v1/categories
- PUT    /api/v1/categories/:id
- DELETE /api/v1/categories/:id

### 目录配置模块
- GET    /api/v1/catalog
- GET    /api/v1/catalog/tree
- GET    /api/v1/catalog/all
- GET    /api/v1/catalog/:id
- POST   /api/v1/catalog
- PUT    /api/v1/catalog/:id
- DELETE /api/v1/catalog/:id

### 搜索模块
- GET    /api/v1/search
- POST   /api/v1/search
- GET    /api/v1/search/suggestions
- POST   /api/v1/search/reindex
- GET    /api/v1/search/status

### 管理员模块

#### 用户管理
- GET    /api/v1/admin/users
- POST   /api/v1/admin/users/search
- GET    /api/v1/admin/users/:id
- POST   /api/v1/admin/users
- PUT    /api/v1/admin/users/:id
- DELETE /api/v1/admin/users/:id
- POST   /api/v1/admin/users/batch-delete
- POST   /api/v1/admin/users/:id/reset-password
- PUT    /api/v1/admin/users/:id/status
- GET    /api/v1/admin/users/:id/permissions

#### 组管理
- GET    /api/v1/admin/groups
- POST   /api/v1/admin/groups
- GET    /api/v1/admin/groups/:id
- PUT    /api/v1/admin/groups/:id
- DELETE /api/v1/admin/groups/:id
- POST   /api/v1/admin/groups/:id/categories
- POST   /api/v1/admin/groups/:id/roles

#### 角色管理
- GET    /api/v1/admin/roles
- POST   /api/v1/admin/roles
- GET    /api/v1/admin/roles/:id
- PUT    /api/v1/admin/roles/:id
- DELETE /api/v1/admin/roles/:id
- POST   /api/v1/admin/roles/:id/permissions

#### 权限管理
- GET    /api/v1/admin/permissions
- GET    /api/v1/admin/permissions/:id

#### 等级管理
- GET    /api/v1/admin/levels
- GET    /api/v1/admin/levels/:id
- POST   /api/v1/admin/levels
- PUT    /api/v1/admin/levels/:id
- DELETE /api/v1/admin/levels/:id

#### 工作流统计
- GET    /api/v1/admin/workflow/stats

---

## 日志查看

### 实时日志
```bash
tail -f /home/ec2-user/openwan/logs/api.log
```

### 最近日志
```bash
tail -100 /home/ec2-user/openwan/logs/api.log
```

### 错误日志
```bash
grep -i "error\|failed\|panic" /home/ec2-user/openwan/logs/api.log
```

---

## 服务管理命令

### 查看服务状态
```bash
ps aux | grep bin/openwan | grep -v grep
```

### 停止服务
```bash
pkill -f "bin/openwan"
```

### 启动服务 (手动)
```bash
cd /home/ec2-user/openwan
nohup ./bin/openwan > logs/api.log 2>&1 &
```

### 重启服务 (使用脚本)
```bash
cd /home/ec2-user/openwan
./deploy-backend.sh
```

---

## 测试命令

### 测试健康检查
```bash
curl http://localhost:8080/health | jq .
```

### 测试登录 (需要在服务器上执行)
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 测试文件列表 (需要登录后的cookie)
```bash
curl http://localhost:8080/api/v1/files \
  -H "Cookie: openwan_session=YOUR_SESSION_ID"
```

---

## 配置信息

### 数据库
- **主机**: 127.0.0.1:3306
- **数据库**: openwan_db
- **用户**: root
- **连接池**: 最大100连接

### Redis
- **主机**: localhost:6379
- **用途**: 会话存储 + 缓存

### 存储
- **类型**: Local (本地文件系统)
- **路径**: /home/ec2-user/openwan/data

---

## 最近修复记录

### 2026-02-05 14:00 - 等级管理逻辑修复
- ✅ Levels模型：Weight改为Level
- ✅ ACL检查：正确比较level数值
- ✅ 前端UI：权重→级别，范围1-10
- ✅ 数据库迁移脚本创建

### 2026-02-05 13:40 - 权限格式修复
- ✅ hasPermission支持字符串格式
- ✅ 移除admin特殊bypass逻辑
- ✅ 实现正确的权限设计

### 2026-02-05 13:25 - Admin角色识别修复
- ✅ 后端支持中文角色名"超级管理员"
- ✅ Admin API从403改为200

---

## 下一步

1. **测试登录功能** - 从浏览器访问前端，测试登录
2. **验证所有修复** - 确认等级管理、权限检查正常工作
3. **监控日志** - 观察是否有错误或警告
4. **性能测试** - 测试并发访问和响应时间

---

**文档创建时间**: 2026-02-05 14:00 UTC  
**服务状态**: ✅ **健康运行**
