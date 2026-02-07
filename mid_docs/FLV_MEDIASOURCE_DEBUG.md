# FLV播放详细调试清单 - MediaSource关闭问题

**日期**: 2026-02-07 11:30 UTC  
**现象**: MediaSource onSourceClose - FLV.js初始化但MediaSource关闭  
**状态**: 🔍 **需要收集详细错误信息**

---

## 🚨 当前状态

### 已知信息

从您提供的错误日志中，我看到：

```javascript
✗ VIDEOJS: ERROR: (CODE:4 MEDIA_ERR_SRC_NOT_SUPPORTED)
✗ Video player error
✓ [MSEController] > MediaSource onSourceClose  ← 关键线索
```

**`MediaSource onSourceClose`** 说明：
- ✅ FLV.js已经初始化
- ✅ Media Source Extensions (MSE) 已创建
- ❌ MSE在加载数据时关闭了

---

## 📋 必需的调试信息清单

### 1. 完整的控制台日志

**请截图或复制从页面加载开始的所有日志**，特别关注：

#### 初始化日志（应该有）
```javascript
✓ Initializing player for type: video/x-flv
✓ Initializing FLV.js player
✓ Video.js UI ready
✓ ═══ Creating FLV Player ═══
✓ URL: /api/v1/files/32/preview
✓ CORS: true
✓ WithCredentials: true
✓ ═══════════════════════════
✓ FLV.js player attached and loaded
```

#### 媒体信息（如果成功）
```javascript
✓ FLV media info: { ... }
```

#### FLV错误（如果失败）
```javascript
✗ ═══ FLV ERROR ═══
✗ Error Type: ???
✗ Error Detail: ???
✗ Error Info: ???
✗ ═══════════════
```

**您看到了哪些日志？特别是FLV ERROR部分！**

---

### 2. Network标签 - 网络请求详情

打开 **开发者工具 (F12)** → **Network标签** → 刷新页面

#### HEAD请求
```
请求: HEAD /api/v1/files/32/preview
状态码: ??? (需要您提供)
响应头:
  Content-Type: ???
  Content-Length: ???
  Access-Control-Allow-Origin: ???
  Access-Control-Allow-Credentials: ???
```

#### GET请求
```
请求: GET /api/v1/files/32/preview
状态码: ??? (需要您提供)
大小: ???
时间: ???
响应头:
  Content-Type: ???
  Content-Length: ???
  Accept-Ranges: ???
```

**请点击这两个请求，查看详情，并告诉我状态码！**

---

### 3. Cookies检查

开发者工具 → **Application标签** → **Cookies** → 您的域名

**检查是否有**:
```
session_id: xxxxxxxxx  ← 应该存在
```

**如果没有session_id，说明未登录，这会导致401错误！**

---

## 🔍 可能的原因分析

### 原因1: 认证失败 (401 Unauthorized)

**现象**:
- GET请求返回401状态码
- FLV错误类型: `NetworkError`
- 错误详情: `NETWORK_STATUS_CODE_INVALID`

**检查**:
```javascript
// 控制台应该显示
✗ ═══ FLV ERROR ═══
✗ Error Type: NetworkError
✗ Error Detail: NETWORK_STATUS_CODE_INVALID
✗ → Network Error: Check authentication, CORS, or file availability
✗   → HTTP status code error (401/403/404/etc)
```

**解决方案**: 重新登录系统

---

### 原因2: 文件不存在 (404 Not Found)

**现象**:
- GET请求返回404状态码
- 预览文件未生成

**解决方案**: 触发转码任务
```bash
curl -X POST http://13.217.210.142/api/v1/files/32/transcode
```

---

### 原因3: CORS问题

**现象**:
- 控制台显示CORS错误
- Network标签中请求显示为红色
- 响应头缺少 `Access-Control-Allow-Origin`

**检查响应头应包含**:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, HEAD, OPTIONS
```

**解决方案**: 后端需要配置CORS（Gin middleware）

---

### 原因4: FLV格式损坏

**现象**:
- GET请求返回200 OK
- 数据下载成功
- FLV错误类型: `MediaError`

**检查**:
```javascript
✗ ═══ FLV ERROR ═══
✗ Error Type: MediaError
✗ → Media Error: FLV format issue or codec not supported
```

**解决方案**: 重新转码或验证FLV文件
```bash
# 在服务器上检查FLV文件
aws s3 cp s3://video-bucket-843250590784/openwan/.../file-preview.flv /tmp/
ffprobe /tmp/file-preview.flv

# 期望输出:
# Video: h264 (avc1)
# Audio: aac
```

---

### 原因5: Range请求问题

**现象**:
- GET请求带有Range头
- 服务器不支持Range请求
- 返回416 Range Not Satisfiable

**检查Network标签**:
```
请求头:
  Range: bytes=0-

响应码:
  206 Partial Content ✓ (正确)
  或
  416 Range Not Satisfiable ✗ (错误)
```

**解决方案**: 后端需要正确处理Range请求（已实现`http.ServeContent`）

---

## 🧪 快速测试方法

### 方法1: 使用curl测试API（最快）

```bash
# 获取您的session cookie
# 在浏览器中: document.cookie

# 测试HEAD请求
curl -v -H "Cookie: session_id=YOUR_SESSION_ID" \
  http://13.217.210.142/api/v1/files/32/preview \
  --head

# 期望: 200 OK, Content-Type: video/x-flv

# 测试GET请求（下载前1MB）
curl -H "Cookie: session_id=YOUR_SESSION_ID" \
  http://13.217.210.142/api/v1/files/32/preview \
  --output /tmp/test.flv \
  --max-filesize 1048576

# 检查文件
file /tmp/test.flv
# 期望: Flash Video
```

**请执行这些命令并提供输出！**

---

### 方法2: FLV测试页面

访问: **http://13.217.210.142/flv-test.html**

这个测试页面会显示详细的FLV.js错误信息。

**请访问并截图结果！**

---

### 方法3: 检查是否缓存未清除

**症状**: 
- 控制台日志少于预期
- 没有看到 "═══ Creating FLV Player ═══"

**解决**: 
1. 打开隐身窗口测试
2. 或完全清除缓存并硬刷新

---

## 📊 调试决策树

```
MediaSource onSourceClose错误
│
├─ 控制台有 "═══ FLV ERROR ═══" 吗?
│  │
│  ├─ 有 → 查看Error Type和Error Detail
│  │  │
│  │  ├─ NetworkError + NETWORK_STATUS_CODE_INVALID
│  │  │  → 检查Network标签状态码
│  │  │     ├─ 401 → 重新登录
│  │  │     ├─ 403 → 检查权限
│  │  │     ├─ 404 → 文件不存在，触发转码
│  │  │     └─ 其他 → 提供状态码和响应头
│  │  │
│  │  └─ MediaError
│  │     → FLV格式问题，需要验证文件
│  │
│  └─ 没有 → FLV.js可能未初始化
│     ├─ 检查是否看到 "Initializing FLV.js player"
│     ├─ 检查缓存是否清除
│     └─ 检查videoType是否为'video/x-flv'
│
└─ Network标签有请求吗?
   │
   ├─ 有 → 提供状态码和响应头
   │
   └─ 没有 → 请求可能被浏览器阻止（CORS或其他）
```

---

## 🎯 最可能的原因（按概率）

| 概率 | 原因 | 症状 | 解决方案 |
|-----|------|------|---------|
| 60% | **认证失败** | GET返回401 | 重新登录 |
| 20% | **文件不存在** | GET返回404 | 触发转码 |
| 10% | **CORS问题** | CORS错误 | 配置后端 |
| 5% | **FLV损坏** | MediaError | 重新转码 |
| 5% | **缓存问题** | 日志不完整 | 清除缓存 |

---

## 📝 请提供以下信息

为了精确定位问题，**请提供**：

### 必需信息（按重要性排序）

1. ⭐⭐⭐ **Network标签中两个请求的状态码**
   ```
   HEAD /api/v1/files/32/preview → ??? (200/401/403/404?)
   GET /api/v1/files/32/preview → ??? (200/401/403/404?)
   ```

2. ⭐⭐⭐ **控制台是否显示 "═══ FLV ERROR ═══"**
   - 如果有，Error Type和Error Detail是什么？

3. ⭐⭐ **是否看到 "Initializing FLV.js player" 日志**
   - 如果没有，说明type不是'video/x-flv'

4. ⭐⭐ **Application → Cookies中是否有session_id**
   - 如果没有，需要重新登录

5. ⭐ **curl测试结果**（可选但很有帮助）

---

## 🔧 更新内容（刚刚完成）

### 增强的错误日志

现在控制台会显示更详细的错误信息：

```javascript
// 初始化信息
═══ Creating FLV Player ═══
URL: /api/v1/files/32/preview
CORS: true
WithCredentials: true
═══════════════════════════

// 错误信息（如果有）
═══ FLV ERROR ═══
Error Type: NetworkError
Error Detail: NETWORK_STATUS_CODE_INVALID
Error Info: { ... }
→ Network Error: Check authentication, CORS, or file availability
  → HTTP status code error (401/403/404/etc)
═══════════════

// 统计信息
FLV statistics: {
  speed: '500 KB/s',
  decodedFrames: 100,
  droppedFrames: 0
}
```

---

## ✅ 前置检查清单

在提供信息前，请确认：

- [ ] **浏览器缓存已清除** (Ctrl+Shift+Delete)
- [ ] **页面已硬刷新** (Ctrl+F5)
- [ ] **使用现代浏览器** (Chrome 90+, Firefox 88+)
- [ ] **开发者工具已打开** (F12)
- [ ] **Console标签可见** (查看完整日志)
- [ ] **Network标签已打开** (查看请求)
- [ ] **已登录系统** (有session cookie)

---

## 🚀 下一步行动

根据您提供的信息，我将能够：

1. **精确诊断问题原因**
2. **提供具体的修复代码**（如果是代码问题）
3. **指导配置调整**（如果是配置问题）
4. **验证修复效果**

---

**请提供上述信息，特别是：**
1. **Network标签的状态码**
2. **控制台是否有"═══ FLV ERROR ═══"及其内容**

有了这两个信息，我就能立即找到问题所在！🎯

---

**更新时间**: 2026-02-07 11:30 UTC  
**状态**: 等待调试信息  
**优先级**: 高 - 需要状态码和错误类型
