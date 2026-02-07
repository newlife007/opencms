# 登录功能测试指南

## 🎯 问题已修复

**问题**: 登录成功但无法进入系统页面  
**原因**: 后端未返回token，路由守卫拦截  
**修复**: 后端添加token返回，前端自动适配

---

## ✅ 快速测试步骤

### 1. 访问系统

在浏览器中打开:
```
http://13.217.210.142/
```

应该自动跳转到登录页面 `/login`

---

### 2. 打开开发者工具

**快捷键**: 
- Windows/Linux: `F12`
- Mac: `Cmd + Option + I`

**切换到以下标签**:
1. **Console** - 查看日志
2. **Network** - 查看网络请求
3. **Application** → **Local Storage** - 查看存储

---

### 3. 执行登录

**测试账号**:
- 用户名: `admin`
- 密码: `admin`

点击"登录"按钮

---

### 4. 验证结果

#### ✅ 成功的表现

1. **页面行为**:
   - 显示"登录成功"绿色提示消息
   - 页面自动跳转到 `/dashboard`
   - 显示Dashboard内容（首页）
   - 左侧显示导航菜单
   - 右上角显示用户信息

2. **Network标签**:
   ```
   POST /api/v1/auth/login
   Status: 200 OK
   
   Response:
   {
     "success": true,
     "message": "Login successful",
     "token": "session_1_admin",  ← 有token
     "user": {
       "id": 1,
       "username": "admin",
       ...
     }
   }
   ```

3. **Console标签**:
   ```
   Login successful: {
     user: { id: 1, username: "admin", ... },
     token: "session_1_admin"  ← 有token
   }
   ```

4. **Local Storage**:
   ```
   Key: token
   Value: session_1_admin  ← 已保存
   ```

#### ❌ 失败的表现

如果出现以下情况，说明有问题:

1. **停留在登录页**:
   - 显示"登录成功"，但页面没跳转
   - 地址栏还是 `/login`

2. **Network标签显示**:
   ```json
   {
     "success": true,
     "token": "",  ← token为空
     ...
   }
   ```
   或者没有token字段

3. **Console错误**:
   ```
   Error: No token received
   Redirected to /login
   ```

---

### 5. 测试其他功能

登录成功后，测试以下功能确认正常:

1. **导航菜单**:
   - 点击"首页" → 进入Dashboard
   - 点击"文件管理" → 进入文件列表
   - 点击"搜索" → 进入搜索页面

2. **页面刷新**:
   - 按 `F5` 或 `Ctrl+R` 刷新页面
   - 应该保持登录状态
   - 不应该跳转回登录页

3. **退出登录**:
   - 点击右上角用户菜单
   - 选择"退出"
   - 应该跳转回登录页
   - Local Storage中的token应该被清除

---

## 🐛 问题排查

### 如果登录失败

#### 1. 检查API响应

**Network标签** → 找到 `auth/login` 请求 → **Preview/Response**

**可能的错误**:

**401 Unauthorized**:
```json
{
  "success": false,
  "message": "Invalid username or password"
}
```
→ 检查用户名和密码是否正确

**500 Internal Server Error**:
```json
{
  "success": false,
  "message": "Database connection failed"
}
```
→ 后端服务有问题

**EOF / Invalid request body**:
```json
{
  "success": false,
  "message": "Invalid request body",
  "error": "EOF"
}
```
→ 请求数据格式错误（前端问题）

#### 2. 检查服务状态

在服务器上执行:

```bash
# 检查后端
ps aux | grep bin/openwan
curl http://localhost:8080/health

# 检查前端
ps aux | grep "npm.*dev"
curl http://localhost:3001/

# 检查nginx
sudo systemctl status nginx
```

#### 3. 查看日志

```bash
# 后端日志
tail -50 /home/ec2-user/openwan/api-server.log

# 前端日志
tail -50 /home/ec2-user/openwan/frontend/vite-server.log

# Nginx日志
sudo tail -50 /var/log/nginx/openwan_error.log
```

---

## 🔧 常见问题解决

### Q1: 显示"登录成功"但停留在登录页

**原因**: token为空或未保存

**解决**:
1. 打开Console，查找错误信息
2. 检查Network → login响应是否有token
3. 检查Local Storage是否有token
4. 如果后端没返回token，说明后端代码未更新，需要重新编译

```bash
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api
kill $(ps aux | grep "[b]in/openwan" | awk '{print $2}')
nohup ./bin/openwan > api-server.log 2>&1 &
```

### Q2: 刷新页面后退出登录

**原因**: token过期或验证失败

**解决**:
1. 检查后端是否实现了token验证
2. 当前使用简单token，不会过期
3. 如果退出，查看Console和Network的错误

### Q3: 点击导航菜单没有反应

**原因**: 路由配置问题或组件加载失败

**解决**:
1. 打开Console查看错误
2. 检查是否有404 (Not Found)
3. 查看Vite是否正常运行

```bash
ps aux | grep "npm.*dev"
tail -20 /home/ec2-user/openwan/frontend/vite-server.log
```

---

## 📊 预期测试结果

### ✅ 全部通过

- [x] 登录页面正常显示
- [x] 输入用户名密码后点击登录
- [x] 显示"登录成功"提示
- [x] 自动跳转到Dashboard
- [x] 左侧导航菜单显示
- [x] 右上角显示用户信息
- [x] Network显示login返回token
- [x] Local Storage保存了token
- [x] Console无错误信息
- [x] 刷新页面保持登录状态
- [x] 可以正常导航到其他页面

---

## 📞 反馈

如果测试中发现问题，请提供以下信息:

1. **浏览器Console的错误信息**（截图或文本）
2. **Network标签中login请求的完整响应**（截图或JSON）
3. **Local Storage中的内容**（截图）
4. **当前显示的页面**（URL和截图）
5. **操作步骤**（如何复现问题）

---

**测试账号**: admin / admin  
**访问地址**: http://13.217.210.142/  
**文档路径**: `/home/ec2-user/openwan/docs/login-redirect-fix.md`
