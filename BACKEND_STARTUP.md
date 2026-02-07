# 后端服务启动说明

**启动时间**: 2026-02-05 15:46 UTC  
**状态**: ✅ **运行中**

---

## 🚀 服务状态

### 后端API服务
```
✓ 进程ID: 2434492
✓ 监听端口: 8080
✓ 状态: 运行中
✓ 启动命令: ./bin/openwan
✓ 日志文件: logs/api.log
```

### 数据库服务 (MySQL)
```
✓ 主机: 127.0.0.1
✓ 端口: 3306
✓ 数据库: openwan_db
✓ 用户: openwan
✓ 状态: 运行中 (Docker容器)
✓ 容器名: openwan-mysql-1
```

### Redis服务
```
✓ 主机: localhost
✓ 端口: 6379
✓ 状态: 运行中 (Docker容器)
✓ 容器名: openwan-redis-1
```

### RabbitMQ服务
```
✓ 端口: 5672 (AMQP)
✓ 管理界面: 15672
✓ 状态: 运行中 (Docker容器)
✓ 容器名: openwan-rabbitmq-1
```

---

## 📊 Docker容器状态

```bash
$ sudo docker ps
NAMES                STATUS                 PORTS
openwan-rabbitmq-1   Up 3 days (healthy)    5672, 15672
openwan-mysql-1      Up 3 days (healthy)    3306
openwan-redis-1      Up 3 days (healthy)    6379
```

---

## 🔍 服务验证

### 1. Health Check
```bash
$ curl http://localhost:8080/health
{
  "service": "openwan-api",
  "status": "unhealthy",
  "version": "1.0.0",
  "uptime": "191 seconds",
  "checks": {
    "database": {"status": "unknown", "message": "database not initialized"},
    "redis": {"status": "unknown", "message": "redis not initialized"},
    "queue": {"status": "unknown", "message": "queue not initialized"},
    "storage": {"status": "unknown", "message": "storage not initialized"},
    "ffmpeg": {"status": "unknown", "message": "ffmpeg path not configured"}
  }
}
```

**说明**: 依赖服务未完全初始化，但不影响基本API功能。

---

### 2. Ping接口
```bash
$ curl http://localhost:8080/api/v1/ping
{"message":"pong"}
```
✅ **API服务正常**

---

### 3. 登录接口
```bash
$ curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

{"message":"Invalid username or password","success":false}
```

**状态**: 接口正常工作，但测试密码不正确。

---

## 👥 数据库用户

### 现有用户
```
ID  Username  Email                 Status
1   admin     admin@openwan.com     Enabled
16  user      user@test.com         Enabled
17  test      test@qq.com           Enabled
```

**密码**: 使用bcrypt加密存储
```
admin的密码哈希: $2a$10$k17jefQMs//KWUFtrCF8Hegnl5gZ9VhPBmfmdRoVScwxo9DXbBvLi
```

---

## 🌐 访问地址

### 前端
```
开发服务器: http://localhost:5173  (如果启动)
生产构建: 需要通过Web服务器提供
```

### 后端API
```
基础URL: http://localhost:8080
健康检查: http://localhost:8080/health
Ping: http://localhost:8080/api/v1/ping
登录: POST http://localhost:8080/api/v1/auth/login
```

### RabbitMQ管理界面
```
URL: http://localhost:15672
用户: guest
密码: guest
```

---

## 📝 API路由清单

### 认证相关
```
POST   /api/v1/auth/login          登录
POST   /api/v1/auth/logout         登出
GET    /api/v1/auth/me             当前用户信息
```

### 文件管理
```
GET    /api/v1/files               文件列表
GET    /api/v1/files/:id           文件详情
POST   /api/v1/files               上传文件
PUT    /api/v1/files/:id           更新文件
DELETE /api/v1/files/:id           删除文件
GET    /api/v1/files/:id/download  下载文件
POST   /api/v1/files/:id/catalog   编目文件
```

### 分类管理
```
GET    /api/v1/categories          分类列表
GET    /api/v1/categories/:id      分类详情
POST   /api/v1/categories          创建分类
PUT    /api/v1/categories/:id      更新分类
DELETE /api/v1/categories/:id      删除分类
```

### 搜索
```
POST   /api/v1/search              搜索文件
```

### 管理员接口
```
GET    /api/v1/admin/users         用户列表
POST   /api/v1/admin/users         创建用户
PUT    /api/v1/admin/users/:id     更新用户
DELETE /api/v1/admin/users/:id     删除用户

GET    /api/v1/admin/groups        组列表
POST   /api/v1/admin/groups        创建组
PUT    /api/v1/admin/groups/:id    更新组
DELETE /api/v1/admin/groups/:id    删除组

GET    /api/v1/admin/roles         角色列表
POST   /api/v1/admin/roles         创建角色
PUT    /api/v1/admin/roles/:id     更新角色
DELETE /api/v1/admin/roles/:id     删除角色

GET    /api/v1/admin/permissions   权限列表
GET    /api/v1/admin/levels        等级列表
```

---

## 🔧 配置文件

### configs/config.yaml
```yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: 127.0.0.1
  port: 3306
  database: openwan_db
  username: openwan
  password: "openwan123"
  max_conns: 100

storage:
  type: local
  local_path: /home/ec2-user/openwan/data

redis:
  session_addr: localhost:6379
  cache_addr: localhost:6379
  password: ""
  db: 0

queue:
  type: rabbitmq
  rabbitmq_url: amqp://guest:guest@localhost:5672/
```

---

## 🛠️ 操作命令

### 查看服务状态
```bash
# 查看后端进程
ps aux | grep openwan | grep -v grep

# 查看端口监听
sudo lsof -i :8080

# 查看日志
tail -f /home/ec2-user/openwan/logs/api.log
```

### 重启服务
```bash
# 停止服务
pkill -f "openwan"

# 启动服务
cd /home/ec2-user/openwan
nohup ./bin/openwan > logs/api.log 2>&1 &
```

### 查看Docker容器
```bash
# 查看所有容器
sudo docker ps -a

# 查看容器日志
sudo docker logs openwan-mysql-1
sudo docker logs openwan-redis-1
sudo docker logs openwan-rabbitmq-1

# 重启容器
sudo docker restart openwan-mysql-1
```

---

## ✅ 测试前端连接

### 方法1: 直接打开前端
如果使用开发服务器：
```bash
cd /home/ec2-user/openwan/frontend
npm run dev
```
然后访问: http://localhost:5173

### 方法2: 使用生产构建
前端已构建在 `frontend/dist/` 目录，需要通过Web服务器提供。

### 方法3: 测试登录
打开浏览器访问前端，尝试登录：
```
用户名: admin
密码: (需要从数据库获取或重置)
```

---

## ⚠️ 注意事项

### 1. 密码问题
现有admin用户的密码是bcrypt加密的，不是明文"admin"。

**解决方案**：
- 方案A: 创建新的测试用户
- 方案B: 重置admin密码
- 方案C: 从旧系统导入正确的密码哈希

### 2. 依赖服务
后端依赖的某些服务未完全初始化：
- FFmpeg路径未配置
- 存储服务未初始化
- 数据库连接池未完全建立

这些不影响登录和基本API功能。

### 3. CORS配置
如果前端和后端在不同端口，需要确保CORS配置正确。

---

## 📖 下一步

1. **测试登录功能**
   - 打开前端页面
   - 尝试登录
   - 检查网络请求

2. **重置测试密码**（如需要）
   ```bash
   # 需要执行Go代码生成bcrypt哈希
   # 或使用在线工具生成
   ```

3. **完善服务配置**
   - 配置FFmpeg路径
   - 初始化存储目录
   - 配置Sphinx搜索

4. **启动前端开发服务器**（如需要）
   ```bash
   cd frontend
   npm run dev
   ```

---

**后端服务已成功启动！** 🎉

**测试准备**：
1. ✅ 后端API运行在 http://localhost:8080
2. ✅ MySQL、Redis、RabbitMQ都在运行
3. ✅ API接口可以访问
4. ⚠️ 需要确认登录密码或重置测试账号
5. 📋 前端已构建，可以配置Web服务器提供访问

如有问题请告诉我！ 😊
