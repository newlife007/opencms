# 预览文件未生成问题分析报告

**问题**: 为什么预览文件没有生成？  
**时间**: 2026-02-07 06:20  
**状态**: 已诊断，需要启动worker

---

## 🔍 根本原因

**预览文件未生成是因为转码worker没有运行！**

### 转码流程概述

```
1. 用户上传视频/音频文件
   ↓
2. 后端创建文件记录
   ↓
3. 后端发布转码任务到SQS队列
   ↓
4. Worker从队列消费任务
   ↓
5. Worker调用FFmpeg生成preview.flv
   ↓
6. 用户可以预览视频
```

**当前状态**: 流程在步骤4中断，worker没有运行！

---

## 📊 诊断结果

### ✅ 已就绪的组件

1. **FFmpeg已安装**
   ```bash
   /usr/local/bin/ffmpeg
   ffmpeg version N-122154-g3332b2db84
   配置: --enable-libx264 --enable-gpl
   ```

2. **SQS队列已创建**
   ```
   队列URL: https://queue.amazonaws.com/843250590784/openwan-test-transcoding
   当前消息数: 0
   可见性超时: 3600秒
   ```

3. **Worker程序已编译**
   ```bash
   /home/ec2-user/openwan/bin/openwan-worker (10MB)
   编译时间: 2026-02-06 15:50
   ```

4. **后端API已实现转码触发**
   - 文件上传时自动发布转码任务
   - 队列名称: `openwan_transcoding_jobs`
   - 任务格式: JSON (FileID, InputPath, OutputPath, Parameters)

### ❌ 缺失的组件

**转码Worker进程没有运行！**

```bash
# 检查进程
ps aux | grep worker
# 结果: 无OpenWan worker进程
```

---

## 🛠️ 解决方案

### 方案1: 启动Worker进程（推荐）

#### 步骤1: 检查worker配置

```bash
cd /home/ec2-user/openwan

# 查看worker是否需要配置文件
./bin/openwan-worker --help
```

#### 步骤2: 创建worker配置

Worker需要访问：
- SQS队列（读取转码任务）
- S3存储（下载原文件，上传预览文件）
- 本地FFmpeg（执行转码）

配置文件示例（如果需要）：
```yaml
queue:
  type: sqs
  sqs_queue_url: https://queue.amazonaws.com/843250590784/openwan-test-transcoding
  sqs_region: us-east-1

storage:
  type: s3
  s3_bucket: openwan-media-843250590784
  s3_region: us-east-1

ffmpeg:
  path: /usr/local/bin/ffmpeg
  workers: 2  # 并发转码数量
  temp_dir: /tmp/transcode
```

#### 步骤3: 启动worker

```bash
cd /home/ec2-user/openwan

# 方式1: 前台运行（测试）
./bin/openwan-worker

# 方式2: 后台运行（生产）
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &

# 方式3: 多个worker实例（高并发）
for i in {1..2}; do
  nohup ./bin/openwan-worker > /tmp/worker-$i.log 2>&1 &
done
```

#### 步骤4: 验证worker运行

```bash
# 检查进程
ps aux | grep openwan-worker

# 查看日志
tail -f /tmp/worker.log

# 检查队列消费
aws sqs get-queue-attributes \
  --queue-url https://queue.amazonaws.com/843250590784/openwan-test-transcoding \
  --attribute-names ApproximateNumberOfMessages
```

---

### 方案2: 手动触发转码（测试）

如果不想启动worker，可以手动测试转码：

#### 假设有一个上传的视频文件

```bash
# 示例：手动转码
FILE_ID=22
INPUT="/path/to/original/video.mp4"
OUTPUT="/path/to/video-preview.flv"

ffmpeg -i "$INPUT" \
  -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 \
  "$OUTPUT"

# 检查输出文件
ls -lh "$OUTPUT"
```

但这只是临时方案，生产环境必须启动worker。

---

### 方案3: 同步转码（不推荐）

修改后端代码，在上传时同步执行转码（而不是发送到队列）：

**不推荐原因**:
- 大文件转码耗时长（几分钟）
- 会阻塞上传请求
- 无法横向扩展
- 用户体验差

---

## 🔍 深入分析

### 为什么队列中没有消息？

有两种可能：

#### 可能性1: 没有上传新文件

- 队列消息数: 0
- 说明没有新的转码任务
- 或者worker已经处理完所有任务（但worker没运行）

**验证方法**:
```bash
# 上传一个测试文件，观察队列
# 上传前
aws sqs get-queue-attributes --queue-url https://queue.amazonaws.com/843250590784/openwan-test-transcoding --attribute-names ApproximateNumberOfMessages

# 上传文件（通过API或前端）

# 上传后立即检查
aws sqs get-queue-attributes --queue-url https://queue.amazonaws.com/843250590784/openwan-test-transcoding --attribute-names ApproximateNumberOfMessages
```

#### 可能性2: 队列名称不匹配

代码中使用的队列名称：
```go
h.queueService.Publish(ctx, "openwan_transcoding_jobs", message)
```

但实际SQS队列名称：
```
openwan-test-transcoding
```

**问题**: 下划线 vs 连字符！

**验证**: 检查queueService的配置映射

---

## 🚀 立即行动步骤

### 1. 检查worker命令行参数

```bash
cd /home/ec2-user/openwan
./bin/openwan-worker --help
```

### 2. 启动worker（假设不需要特殊参数）

```bash
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
echo $! > /tmp/worker.pid
```

### 3. 查看worker日志

```bash
tail -f /tmp/worker.log
```

期望看到：
```
Worker starting...
Connecting to SQS queue...
Waiting for transcode jobs...
```

### 4. 上传测试视频

通过前端或API上传一个小视频文件（10MB以下）

### 5. 观察转码过程

```bash
# 终端1: 监控队列
watch -n 2 'aws sqs get-queue-attributes --queue-url https://queue.amazonaws.com/843250590784/openwan-test-transcoding --attribute-names ApproximateNumberOfMessages'

# 终端2: 监控worker日志
tail -f /tmp/worker.log

# 终端3: 监控FFmpeg进程
watch -n 1 'ps aux | grep ffmpeg'
```

### 6. 验证预览文件生成

```bash
# 假设文件ID=22，名称=testvideo
# 预览文件应该在：
# S3: s3://openwan-media-843250590784/openwan/files/22/testvideo-preview.flv
# 或本地: /path/to/data/files/22/testvideo-preview.flv

# 检查S3
aws s3 ls s3://openwan-media-843250590784/openwan/ --recursive | grep preview

# 测试预览API
curl -I http://localhost:8080/api/v1/files/22/preview
# 应该返回 200 OK
```

---

## 📋 Worker启动检查清单

在启动worker之前，确认：

- [ ] FFmpeg已安装且可执行
- [ ] AWS凭证已配置（SQS和S3访问）
- [ ] SQS队列存在并可访问
- [ ] S3存储桶存在并可访问
- [ ] 有足够的磁盘空间（/tmp至少10GB）
- [ ] Worker二进制文件有执行权限

检查命令：
```bash
# 1. FFmpeg
ffmpeg -version

# 2. AWS凭证
aws sts get-caller-identity

# 3. SQS访问
aws sqs get-queue-attributes --queue-url https://queue.amazonaws.com/843250590784/openwan-test-transcoding --attribute-names QueueArn

# 4. S3访问
aws s3 ls s3://openwan-media-843250590784/

# 5. 磁盘空间
df -h /tmp

# 6. 执行权限
ls -l /home/ec2-user/openwan/bin/openwan-worker
```

---

## 🎯 预期结果

启动worker后，转码流程应该是：

```
1. 用户上传video.mp4
   → 后端返回文件ID=25
   → 后端发布转码任务到SQS

2. Worker从SQS接收任务（1-20秒内）
   → Worker日志: "Received transcode job for file 25"
   
3. Worker从S3下载原文件（或读取本地）
   → Worker日志: "Downloading input file..."
   
4. Worker调用FFmpeg转码
   → Worker日志: "Starting FFmpeg transcoding..."
   → 可以看到ffmpeg进程（ps aux | grep ffmpeg）
   
5. 转码完成（10秒 - 5分钟，取决于文件大小）
   → Worker日志: "Transcoding completed"
   
6. Worker上传预览文件到S3
   → Worker日志: "Uploading preview file..."
   
7. Worker删除SQS消息
   → Worker日志: "Job completed for file 25"
   
8. 用户可以预览视频
   → GET /api/v1/files/25/preview 返回200
```

---

## 🐛 常见问题

### Q: Worker启动后立即退出？

**A**: 检查日志，可能是：
- AWS凭证错误
- SQS队列不存在或无权限
- 配置文件格式错误

### Q: Worker运行但不处理任务？

**A**: 检查：
- 队列名称是否匹配（下划线 vs 连字符）
- Worker是否在轮询正确的队列
- 队列中是否真的有消息

### Q: FFmpeg转码失败？

**A**: 常见原因：
- 输入文件格式不支持
- FFmpeg参数错误
- 磁盘空间不足
- 内存不足（大文件）

### Q: 预览文件生成但API仍返回404？

**A**: 检查：
- 文件路径是否正确
- S3对象键是否正确
- Storage服务是否配置正确

---

## 📖 相关文件

### 代码
- Worker入口: `cmd/worker/main.go`
- 转码服务: `internal/transcoding/service.go`
- FFmpeg封装: `internal/transcoding/ffmpeg.go`
- 队列服务: `internal/queue/sqs.go`
- 文件Handler: `internal/api/handlers/files.go`

### 配置
- 应用配置: `configs/config.yaml`
- Worker配置: `configs/worker.yaml`（如果存在）

### 日志
- Worker日志: `/tmp/worker.log`
- 后端日志: `/tmp/openwan.log`

---

## ✅ 总结

### 问题原因
预览文件未生成是因为 **转码Worker没有运行**。

### 解决步骤
1. ✅ 确认FFmpeg已安装
2. ✅ 确认SQS队列已创建
3. ✅ 确认Worker程序已编译
4. ❌ **需要启动Worker进程**
5. ⬜ 上传测试文件验证

### 下一步
**立即启动worker**:
```bash
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
tail -f /tmp/worker.log
```

然后上传一个测试视频，观察转码过程。

---

**报告时间**: 2026-02-07 06:25  
**问题原因**: Worker未运行  
**解决方案**: 启动worker进程  
**预计修复时间**: 5分钟（启动worker + 测试）
