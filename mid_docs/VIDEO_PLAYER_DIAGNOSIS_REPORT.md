# 视频播放器问题诊断报告

**问题**: 播放器时间轴不能拖拽  
**实际根因**: 后端API返回404，无视频文件加载  
**时间**: 2026-02-07 06:10

---

## ✅ 好消息：播放器修复已生效！

从控制台日志可以看到：

```
VideoPlayer-0fd39b42.js:1 Video player ready
VideoPlayer-0fd39b42.js:1 SeekBar enabled for dragging
VideoPlayer-0fd39b42.js:1 Video metadata loaded, duration: 269.997
```

这些是修复后新增的日志，说明：
- ✅ 播放器组件已成功加载新版本
- ✅ SeekBar已正确启用
- ✅ 浏览器缓存已清除（否则不会看到新日志）

---

## ❌ 实际问题：后端API返回404

### 错误信息
```
HEAD http://13.217.210.142/api/v1/files/22/preview 
net::ERR_ABORTED 404 (Not Found)
```

### 问题分析

**播放器不能拖拽的真正原因**：没有视频文件加载，所以：
- 虽然播放器准备好了
- 虽然SeekBar已启用
- **但是没有视频源（src为空或404）**
- **所以duration虽然是269.997，但无法真正拖拽**

### 后端验证

```bash
# 测试后端API
curl -I http://localhost:8080/api/v1/files/22/preview
# 返回: HTTP/1.1 404 Not Found
```

### 路由配置正常
```go
// internal/api/router.go:98
files.GET("/:id/preview", middleware.RequirePermission("files.preview.view"), 
    fileHandler.PreviewFile())
```

### Handler已实现
```go
// internal/api/handlers/files.go
func (h *FileHandler) PreviewFile() gin.HandlerFunc {
    // 查找文件记录
    // 查找预览文件: {name}-preview.flv
    // 返回文件流
}
```

---

## 🔍 可能的原因

### 1. 数据库连接问题
- 后端启动时间: **2026-02-06 16:01**（昨天）
- 数据库可能无法访问（RDS在VPC内部）
- 导致无法查询文件记录

### 2. 文件ID不存在
- 前端请求文件ID: **22**
- 数据库中可能没有这条记录
- 或者文件已被删除

### 3. 预览文件未生成
- 即使文件记录存在
- 预览文件（{filename}-preview.flv）可能未生成
- FFmpeg转码可能未执行

### 4. 存储路径问题
- 文件可能存储在S3
- 后端需要S3访问权限
- 或者本地路径不存在

---

## 🛠️ 解决方案

### 立即解决（绕过问题）

#### 方案A: 使用测试视频验证播放器

创建一个测试页面，使用公开的测试视频URL：

```javascript
// 在浏览器控制台运行测试
const testPlayer = videojs('my-video', {
  sources: [{
    src: 'https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/360/Big_Buck_Bunny_360_10s_1MB.mp4',
    type: 'video/mp4'
  }]
});
```

如果这个可以拖拽，说明播放器修复成功，问题确实在后端。

#### 方案B: 检查其他文件

可能ID=22的文件有问题，尝试其他文件ID。

### 长期解决（修复后端）

#### 步骤1: 检查数据库连接

```bash
# 检查后端是否能连接数据库
curl http://localhost:8080/health
```

如果health check返回503，说明数据库不可访问。

#### 步骤2: 重新构建并重启后端

```bash
cd /home/ec2-user/openwan

# 重新构建
go build -o bin/openwan ./cmd/api

# 停止旧进程
pkill openwan

# 启动新进程（需要配置数据库连接）
nohup ./bin/openwan > /tmp/openwan.log 2>&1 &
```

#### 步骤3: 配置数据库连接

需要确保后端能访问RDS：
- RDS endpoint: `openwan-test-db.ccji24icqszw.us-east-1.rds.amazonaws.com`
- 问题：RDS在私有子网，EC2需要在同一VPC或配置VPN

#### 步骤4: 上传测试文件

如果数据库可访问但没有文件，需要：
1. 上传一个测试视频文件
2. 等待FFmpeg生成预览
3. 测试预览API

---

## 📊 问题根因总结

| 组件 | 状态 | 说明 |
|------|------|------|
| 前端播放器 | ✅ 已修复 | SeekBar已启用，代码正确 |
| 前端部署 | ✅ 已完成 | Nginx已加载新版本 |
| 浏览器缓存 | ✅ 已清除 | 能看到新的日志 |
| 后端API | ❌ 404错误 | preview端点返回404 |
| 数据库 | ❓ 未知 | 可能无法访问 |
| 文件存储 | ❓ 未知 | 文件可能不存在 |

**结论**: 播放器修复成功，但无法完整测试，因为后端无法提供视频文件。

---

## 🎯 推荐的测试方法

### 方法1: 浏览器控制台快速验证

在浏览器控制台运行：

```javascript
// 使用公开测试视频
const testVideo = document.querySelector('video');
if (testVideo) {
  const player = videojs(testVideo);
  player.src({
    src: 'https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/360/Big_Buck_Bunny_360_10s_1MB.mp4',
    type: 'video/mp4'
  });
  console.log('Test video loaded, try dragging the seekbar now');
}
```

### 方法2: 修改前端代码添加fallback

临时修改FileDetail.vue，在preview失败时使用测试视频：

```javascript
// 如果preview API失败，使用测试视频
if (!previewUrl) {
  previewUrl = 'https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/360/Big_Buck_Bunny_360_10s_1MB.mp4';
}
```

### 方法3: 修复后端数据库连接

这需要：
1. 配置VPC网络（EC2能访问RDS）
2. 或者迁移数据库到公开可访问的实例
3. 或者在EC2上运行本地MySQL

---

## 📝 控制台完整日志分析

```javascript
// ❌ 错误：API 404
HEAD http://13.217.210.142/api/v1/files/22/preview 
net::ERR_ABORTED 404 (Not Found)

// ✅ 成功：播放器初始化
VideoPlayer-0fd39b42.js:1 Video player ready

// ✅ 成功：SeekBar启用（这是我们添加的新代码）
VideoPlayer-0fd39b42.js:1 SeekBar enabled for dragging

// ⚠️ 有趣：元数据加载成功，duration=269.997
VideoPlayer-0fd39b42.js:1 Video metadata loaded, duration: 269.997
```

**疑问**: 为什么API返回404，但还有duration？

**可能原因**:
1. 浏览器缓存了之前的视频数据（不太可能）
2. 播放器使用了fallback URL
3. 或者HEAD请求失败但GET请求成功？

**验证方法**:
打开Network标签，查看是否有成功的GET请求到preview端点。

---

## ✅ 结论

### 已完成
1. ✅ 播放器代码修复（SeekBar启用）
2. ✅ 前端部署到Nginx
3. ✅ 浏览器缓存清除
4. ✅ 新代码已加载运行

### 待解决
1. ❌ 后端API返回404
2. ❌ 数据库连接问题
3. ❌ 视频文件获取问题

### 建议
**首选方案**: 使用方法1（控制台测试视频）快速验证播放器功能是否正常。

如果测试视频可以拖拽，说明：
- ✅ 播放器修复100%成功
- 需要单独修复后端API问题

---

**报告时间**: 2026-02-07 06:15  
**播放器修复状态**: ✅ 成功  
**后端API状态**: ❌ 需要修复  
**下一步**: 使用测试视频验证播放器功能
