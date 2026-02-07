# FLV播放调试指南

**日期**: 2026-02-07 11:20 UTC  
**错误**: CODE:4 MEDIA_ERR_SRC_NOT_SUPPORTED  
**状态**: 🔍 **需要收集调试信息**

---

## 🔍 必需的调试信息

### 1. 浏览器控制台日志

请在浏览器中打开 **开发者工具 (F12)** → **Console标签**，并查找以下日志：

#### 应该看到的初始化日志：
```javascript
✓ Initializing player for type: video/x-flv
✓ Initializing FLV.js player
✓ Video.js UI ready
✓ FLV.js player attached and loaded
```

#### FLV媒体信息（如果成功加载）：
```javascript
✓ FLV media info: {
    duration: 123.45,
    hasVideo: true,
    hasAudio: true,
    ...
  }
```

#### 错误信息（如果加载失败）：
```javascript
✗ FLV player error: NetworkError - ...
或
✗ FLV player error: MediaError - ...
```

**请截图或复制完整的控制台日志！**

---

### 2. 网络请求状态

打开 **开发者工具 (F12)** → **Network标签**，刷新页面，查找以下请求：

#### HEAD请求：
```
HEAD /api/v1/files/32/preview
状态码: ?
响应头:
  Content-Type: ?
  Content-Length: ?
```

#### GET请求：
```
GET /api/v1/files/32/preview
状态码: ?
响应头:
  Content-Type: ?
  Content-Length: ?
  Accept-Ranges: ?
```

**重要**: 请提供这两个请求的完整状态码和响应头！

---

### 3. 可能的错误原因

根据 CODE:4 错误，可能的原因包括：

#### A. 认证失败 (401/403)
**现象**: 网络请求返回401或403  
**原因**: Session过期或用户无权限  
**解决**: 重新登录

#### B. 文件不存在 (404)
**现象**: GET请求返回404  
**原因**: 预览文件未生成  
**解决**: 触发转码任务

#### C. CORS问题
**现象**: 控制台显示CORS错误  
**原因**: 跨域配置问题  
**解决**: 检查后端CORS设置

#### D. FLV格式问题
**现象**: FLV player error: MediaError  
**原因**: FLV文件损坏或格式不标准  
**解决**: 重新转码或检查FFmpeg参数

#### E. MSE不支持
**现象**: FLV.js is not supported  
**原因**: 浏览器不支持Media Source Extensions  
**解决**: 使用Chrome/Firefox/Edge现代浏览器

---

## 🧪 快速测试方法

### 方法1: 直接访问FLV测试页面

访问: http://13.217.210.142/flv-test.html

这个测试页面会直接使用FLV.js加载视频，绕过Vue和Video.js，可以快速定位问题。

**预期结果**:
- ✓ FLV.js is supported
- ✓ FLV player loaded
- ✓ Media info: {...}
- ✓ 视频开始播放

**如果失败，会显示具体的错误类型和详情**

---

### 方法2: 使用curl测试API

```bash
# 测试HEAD请求
curl -I -H "Cookie: session_id=YOUR_SESSION_ID" \
  http://13.217.210.142/api/v1/files/32/preview

# 预期: 200 OK, Content-Type: video/x-flv

# 测试GET请求（获取前1KB数据）
curl -H "Cookie: session_id=YOUR_SESSION_ID" \
  -H "Range: bytes=0-1023" \
  http://13.217.210.142/api/v1/files/32/preview \
  --output /tmp/test.flv

# 检查文件类型
file /tmp/test.flv
# 预期: Flash Video
```

---

### 方法3: 检查FLV文件完整性

```bash
# 在服务器上检查S3文件
aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/33ab512143b66df625abaec6521383a3/ --recursive

# 下载FLV文件
aws s3 cp s3://video-bucket-843250590784/openwan/2026/02/07/33ab512143b66df625abaec6521383a3/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv /tmp/

# 使用ffprobe验证
ffprobe /tmp/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv

# 预期输出:
# Duration: XX:XX:XX
# Video: h264 (avc1), ...
# Audio: aac, ...
```

---

## 🔧 最新代码修复

### 修复内容（刚刚更新）

1. **禁用Video.js预加载**
   ```javascript
   preload: 'none' // 不让Video.js干扰FLV.js
   ```

2. **移除Video.js源配置**
   ```javascript
   // 不设置sources - 完全由FLV.js接管
   ```

3. **增强FLV错误处理**
   ```javascript
   flvPlayer.on(flvjs.Events.ERROR, (errorType, errorDetail, errorInfo) => {
     console.error('FLV player error:', errorType, errorDetail, errorInfo)
   })
   ```

4. **添加更多FLV事件监听**
   ```javascript
   flvPlayer.on(flvjs.Events.LOADING_COMPLETE, ...)
   flvPlayer.on(flvjs.Events.RECOVERED_EARLY_EOF, ...)
   ```

5. **改进自动播放逻辑**
   ```javascript
   if (props.autoplay) {
     videoEl.addEventListener('loadedmetadata', () => {
       videoEl.play()
     }, { once: true })
   }
   ```

---

## 📊 调试检查清单

在测试前，请确认：

- [ ] **浏览器缓存已清除** (Ctrl+Shift+Delete → 清除缓存)
- [ ] **页面已硬刷新** (Ctrl+F5)
- [ ] **使用现代浏览器** (Chrome 90+, Firefox 88+, Edge 90+)
- [ ] **用户已登录** (检查Application → Cookies → session_id)
- [ ] **后端服务运行** (服务器上: ps aux | grep openwan)
- [ ] **S3文件存在** (aws s3 ls s3://...)
- [ ] **加载新版本JS** (Network → videojs-core-5363c386.js)

---

## 🎯 预期的成功流程

### 完整的成功日志应该是：

```javascript
// 1. 初始化
Initializing player for type: video/x-flv
Initializing FLV.js player
Video.js UI ready

// 2. FLV加载
FLV.js player attached and loaded

// 3. 网络请求
[Network] HEAD /api/v1/files/32/preview → 200 OK
[Network] GET /api/v1/files/32/preview → 200 OK (video/x-flv, 8538824 bytes)

// 4. 媒体信息
FLV media info: {
  "audioCodec": "mp4a.40.2",
  "audioDataRate": 56,
  "audioSampleRate": 22050,
  "duration": 123.45,
  "framerate": 15,
  "hasAudio": true,
  "hasVideo": true,
  "height": 240,
  "videoCodec": "avc1.64001f",
  "videoDataRate": 500,
  "width": 320
}

// 5. 播放
Video metadata loaded, duration: 123.45
SeekBar enabled for interaction

// 6. 无错误
[无 CODE:4 错误] ✅
```

---

## 🚨 常见错误及解决方案

### 错误1: "FLV player error: NetworkError - 401"
**原因**: 认证失败  
**解决**: 
```javascript
// 检查Cookie
document.cookie
// 应包含 session_id=...

// 重新登录
window.location.href = '/login'
```

### 错误2: "FLV player error: NetworkError - 404"
**原因**: 预览文件不存在  
**解决**: 触发转码
```bash
curl -X POST http://13.217.210.142/api/v1/files/32/transcode
```

### 错误3: "FLV player error: MediaError - Format error"
**原因**: FLV格式问题  
**解决**: 检查FFmpeg转码参数
```bash
# 在服务器上重新转码
ffmpeg -i input.mp4 -y -ab 56k -ar 22050 -r 15 -b:v 500k -s 320x240 output.flv
```

### 错误4: "CORS policy error"
**原因**: 跨域配置  
**解决**: 检查Nginx/后端CORS设置
```nginx
# nginx.conf
add_header 'Access-Control-Allow-Origin' '*';
add_header 'Access-Control-Allow-Methods' 'GET, HEAD, OPTIONS';
add_header 'Access-Control-Allow-Headers' 'Range, Authorization, Cookie';
add_header 'Access-Control-Allow-Credentials' 'true';
```

### 错误5: "FLV.js is not supported"
**原因**: 浏览器不支持MSE  
**解决**: 
- 使用Chrome 90+ 或 Firefox 88+
- 不要使用IE浏览器
- 检查浏览器是否开启了隐私模式限制

---

## 📝 请提供的信息

为了帮助您解决问题，请提供：

1. **完整的浏览器控制台日志** (截图或文本)
2. **Network标签中的请求详情** (HEAD和GET请求的状态码、响应头)
3. **使用的浏览器和版本** (例如: Chrome 120, Firefox 115)
4. **用户登录状态** (Application → Cookies → session_id是否存在)
5. **FLV测试页面的结果** (访问 http://13.217.210.142/flv-test.html)

有了这些信息，我可以精确定位问题并提供针对性的解决方案！

---

**更新时间**: 2026-02-07 11:20 UTC  
**状态**: 等待调试信息收集  
**下一步**: 根据实际错误信息提供具体解决方案
