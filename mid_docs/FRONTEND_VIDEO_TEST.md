# 前端视频预览测试指南

## 服务状态

### 后端服务 (Go API)
- **地址**: http://localhost:8080
- **状态**: 运行中
- **Screen会话**: `screen -r openwan`

### 前端服务 (Vue.js/Vite)
- **地址**: http://localhost:3000
- **状态**: 运行中
- **Screen会话**: `screen -r vite`

## 已实施的修复

### 1. 添加调试日志
在 `FileDetail.vue` 中的 `previewUrl` 计算属性添加了控制台日志：
```javascript
console.log('Computing previewUrl:', {
  fileId: fileId.value,
  fileType: fileInfo.value.type,
  isVideo: fileInfo.value.type === 1
})
```

### 2. 扩展视频/音频支持
- 原始代码只支持视频 (type === 1)
- 修改后支持视频和音频 (type === 1 || type === 2)

### 3. 添加页面调试信息
在预览区域显示调试信息：
- 文件类型（视频/音频/图片/富媒体）
- 预览URL路径

## 测试步骤

### 方法1: 使用浏览器开发者工具

1. **在浏览器中打开前端**
   - 访问: http://localhost:3000
   - 或者如果您的EC2实例有公网IP: http://<EC2-PUBLIC-IP>:3000

2. **登录系统**
   - 用户名: `testuser`
   - 密码: `testpass123`

3. **访问文件详情页**
   - 导航到文件列表页面
   - 点击任意视频文件（例如文件ID 71）
   - 或直接访问: http://localhost:3000/files/71

4. **检查浏览器控制台**
   按 F12 打开开发者工具，查看Console标签页，您应该看到：
   ```
   Computing previewUrl: { fileId: "71", fileType: 1, isVideo: true }
   Preview URL generated: /api/v1/files/71/preview
   Initializing video player with src: /api/v1/files/71/preview
   ```

5. **检查网络请求**
   在开发者工具的Network标签页中，筛选XHR/Fetch请求：
   - 应该看到 `GET /api/v1/files/71` (获取文件详情)
   - 应该看到 `GET /api/v1/files/71/preview` (获取预览文件)

6. **查看页面调试信息**
   在预览区域应该显示黄色背景的调试信息框：
   ```
   文件类型: 视频, 预览URL: /api/v1/files/71/preview
   ```

### 方法2: 使用命令行测试API

如果您无法访问浏览器，可以使用以下命令测试API：

```bash
# 1. 登录获取会话Cookie
curl -c /tmp/cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'

# 2. 获取文件详情
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files/71 | jq

# 3. 获取预览文件（应返回二进制视频数据）
curl -b /tmp/cookies.txt http://localhost:8080/api/v1/files/71/preview | head -c 100 | od -c
# 应该看到: \0  \0  \0       f   t   y   p   i   s   o   m （MP4文件头）
```

## 常见问题排查

### 问题1: 前端没有请求预览文件

**可能原因**:
1. `fileInfo.type` 不等于 1 或 2（不是视频或音频）
2. `previewUrl` 计算结果为 null
3. VideoPlayer组件未能正确加载

**排查步骤**:
1. 检查浏览器控制台的日志输出
2. 查看文件详情页的调试信息框
3. 确认文件类型字段正确

### 问题2: 请求了预览文件但返回404

**可能原因**:
1. 数据库中文件路径使用反斜杠 `\` 而不是正斜杠 `/`
2. 文件在存储路径中不存在

**解决方法**:
```bash
# 检查数据库中的文件路径
sudo docker exec openwan-mysql-1 mysql -uroot -prootpassword openwan_db \
  -e "SELECT id, type, name, path FROM ow_files WHERE id=71;"

# 如果path包含反斜杠，需要运行路径规范化脚本
```

### 问题3: VideoPlayer组件未渲染

**可能原因**:
1. 异步组件加载失败
2. video.js或flv.js依赖未正确安装

**解决方法**:
```bash
cd /home/ec2-user/openwan/frontend
npm list video.js videojs-flvjs-es6 flv.js
# 如果缺少依赖，运行:
npm install video.js videojs-flvjs-es6 flv.js
npm run build
```

## 验证成功的标志

✅ **前端成功请求预览文件的标志**:
1. 浏览器控制台显示 "Computing previewUrl" 日志
2. 浏览器控制台显示 "Preview URL generated" 日志
3. Network标签页显示 `GET /api/v1/files/71/preview` 请求（状态200）
4. 页面显示VideoPlayer组件（黑色播放器界面）
5. 页面显示调试信息框（黄色背景）

✅ **后端成功返回预览文件的标志**:
1. 预览请求返回200状态码
2. 响应体包含视频二进制数据（以 ftyp 开头）
3. Content-Type头包含 video/mp4 或 video/x-flv

## 下一步改进

如果基本功能验证通过，可以移除调试信息：

1. 删除 `console.log` 语句
2. 删除页面上的 `.debug-info` 元素
3. 重新构建前端: `npm run build`

## 联系和反馈

如果测试中遇到问题，请提供：
1. 浏览器控制台的完整日志（包括错误信息）
2. Network标签页中的请求详情（请求URL、状态码、响应）
3. 页面截图（显示调试信息框）
