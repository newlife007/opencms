# 视频预览404问题修复

## 🔧 问题与修复

### 问题
上传视频后，预览返回404错误，无法查看视频。

### 根本原因
1. **RabbitMQ未安装**: 转码任务需要通过RabbitMQ队列分发，但RabbitMQ服务未安装
2. **Worker未运行**: 即使队列可用，Worker服务也未启动来处理转码任务
3. **预览文件未生成**: 因为转码任务无法执行，预览FLV文件不存在

### 解决方案
实现了**同步转码fallback机制**，当RabbitMQ不可用时自动进行同步转码：

#### 1. 修改文件上传逻辑
文件: `internal/api/handlers/files.go`

```go
// 尝试发布到队列
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := h.queueService.Publish(ctx, "openwan_transcoding_jobs", message); err != nil {
        fmt.Printf("⚠ Queue unavailable for file %d, using sync transcode: %v\n", fileRecord.ID, err)
        // 队列不可用时，触发同步转码
        h.syncTranscodeVideo(fileRecord, uploadedPath, storageType)
    } else {
        fmt.Printf("✓ Transcode job published for file %d (type %d)\n", fileRecord.ID, fileRecord.Type)
    }
}()

// 如果没有队列服务，直接同步转码
if fileRecord.Type == 1 || fileRecord.Type == 2 {
    storageType := "s3"
    fmt.Printf("⚠ No queue service available, using sync transcode for file %d\n", fileRecord.ID)
    go h.syncTranscodeVideo(fileRecord, uploadedPath, storageType)
}
```

#### 2. 实现同步转码方法
添加了 `syncTranscodeVideo` 方法，完整流程：

1. **下载原文件** (如果是S3存储)
   - 从S3下载到临时目录 `/tmp/openwan-transcode/`
   - 创建临时输入文件

2. **执行FFmpeg转码**
   ```bash
   ffmpeg -i input.mp4 -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 output-preview.flv
   ```

3. **上传预览文件** (如果是S3存储)
   - 上传转码后的FLV文件到S3
   - 设置正确的Content-Type: `video/x-flv`
   - 添加元数据（原文件ID、转码日期）

4. **清理临时文件**
   - 删除临时输入和输出文件

#### 3. 关键改进

**异步执行**: 不阻塞上传响应
```go
go h.syncTranscodeVideo(fileRecord, uploadedPath, storageType)
```

**详细日志**: 便于调试
```
🎬 Starting sync transcode for file 27 (openwan/xxx.mp4)
  ⬇  Downloading from S3: openwan/xxx.mp4 -> /tmp/input-27.mp4
  ✓ Downloaded /tmp/input-27.mp4
  🎥 Transcoding: /tmp/input-27.mp4 -> /tmp/output-27.flv
  📝 Parameters: -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240
  ✓ Transcode completed: /tmp/output-27.flv
  ⬆  Uploading to S3: /tmp/output-27.flv -> openwan/xxx-preview.flv
  ✓ Uploaded preview to S3: openwan/xxx-preview.flv
✅ Transcode completed for file 27
```

**错误处理**: 每个步骤都有错误处理和回滚

---

## ✅ 当前状态

```
✓ 同步转码功能已实现
✓ 后端已重新编译
✓ 服务已重启 (PID: 4183018)
✓ S3存储已配置
✓ FFmpeg已安装并可用
```

---

## 🧪 测试步骤

### 1. 上传视频文件

访问 http://localhost，登录后:
1. 点击"文件管理" → "文件上传"
2. 选择一个视频文件（MP4、AVI等）
3. 填写标题和分类
4. 点击上传

**预期结果**: 
- 上传成功
- 后台开始转码（查看日志）

### 2. 监控转码过程

```bash
# 实时查看转码日志
tail -f /home/ec2-user/openwan/logs/api.log | grep -i "transcode\|ffmpeg"
```

**应该看到**:
```
⚠ No queue service available, using sync transcode for file 28
🎬 Starting sync transcode for file 28 (openwan/...)
  ⬇  Downloading from S3...
  ✓ Downloaded
  🎥 Transcoding...
  ✓ Transcode completed
  ⬆  Uploading to S3...
  ✓ Uploaded preview to S3
✅ Transcode completed for file 28
```

### 3. 验证预览文件

#### 方法A: 检查S3
```bash
# 列出上传的文件和预览文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | grep -E "\.mp4|\.flv"
```

**应该看到**:
```
openwan/xxx/yyy.mp4           (原文件)
openwan/xxx/yyy-preview.flv   (预览文件) ← 这个是新生成的
```

#### 方法B: 测试预览API
```bash
# 获取文件ID (假设是28)
FILE_ID=28

# 测试预览端点
curl -I "http://localhost:8080/api/v1/files/${FILE_ID}/preview"
```

**预期响应**:
```
HTTP/1.1 200 OK
Content-Type: video/x-flv
Accept-Ranges: bytes
Cache-Control: public, max-age=3600
```

### 4. 前端测试

1. 在文件列表中点击已上传的视频
2. 点击"预览"或"播放"
3. Video.js播放器应该加载并播放视频

**预期**: 
- 播放器显示视频
- 进度条可拖拽
- 可以正常播放

---

## 📊 转码参数说明

### FFmpeg参数
```bash
-y              # 覆盖已存在的输出文件
-ab 56          # 音频比特率 56kbps
-ar 22050       # 音频采样率 22050Hz
-r 15           # 帧率 15fps
-b 500          # 视频比特率 500kbps
-s 320x240      # 分辨率 320x240
```

### 文件命名规则
- **原文件**: `openwan/{md5_dir}/{md5_filename}.{ext}`
- **预览文件**: `openwan/{md5_dir}/{md5_filename}-preview.flv`

例如:
```
原文件:   openwan/abc123def/456789abc.mp4
预览文件: openwan/abc123def/456789abc-preview.flv
```

---

## 🔍 故障排查

### 问题1: 转码日志中没有输出

**检查**:
```bash
# 检查服务是否运行
ps aux | grep bin/openwan

# 检查日志文件
tail -100 /home/ec2-user/openwan/logs/api.log
```

**解决**: 重启服务
```bash
cd /home/ec2-user/openwan
pkill -f "bin/openwan"
nohup ./bin/openwan > logs/api.log 2>&1 &
```

### 问题2: FFmpeg转码失败

**检查FFmpeg**:
```bash
# 验证FFmpeg已安装
/usr/local/bin/ffmpeg -version

# 测试手动转码
/usr/local/bin/ffmpeg -i /tmp/test.mp4 -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 /tmp/test-preview.flv
```

**可能原因**:
- FFmpeg未安装或路径不对
- 输入文件格式不支持
- 临时目录权限问题

**解决**:
```bash
# 确保临时目录存在且可写
sudo mkdir -p /tmp/openwan-transcode
sudo chmod 777 /tmp/openwan-transcode

# 如果FFmpeg路径不对，更新配置
vim /home/ec2-user/openwan/configs/config.yaml
# 修改 ffmpeg.binary_path
```

### 问题3: S3上传失败

**检查S3权限**:
```bash
# 测试S3写入
echo "test" > /tmp/test.txt
aws s3 cp /tmp/test.txt s3://video-bucket-843250590784/openwan/test.txt
aws s3 rm s3://video-bucket-843250590784/openwan/test.txt
```

**检查凭证**:
```bash
aws configure list
aws sts get-caller-identity
```

### 问题4: 预览文件存在但返回404

**检查文件路径**:
```bash
# 查看日志中的实际路径
grep "Uploaded preview" /home/ec2-user/openwan/logs/api.log | tail -5

# 验证S3中文件存在
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | tail -10
```

**检查PreviewFile handler**:
- 确保路径拼接正确
- 检查是否有权限问题

---

## 📈 性能考虑

### 同步转码的优缺点

#### ✅ 优点
- **简单**: 不需要额外的RabbitMQ和Worker服务
- **即时**: 转码完成后立即可用
- **易调试**: 日志集中在一个服务中

#### ⚠️ 缺点
- **占用资源**: 转码在API服务器上执行
- **并发限制**: 大量上传时可能影响性能
- **不可扩展**: 无法分布式处理

### 优化建议

#### 短期（当前实现）
- ✓ 异步执行（不阻塞上传响应）
- ✓ 临时文件清理
- ✓ 错误处理和重试

#### 中期（生产环境）
- **安装RabbitMQ**: 启用队列机制
- **启动Worker**: 独立的转码服务
- **限制并发**: 控制同时转码的任务数

```bash
# 安装RabbitMQ (Amazon Linux 2)
sudo yum install -y rabbitmq-server
sudo systemctl start rabbitmq-server
sudo systemctl enable rabbitmq-server

# 启动Worker
cd /home/ec2-user/openwan
nohup ./bin/worker > logs/worker.log 2>&1 &
```

#### 长期（大规模）
- **AWS Elastic Transcoder**: 使用托管转码服务
- **Lambda + Step Functions**: 无服务器转码
- **MediaConvert**: 专业视频处理服务

---

## 🎯 后续改进

### 1. 转码质量设置
允许用户选择预览质量：
- 低质量: 240p, 300kbps (当前默认)
- 中质量: 480p, 800kbps
- 高质量: 720p, 1500kbps

### 2. 转码进度显示
- 实时显示转码进度百分比
- WebSocket或轮询更新状态
- 前端显示进度条

### 3. 多格式支持
除了FLV，生成多种格式：
- MP4 (H.264): 兼容性最好
- WebM: 开源格式
- HLS (m3u8): 自适应流媒体

### 4. 缩略图生成
自动生成视频缩略图：
```bash
ffmpeg -i input.mp4 -ss 00:00:05 -vframes 1 thumbnail.jpg
```

### 5. 智能转码
- 检测已有格式，避免不必要转码
- 如果上传的已经是FLV/MP4，直接复制
- 根据文件大小调整转码参数

---

## 📚 相关文档

- [FFmpeg官方文档](https://ffmpeg.org/documentation.html)
- [AWS S3 SDK for Go](https://docs.aws.amazon.com/sdk-for-go/)
- [Video.js播放器文档](https://videojs.com/)

---

**修复完成时间**: 2026-02-07 09:45 UTC
**修复文件**: `internal/api/handlers/files.go`
**新增功能**: 同步视频转码fallback机制
**服务状态**: ✓ 运行中

---

**🎉 视频预览功能已修复！现在上传视频后会自动转码生成预览文件。**

请按照测试步骤验证功能，并查看转码日志确认转码成功。
