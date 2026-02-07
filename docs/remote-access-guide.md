# OpenWan 远程访问测试指南

**服务器IP**: 13.217.210.142  
**更新时间**: 2026-02-02  
**测试状态**: ✅ Nginx代理已配置并运行

---

## 🌐 访问地址

### 1. 前端应用（Vue.js开发服务器）
```
http://13.217.210.142/
```
- 通过nginx代理到Vite开发服务器（localhost:3000）
- 支持HMR（热模块替换）
- 自动刷新

### 2. 登录测试页面
```
http://13.217.210.142/test_login.html
```
- 简单的登录表单测试
- 预填测试账号：admin/admin
- 显示API响应结果

### 3. API接口（通过nginx代理）
```
http://13.217.210.142/api/v1/...
```
- 后端Go服务运行在localhost:8080
- nginx代理到/api路径

### 4. 健康检查端点
```
http://13.217.210.142/health   # 健康检查
http://13.217.210.142/ready    # 就绪检查
http://13.217.210.142/alive    # 存活检查
```

---

## 🔧 Nginx配置

### 配置文件位置
```
/etc/nginx/conf.d/openwan.conf
```

### 主要配置内容

#### 1. API代理
```nginx
location /api/ {
    proxy_pass http://localhost:8080/api/;
    proxy_http_version 1.1;
    
    # 代理头设置
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # 超时设置（支持大文件上传）
    proxy_connect_timeout 300s;
    proxy_send_timeout 600s;
    proxy_read_timeout 600s;
    
    # 禁用缓冲（支持大文件上传）
    proxy_request_buffering off;
}
```

#### 2. 前端代理（开发模式）
```nginx
location / {
    proxy_pass http://localhost:3000;
    proxy_http_version 1.1;
    
    # WebSocket支持（Vite HMR需要）
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    
    # 代理头设置
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

#### 3. 文件上传限制
```nginx
client_max_body_size 500M;
client_body_buffer_size 128k;
client_body_timeout 600s;
```

---

## 🧪 测试方法

### 1. 命令行测试（curl）

#### 测试前端访问
```bash
curl -I http://13.217.210.142/
```
预期结果：HTTP 200 OK

#### 测试登录API
```bash
curl -X POST http://13.217.210.142/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```
预期结果：
```json
{
  "success": true,
  "message": "Login successful",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "thinkgem@gmail.com",
    "group_id": 1,
    "level_id": 5,
    "permissions": []
  }
}
```

#### 测试健康检查
```bash
curl http://13.217.210.142/health
```

### 2. 浏览器测试

#### 访问测试页面
1. 打开浏览器访问：http://13.217.210.142/test_login.html
2. 查看预填的测试账号（admin/admin）
3. 点击"登录"按钮
4. 查看响应结果

#### 访问前端应用
1. 打开浏览器访问：http://13.217.210.142/
2. 应该看到Vue.js应用首页
3. 尝试登录功能

### 3. 开发者工具测试

打开浏览器开发者工具（F12）：

#### Network标签
- 查看请求URL（应该是相对路径 /api/...）
- 查看请求头（X-Real-IP, X-Forwarded-For）
- 查看响应头（CORS头应该存在）
- 查看响应状态码

#### Console标签
- 检查是否有CORS错误
- 检查是否有其他JavaScript错误

---

## 🚀 服务状态

### 检查服务运行状态

```bash
# Nginx状态
sudo systemctl status nginx

# 后端服务
ps aux | grep "bin/openwan" | grep -v grep

# 前端开发服务器
ps aux | grep vite | grep -v grep

# 端口监听
sudo netstat -tlnp | grep -E ':(80|3000|8080)\s'
```

### 预期输出
```
tcp   0.0.0.0:80      nginx (master)
tcp   :::8080         ./bin/openwan (Go backend)
tcp   :::3000         node vite (Frontend)
```

---

## 🔍 故障排查

### 如果无法访问前端

#### 1. 检查Vite开发服务器
```bash
ps aux | grep vite
cd /home/ec2-user/openwan/frontend
npm run dev
```

#### 2. 检查nginx配置
```bash
sudo nginx -t
sudo nginx -s reload
```

#### 3. 查看nginx错误日志
```bash
sudo tail -f /var/log/nginx/openwan_error.log
```

### 如果API返回404

#### 1. 检查后端服务
```bash
ps aux | grep openwan
cd /home/ec2-user/openwan
./bin/openwan
```

#### 2. 直接测试后端（绕过nginx）
```bash
curl http://localhost:8080/api/v1/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

#### 3. 查看后端日志
```bash
tail -f /home/ec2-user/openwan/api-server.log
```

### 如果有CORS错误

#### 1. 检查nginx配置中的CORS头
```bash
grep -A 5 "Access-Control" /etc/nginx/conf.d/openwan.conf
```

#### 2. 检查后端CORS配置
- 文件：`/home/ec2-user/openwan/internal/api/router.go`
- 查找CORS中间件配置

### 如果文件上传失败

#### 1. 检查nginx文件大小限制
```bash
grep "client_max_body_size" /etc/nginx/conf.d/openwan.conf
```

#### 2. 增加超时时间
```bash
# 编辑nginx配置
sudo vim /etc/nginx/conf.d/openwan.conf
# 增加：
# proxy_read_timeout 900s;
# client_body_timeout 900s;
```

---

## 📊 性能监控

### 实时访问日志
```bash
sudo tail -f /var/log/nginx/openwan_access.log
```

### 实时错误日志
```bash
sudo tail -f /var/log/nginx/openwan_error.log
```

### 后端日志
```bash
tail -f /home/ec2-user/openwan/api-server.log
```

---

## 🔒 安全注意事项

### 当前配置（开发环境）
- ✅ 所有流量通过nginx代理
- ✅ 后端仅监听localhost:8080（不对外暴露）
- ✅ 前端开发服务器仅监听localhost:3000
- ✅ CORS头正确配置
- ⚠️ HTTP（未加密）- 生产环境需要HTTPS

### 生产部署建议
1. **启用HTTPS**
   ```bash
   # 安装Let's Encrypt证书
   sudo certbot --nginx -d yourdomain.com
   ```

2. **限制访问IP（可选）**
   ```nginx
   allow 192.168.1.0/24;
   deny all;
   ```

3. **添加基本认证（可选）**
   ```nginx
   auth_basic "Restricted";
   auth_basic_user_file /etc/nginx/.htpasswd;
   ```

4. **速率限制**
   ```nginx
   limit_req_zone $binary_remote_addr zone=login:10m rate=10r/m;
   location /api/v1/auth/login {
       limit_req zone=login burst=5;
   }
   ```

---

## 📝 测试清单

- [ ] 浏览器访问 http://13.217.210.142/ 显示前端
- [ ] 浏览器访问 http://13.217.210.142/test_login.html 显示测试页面
- [ ] 测试页面点击登录成功
- [ ] 前端应用登录功能正常
- [ ] API响应包含正确的用户数据
- [ ] CORS头存在且正确
- [ ] WebSocket连接正常（Vite HMR）
- [ ] 文件上传功能测试（如果已实现）
- [ ] 健康检查端点返回200

---

## 🆘 联系支持

如果遇到问题：
1. 查看本指南的"故障排查"部分
2. 检查服务日志
3. 验证所有服务都在运行
4. 提供详细的错误信息和日志

**当前架构**：
```
Internet → nginx:80 → {
    / → Vite (localhost:3000) → Vue.js Frontend
    /api → Go Backend (localhost:8080)
    /test_login.html → Static HTML
}
```
