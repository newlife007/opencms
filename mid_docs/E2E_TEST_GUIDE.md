# OpenWan 端到端测试指南

## 🎯 环境状态

### ✅ 所有服务已启动

| 服务 | 状态 | 地址 | PID/容器 |
|------|------|------|---------|
| MySQL | ✅ 运行中 | localhost:3306 | openwan-mysql-1 |
| Redis | ✅ 运行中 | localhost:6379 | openwan-redis-1 |
| RabbitMQ | ✅ 运行中 | localhost:5672, 15672 | openwan-rabbitmq-1 |
| 后端API | ✅ 运行中 | http://localhost:8080 | PID: 3084105 |
| 前端Dev | ✅ 运行中 | http://localhost:3000 | PID: 3074835 |

---

## 📋 测试账号

已在数据库中创建以下测试用户：

| 用户名 | 密码 | 角色 | 用途 |
|--------|------|------|------|
| admin | admin | 超级管理员 | 完整管理权限测试 |
| yc75 | (从数据库查看) | 普通用户 | 普通用户权限测试 |

### 获取admin密码

```bash
# 查看admin用户的密码哈希
mysql -h 127.0.0.1 -P 3306 -u root -prootpassword openwan_db \
  -e "SELECT username, password FROM ow_users WHERE username='admin';"

# 输出：admin  $1$kI0.dK0.$mZfeLOhcTZ.xHq5uw8fk3.
# 这是 'admin' 的MD5 crypt哈希
```

---

## 🧪 端到端测试步骤

### 第1步：基础连接测试

#### 1.1 后端API健康检查

```bash
# 健康检查端点
curl http://localhost:8080/health | jq

# 预期响应：
# {
#   "service": "openwan-api",
#   "status": "unhealthy" 或 "healthy",
#   "version": "1.0.0",
#   "uptime": "xxx seconds",
#   "checks": { ... }
# }
```

#### 1.2 API Ping测试

```bash
# Ping端点
curl http://localhost:8080/api/v1/ping

# 预期响应：
# {"message":"pong"}
```

#### 1.3 前端访问测试

```bash
# 访问前端首页
curl -I http://localhost:3000/

# 预期响应：
# HTTP/1.1 200 OK
```

**浏览器访问**：
- 打开 http://localhost:3000
- 应该看到OpenWan登录页面

---

### 第2步：用户认证测试

#### 2.1 测试未认证访问（应返回401）

```bash
# 尝试访问需要认证的端点
curl http://localhost:8080/api/v1/admin/users

# 预期响应：
# {
#   "success": false,
#   "message": "Authentication required",
#   "error": "No session cookie found"
# }
# HTTP状态码：401
```

#### 2.2 用户登录测试

```bash
# 方式1：使用curl（保存Cookie）
curl -c /tmp/cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 预期响应：
# {
#   "success": true,
#   "message": "Login successful",
#   "user": {
#     "id": 1,
#     "username": "admin",
#     "email": "thinkgem@gmail.com",
#     "group_id": 1,
#     "level_id": 1,
#     "permissions": [...]
#   }
# }

# 方式2：浏览器测试
# 1. 访问 http://localhost:3000
# 2. 在登录页面输入：
#    用户名：admin
#    密码：admin
# 3. 点击登录按钮
# 4. 应该跳转到仪表板页面
```

#### 2.3 测试认证后访问

```bash
# 使用保存的Cookie访问保护端点
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/admin/users

# 预期响应：用户列表（JSON数组）
```

#### 2.4 测试当前用户信息

```bash
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/auth/me

# 预期响应：当前登录用户的详细信息
```

#### 2.5 测试登出

```bash
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/auth/logout

# 预期响应：
# {
#   "success": true,
#   "message": "Logout successful"
# }
```

---

### 第3步：前端界面测试

使用浏览器访问 `http://localhost:3000`，按顺序测试以下功能：

#### 3.1 登录流程

1. ✅ 打开登录页面
2. ✅ 输入用户名：`admin`
3. ✅ 输入密码：`admin`
4. ✅ 点击"登录"按钮
5. ✅ 验证跳转到仪表板
6. ✅ 验证顶部显示用户名

#### 3.2 仪表板页面

1. ✅ 查看欢迎信息
2. ✅ 查看统计卡片（文件数、分类数、用户数等）
3. ✅ 验证左侧导航菜单显示

#### 3.3 文件管理

**文件列表**（路由：`/files`）
1. ✅ 点击左侧菜单"文件管理"
2. ✅ 验证文件列表页面加载
3. ✅ 测试筛选功能（文件类型、状态、分类）
4. ✅ 测试搜索功能
5. ✅ 测试分页功能
6. ✅ 测试排序功能（按标题、日期、大小）
7. ✅ 切换视图模式（列表/网格）

**文件上传**（路由：`/files/upload`）
1. ✅ 点击"上传文件"按钮
2. ✅ 测试拖拽上传区域
3. ✅ 选择文件上传
4. ✅ 填写文件元数据（标题、描述、分类）
5. ✅ 提交上传
6. ✅ 查看上传进度
7. ✅ 验证上传成功消息

**文件详情**（路由：`/files/:id`）
1. ✅ 在文件列表点击某个文件
2. ✅ 查看文件详细信息
3. ✅ 验证文件预览（图片/视频）
4. ✅ 测试下载按钮
5. ✅ 测试编辑按钮

#### 3.4 搜索功能

**搜索页面**（路由：`/search`）
1. ✅ 点击左侧菜单"搜索"
2. ✅ 在搜索框输入关键词
3. ✅ 测试高级筛选（类型、状态、日期范围）
4. ✅ 查看搜索结果
5. ✅ 验证结果高亮显示
6. ✅ 测试分页
7. ✅ 点击搜索结果查看详情

#### 3.5 管理员功能

**用户管理**（路由：`/admin/users`）
1. ✅ 点击左侧菜单"管理" → "用户管理"
2. ✅ 查看用户列表
3. ✅ 点击"添加用户"按钮
4. ✅ 填写用户信息表单
5. ✅ 提交创建用户
6. ✅ 测试编辑用户
7. ✅ 测试删除用户（带确认对话框）
8. ✅ 测试重置密码功能

**组管理**（路由：`/admin/groups`）
1. ✅ 点击"组管理"
2. ✅ 查看组列表
3. ✅ 创建新组
4. ✅ 分配用户到组
5. ✅ 分配角色到组
6. ✅ 分配分类访问权限

**角色管理**（路由：`/admin/roles`）
1. ✅ 点击"角色管理"
2. ✅ 查看角色列表
3. ✅ 创建新角色
4. ✅ 分配权限到角色
5. ✅ 编辑角色信息

**分类管理**（路由：`/admin/categories`）
1. ✅ 点击"分类管理"
2. ✅ 查看分类树形结构
3. ✅ 展开/折叠分类节点
4. ✅ 添加根分类
5. ✅ 添加子分类
6. ✅ 编辑分类信息
7. ✅ 拖拽排序分类
8. ✅ 删除分类（带确认）

**目录配置**（路由：`/admin/catalog`）
1. ✅ 点击"目录配置"
2. ✅ 查看目录字段列表
3. ✅ 添加新字段
4. ✅ 设置字段类型（文本、数字、日期、下拉等）
5. ✅ 设置必填字段
6. ✅ 排序字段顺序
7. ✅ 预览表单效果

---

### 第4步：API端点功能测试

#### 4.1 文件管理API

```bash
# 获取文件列表
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files

# 获取单个文件详情
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files/1

# 更新文件信息
curl -b /tmp/cookies.txt -X PUT http://localhost:8080/api/v1/files/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"更新后的标题"}'

# 删除文件
curl -b /tmp/cookies.txt -X DELETE http://localhost:8080/api/v1/files/1
```

#### 4.2 分类管理API

```bash
# 获取分类列表（树形）
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/categories

# 创建分类
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"新分类","description":"测试分类"}'

# 更新分类
curl -b /tmp/cookies.txt -X PUT http://localhost:8080/api/v1/categories/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"更新分类名称"}'
```

#### 4.3 搜索API

```bash
# GET方式搜索
curl -b /tmp/cookies.txt "http://localhost:8080/api/v1/search?q=测试&type=1&page=1&page_size=20"

# POST方式搜索（带高级筛选）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "测试",
    "filters": {
      "type": [1, 2],
      "status": 2,
      "category_id": 1
    },
    "page": 1,
    "page_size": 20
  }'
```

#### 4.4 用户管理API

```bash
# 获取用户列表
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/admin/users

# 创建用户
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser2",
    "password": "password123",
    "email": "testuser2@example.com",
    "group_id": 3,
    "level_id": 1
  }'

# 更新用户
curl -b /tmp/cookies.txt -X PUT http://localhost:8080/api/v1/admin/users/8 \
  -H "Content-Type: application/json" \
  -d '{"nickname":"新昵称"}'

# 删除用户
curl -b /tmp/cookies.txt -X DELETE http://localhost:8080/api/v1/admin/users/8
```

---

### 第5步：工作流测试

#### 5.1 文件工作流状态

测试文件状态转换：
- 0: 新建（new）
- 1: 待审核（pending）
- 2: 已发布（published）
- 3: 已拒绝（rejected）
- 4: 已删除（deleted）

```bash
# 提交文件审核（new → pending）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/files/1/submit

# 发布文件（pending → published）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/files/1/publish

# 拒绝文件（pending → rejected）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/files/1/reject \
  -H "Content-Type: application/json" \
  -d '{"reason":"不符合要求"}'

# 直接更新状态
curl -b /tmp/cookies.txt -X PUT http://localhost:8080/api/v1/files/1/status \
  -H "Content-Type: application/json" \
  -d '{"status":2}'

# 获取工作流统计
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/admin/workflow/stats
```

---

### 第6步：权限测试

#### 6.1 测试不同角色权限

```bash
# 使用admin用户登录
curl -c /tmp/admin_cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 使用普通用户登录
curl -c /tmp/user_cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"yc75","password":"<从数据库获取>"}'

# admin应该可以访问用户管理
curl -b /tmp/admin_cookies.txt http://localhost:8080/api/v1/admin/users
# 预期：返回用户列表

# 普通用户应该不能访问用户管理
curl -b /tmp/user_cookies.txt http://localhost:8080/api/v1/admin/users
# 预期：返回403 Forbidden
```

---

### 第7步：错误处理测试

#### 7.1 测试错误响应

```bash
# 404 Not Found
curl http://localhost:8080/api/v1/files/999999

# 400 Bad Request（缺少必填字段）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin"}'

# 422 Validation Error（字段验证失败）
curl -b /tmp/cookies.txt -X POST http://localhost:8080/api/v1/admin/users \
  -H "Content-Type: application/json" \
  -d '{"username":"a","password":"123"}'

# 429 Rate Limited（如果实现了速率限制）
# 快速发送多个请求...
```

---

### 第8步：性能测试

#### 8.1 并发请求测试

```bash
# 使用ab（Apache Bench）
ab -n 1000 -c 10 http://localhost:8080/api/v1/ping

# 使用wrk
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/ping

# 预期：
# - 平均响应时间 < 100ms
# - p95 响应时间 < 500ms
# - 错误率 < 1%
```

#### 8.2 数据库连接池测试

```bash
# 同时发起多个需要数据库查询的请求
for i in {1..50}; do
  curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files &
done
wait

# 验证所有请求都成功完成
```

---

## 🔧 故障排查

### 查看日志

```bash
# 后端API日志
tail -f /home/ec2-user/openwan/api-server.log

# 前端dev日志
tail -f /home/ec2-user/openwan/frontend-dev.log

# MySQL日志
sudo docker logs -f openwan-mysql-1

# Redis日志
sudo docker logs -f openwan-redis-1

# RabbitMQ日志
sudo docker logs -f openwan-rabbitmq-1
```

### 重启服务

```bash
# 重启后端API
pkill -f "bin/openwan"
cd /home/ec2-user/openwan && nohup ./bin/openwan > api-server.log 2>&1 &

# 重启前端
pkill -f "vite"
cd /home/ec2-user/openwan/frontend && nohup npm run dev > ../frontend-dev.log 2>&1 &

# 重启Docker容器
cd /home/ec2-user/openwan
sudo docker-compose restart
```

### 数据库检查

```bash
# 连接数据库
mysql -h 127.0.0.1 -P 3306 -u root -prootpassword openwan_db

# 查看表
SHOW TABLES;

# 查看用户
SELECT id, username, email, enabled FROM ow_users;

# 查看文件
SELECT id, title, type, status, upload_username FROM ow_files LIMIT 10;

# 查看分类
SELECT id, name, parent_id FROM ow_category;
```

---

## ✅ 测试检查清单

### 基础功能
- [ ] 所有Docker容器运行正常
- [ ] 后端API服务启动成功
- [ ] 前端dev服务启动成功
- [ ] 数据库连接正常
- [ ] Redis连接正常
- [ ] RabbitMQ连接正常

### 认证和授权
- [ ] 用户可以成功登录
- [ ] 登出功能正常
- [ ] 未认证访问被正确拦截（401）
- [ ] 权限检查正常（403）
- [ ] Session持久化正常

### 文件管理
- [ ] 文件列表加载正常
- [ ] 文件搜索和筛选正常
- [ ] 文件上传功能正常
- [ ] 文件详情查看正常
- [ ] 文件编辑功能正常
- [ ] 文件删除功能正常
- [ ] 文件下载功能正常

### 管理功能
- [ ] 用户管理CRUD正常
- [ ] 组管理CRUD正常
- [ ] 角色管理CRUD正常
- [ ] 分类管理CRUD正常
- [ ] 目录配置CRUD正常

### 前端UI
- [ ] 登录页面显示正常
- [ ] 仪表板页面显示正常
- [ ] 所有导航链接正常
- [ ] 表单验证正常
- [ ] 错误提示正常
- [ ] 加载状态显示正常
- [ ] 响应式布局正常

### API性能
- [ ] 响应时间符合要求
- [ ] 并发处理正常
- [ ] 错误处理正确
- [ ] 数据验证正常

---

## 📞 支持

如遇问题，请检查：
1. 日志文件：`api-server.log`, `frontend-dev.log`
2. 服务状态：`ps aux | grep "openwan\|vite"`
3. 容器状态：`sudo docker ps`
4. 端口占用：`netstat -tlnp | grep ":8080\|:3000"`

---

**测试环境准备完成！现在可以开始端到端测试。**

祝测试顺利！🚀
