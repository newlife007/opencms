# 浏览器缓存清除指南

## 问题
前端显示为旧版本，因为：
1. Nginx之前配置了1年的缓存 (`expires 1y`)
2. 浏览器已经缓存了旧的JS/CSS文件
3. 即使服务器更新了文件，浏览器仍使用缓存版本

## 解决方案

### 1. 服务器端已修复
✅ Nginx配置已更新：
- 静态资源缓存时间从1年改为1小时
- 添加了no-cache头禁用缓存
- 文件位置：`/etc/nginx/conf.d/openwan.conf`

### 2. 用户端清除缓存方法

#### 方法1: 硬刷新（推荐）
**Windows/Linux**:
- Chrome/Edge/Firefox: `Ctrl + Shift + R`
- 或者: `Ctrl + F5`

**Mac**:
- Chrome/Safari: `Cmd + Shift + R`
- Firefox: `Cmd + Shift + R`

#### 方法2: 清除浏览器缓存
**Chrome/Edge**:
1. 按 `Ctrl + Shift + Delete`（Mac: `Cmd + Shift + Delete`）
2. 选择"缓存图片和文件"
3. 时间范围选择"全部"
4. 点击"清除数据"

**Firefox**:
1. 按 `Ctrl + Shift + Delete`
2. 选择"缓存"
3. 时间范围选择"全部"
4. 点击"清除"

#### 方法3: 无痕/隐私模式
临时测试用：
- Chrome/Edge: `Ctrl + Shift + N`
- Firefox: `Ctrl + Shift + P`
- Safari: `Cmd + Shift + N`

#### 方法4: 开发者工具（开发人员用）
1. 按 `F12` 打开开发者工具
2. 右键点击刷新按钮
3. 选择"清空缓存并硬性重新加载"

### 3. 验证方法

#### 检查文件是否最新
1. 打开开发者工具（F12）
2. 进入"网络"（Network）标签
3. 刷新页面
4. 找到 `index-*.js` 文件
5. 检查文件大小和时间戳

#### 检查登录功能
1. 清除缓存后访问登录页
2. 输入 admin/admin
3. 应该自动跳转到dashboard
4. 查看浏览器localStorage是否有token：
   - F12 → Application/存储 → Local Storage
   - 应该有 "token" 键，值为长字符串（213字符）

### 4. 为什么会出现这个问题

**原因**: 
```nginx
# 旧配置（问题）
location ~* \.(js|css)$ {
    expires 1y;                          # ❌ 1年缓存
    add_header Cache-Control "public, immutable";  # ❌ 永久不变
}
```

**后果**:
- 浏览器会缓存JS文件1年
- 即使服务器更新文件，浏览器不会请求新版本
- 用户看到的是旧代码

**新配置（已修复）**:
```nginx
# 新配置（正确）
location ~* \.(js|css)$ {
    expires 1h;                          # ✅ 1小时缓存
    add_header Cache-Control "no-cache, no-store, must-revalidate";  # ✅ 每次验证
}
```

### 5. 长期解决方案

#### 使用文件指纹（Hash）
Vite/Webpack自动为每个构建版本的文件添加hash：
```
index-131b9afe.js  ← hash确保唯一性
Login-6712e870.js
user-b957fd85.js
```

当代码更改时，hash会变化：
```
index-131b9afe.js  → index-a8f3d2c1.js
```

浏览器会识别为新文件，自动下载。

#### HTML不缓存
`index.html` 应该永不缓存（或短期缓存），因为它引用的JS文件名会变化。

### 6. 当前系统状态

**后端**:
- ✅ 运行在端口8080
- ✅ JWT认证正常
- ✅ 返回213字符token
- ✅ 125个权限已加载

**前端**:
- ✅ 最新版本已构建（2026-02-07 07:34）
- ✅ 部署到 /home/ec2-user/openwan/frontend/dist
- ✅ Nginx配置已更新（禁用长期缓存）
- ⚠️ 用户需要清除浏览器缓存

**登录流程**:
1. 用户输入admin/admin
2. POST /api/v1/auth/login → 返回token
3. 前端存储token到localStorage
4. 路由跳转到/dashboard
5. 后续请求带Authorization: Bearer <token>

### 7. 故障排查

#### 如果清除缓存后仍显示旧版本

1. **检查文件时间戳**:
```bash
ls -lh /home/ec2-user/openwan/frontend/dist/assets/ | grep js
# 应该显示最新时间（07:34或更新）
```

2. **检查Nginx是否加载新配置**:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

3. **检查浏览器控制台**:
- F12 → Console
- 查看是否有JavaScript错误
- 检查网络请求是否返回200

4. **检查localStorage**:
- F12 → Application → Local Storage
- 查找 "token" 键
- 应该有值（登录后）

5. **测试API直接访问**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq
```

应该返回包含token的JSON。

### 8. 联系支持

如果以上方法都无效：
1. 提供浏览器控制台截图（F12 → Console）
2. 提供网络请求截图（F12 → Network）
3. 说明浏览器类型和版本
4. 说明清除缓存的具体方法

---

**更新时间**: 2026-02-07 07:40  
**问题状态**: 服务器端已修复，等待用户清除浏览器缓存
