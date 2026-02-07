# 登录页空白问题诊断

## 问题现象
- 访问 http://13.217.210.142/ 或 http://13.217.210.142/login
- 页面显示空白

## 可能的原因

### 1. JavaScript加载失败
**检查步骤**:
1. 打开浏览器开发者工具（F12）
2. 切换到 **Network** 标签
3. 刷新页面
4. 查看是否有红色的失败请求

**预期**:
- index.html: 200 OK
- /assets/index-*.js: 200 OK
- /assets/element-plus-*.js: 200 OK
- /assets/vue-vendor-*.js: 200 OK
- /assets/index-*.css: 200 OK

### 2. JavaScript执行错误
**检查步骤**:
1. 打开浏览器开发者工具（F12）
2. 切换到 **Console** 标签
3. 查看是否有红色错误信息

**可能的错误**:
- `Failed to load module`
- `Uncaught Error`
- `Cannot read property`

### 3. API Base URL问题
**检查步骤**:
1. 在Console中输入: `localStorage.getItem('token')`
2. 查看是否有token
3. 如果没有token，是正常的（未登录）

### 4. CORS或网络问题
**检查步骤**:
1. Network标签中查看API请求
2. 查看是否有CORS错误

## 快速测试

### 测试1: HTML是否正确返回
```bash
curl http://localhost/ | grep "<div id=\"app\">"
```
**预期**: 应该看到 `<div id=\"app\"></div>`

### 测试2: JavaScript文件是否可访问
```bash
curl -I http://localhost/assets/index-0896c63b.js
```
**预期**: HTTP/1.1 200 OK

### 测试3: API是否可访问
```bash
curl http://localhost/api/v1/ping
```
**预期**: `{"message":"pong","success":true}`

## 浏览器Console检查命令

打开浏览器Console（F12 → Console），执行:

```javascript
// 1. 检查Vue是否加载
console.log(window.Vue)

// 2. 检查ElementPlus是否加载
console.log(window.ElementPlus)

// 3. 检查API base URL
console.log(import.meta.env.VITE_API_BASE_URL)

// 4. 测试API
fetch('/api/v1/ping').then(r => r.json()).then(console.log)
```

## 常见解决方案

### 方案1: 清除浏览器缓存
1. Ctrl+Shift+Delete 打开清除浏览器数据
2. 选择"缓存的图片和文件"
3. 清除数据
4. 刷新页面（Ctrl+F5 强制刷新）

### 方案2: 检查Nginx配置
```bash
# 验证Nginx配置
sudo nginx -t

# 重新加载配置
sudo systemctl reload nginx

# 检查错误日志
sudo tail -50 /var/log/nginx/openwan_error.log
```

### 方案3: 重新编译前端
```bash
cd /home/ec2-user/openwan/frontend
npm run build
```

### 方案4: 检查文件权限
```bash
# 确保Nginx可以读取文件
ls -la /home/ec2-user/openwan/frontend/dist/
sudo chmod -R 755 /home/ec2-user/openwan/frontend/dist/
```

## 提供更多信息

请在浏览器中执行以下操作，并告诉我结果：

### 1. 打开开发者工具
- Windows/Linux: F12 或 Ctrl+Shift+I
- Mac: Cmd+Option+I

### 2. 查看Network标签
- 刷新页面（F5）
- 截图或告诉我哪些文件加载失败（红色）

### 3. 查看Console标签
- 是否有红色错误信息
- 复制完整的错误信息

### 4. 查看Elements标签
- 查看 `<div id="app"></div>` 是否为空
- 还是里面有内容

## 远程访问测试

从您的本地电脑浏览器访问:
```
http://13.217.210.142/
```

**如果看到白屏**:
1. 右键 → 检查 → Console
2. 查看错误信息

**如果看到登录页**:
- 说明部署成功！

## 临时解决方案

如果浏览器访问有问题，可以先用curl测试：

```bash
# 1. 测试HTML
curl http://localhost/ > /tmp/page.html
cat /tmp/page.html

# 2. 测试JS
curl http://localhost/assets/index-0896c63b.js | head -10

# 3. 测试API
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

---

请执行上述检查，并告诉我：
1. **Network标签**中是否有失败的请求？
2. **Console标签**中是否有错误信息？
3. **Elements标签**中 `<div id="app">` 是否为空？
