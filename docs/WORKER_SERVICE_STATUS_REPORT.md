# Worker服务启动完成报告

**时间**: 2026-02-07 06:35  
**状态**: ✅ Worker已启动并运行

---

## ✅ Worker服务状态

### 基本信息

```
进程ID: 4025524
日志文件: /tmp/worker.log
配置文件: /home/ec2-user/openwan/configs/config.yaml
Worker线程数: 4
```

### 运行状态

```bash
ps aux | grep openwan-worker
# ec2-user 4025526 ... ./bin/openwan-worker
```

✅ **Worker正常运行**

---

## 📋 Worker配置

### 从 config.yaml 加载的配置

```yaml
storage:
  type: s3
  s3_bucket: video-bucket-843250590784
  s3_region: us-east-1
  s3_prefix: openwan/

ffmpeg:
  binary_path: /usr/local/bin/ffmpeg
  parameters: "-y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240"
  worker_count: 4
  temp_dir: /tmp/openwan-transcode

queue:
  type: rabbitmq
  rabbitmq_url: amqp://guest:guest@localhost:5672/
  queues:
    transcoding: openwan_transcoding_jobs
```

---

## 🔄 Worker工作流程

### 正常流程

```
1. Worker从RabbitMQ队列订阅消息
   ↓
2. 接收转码任务消息
   {
     "file_id": 22,
     "input_path": "openwan/.../file.mp4",
     "output_path": "openwan/.../file-preview.flv",
     "storage_type": "s3"
   }
   ↓
3. 从S3下载原文件到本地临时目录
   /tmp/openwan-transcode/input-{file_id}.mp4
   ↓
4. 调用FFmpeg转码
   ffmpeg -i /tmp/.../input.mp4 {params} /tmp/.../output.flv
   ↓
5. 上传预览文件到S3
   s3://.../file-preview.flv
   ↓
6. 清理本地临时文件
   ↓
7. 确认消息处理完成（ACK）
   ↓
8. 等待下一个任务
```

---

## 🔍 启动日志分析

### Worker初始化 ✅

```
========================================
OpenWan Transcoding Worker
Version: 1.0.0
========================================

✓ Configuration loaded
  FFmpeg: /usr/local/bin/ffmpeg
  Queue: amqp://guest:guest@localhost:5672/
  Workers: 4

✓ FFmpeg service initialized
✓ Connected to message queue

Starting 4 workers...
✓ All workers started

========================================
Worker service is running
Waiting for transcoding jobs...
========================================
```

### Worker线程启动 ✅

```
[Worker 1] Started, subscribing to queue: openwan_transcoding_jobs
[Worker 2] Started, subscribing to queue: openwan_transcoding_jobs
[Worker 3] Started, subscribing to queue: openwan_transcoding_jobs
[Worker 4] Started, subscribing to queue: openwan_transcoding_jobs

2026/02/07 06:24:13 Started consuming from queue (x4)
```

### 处理历史任务 ⚠️

Worker启动后发现队列中有一个旧任务（文件ID=22）：

```
[Worker 2] Processing job for file 22
  Input: openwan/2026/02/06/1bf98d6756da9af30fc8b983b0e52dc0/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
  Output: openwan/.../6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv
  Storage: s3

✗ Transcoding failed: input file does not exist
```

**问题原因**: Worker在S3模式下查找文件时路径处理有误
- S3中确实有该文件
- 但Worker可能在本地文件系统查找，而不是从S3下载

**结果**: 任务重试3次后失败，发送到死信队列（DLQ）

---

## 🐛 发现的问题

### 问题1: S3文件路径处理

**症状**:
```
Transcoding failed: input file does not exist: openwan/2026/02/06/.../file.mp4
```

**验证S3文件存在**:
```bash
aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/06/1bf98d6756da9af30fc8b983b0e52dc0/
# 输出: 2026-02-06 16:10:48  8924094 6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
```

文件确实存在于S3！

**可能原因**:
1. Worker在检查文件存在时使用了本地路径逻辑
2. S3下载逻辑未正确触发
3. 路径拼接错误（缺少bucket名称或完整S3 URI）

**建议修复**:
- 检查 `internal/transcoding/service.go` 中的S3下载逻辑
- 确保在S3模式下正确构建S3对象键
- 添加详细的日志输出S3下载过程

---

## ✅ RabbitMQ队列状态

### 当前队列状态

```
队列名称: openwan_transcoding_jobs
消息总数: 0
待处理: 0
处理中: 0
消费者数: 4 ✓
```

**解读**:
- ✅ 4个Worker已连接并准备好消费消息
- ✅ 队列中没有待处理的消息
- ✅ 失败的任务已发送到DLQ
- ✅ Worker等待新任务

### 访问RabbitMQ管理界面

```
URL: http://localhost:15672
用户名: guest
密码: guest
```

可以查看：
- Queues页面：查看队列深度和消费速率
- DLQ（死信队列）：查看失败的消息
- Connections：查看Worker连接状态

---

## 🧪 测试Worker

### 方法1: 上传新文件（推荐）

通过前端上传一个新的视频文件：
1. 访问 http://13.217.210.142/
2. 登录
3. 上传视频文件
4. 观察Worker日志

```bash
# 监控Worker日志
tail -f /tmp/worker.log
```

**期望输出**:
```
[Worker X] Processing job for file {NEW_ID}
[Worker X]   Input: openwan/2026/02/07/.../file.mp4
[Worker X]   Output: openwan/2026/02/07/.../file-preview.flv
[Worker X]   Storage: s3
[Worker X] ✓ Downloading from S3...
[Worker X] ✓ Starting FFmpeg transcoding...
[Worker X] ✓ Transcoding completed in 15.3s
[Worker X] ✓ Uploading preview to S3...
[Worker X] ✓ Job completed for file {NEW_ID}
```

### 方法2: 手动发送测试消息

创建一个简单的测试脚本发送消息到RabbitMQ：

```bash
# 安装pika（Python RabbitMQ客户端）
pip3 install pika

# 创建测试脚本
cat > /tmp/test_queue.py << 'EOF'
import pika
import json

connection = pika.BlockingConnection(
    pika.ConnectionParameters('localhost')
)
channel = connection.channel()

message = {
    "file_id": 999,
    "input_path": "openwan/test/test.mp4",
    "output_path": "openwan/test/test-preview.flv",
    "storage_type": "s3"
}

channel.basic_publish(
    exchange='',
    routing_key='openwan_transcoding_jobs',
    body=json.dumps(message)
)

print("✓ 测试消息已发送")
connection.close()
EOF

python3 /tmp/test_queue.py
```

---

## 📊 监控命令

### 实时监控Worker

```bash
# 监控Worker日志
tail -f /tmp/worker.log

# 监控Worker进程
watch -n 1 'ps aux | grep openwan-worker'

# 监控FFmpeg进程
watch -n 1 'ps aux | grep ffmpeg | grep -v grep'

# 监控临时目录
watch -n 1 'ls -lh /tmp/openwan-transcode/'
```

### 监控队列

```bash
# 查看队列深度
curl -s http://localhost:15672/api/queues/%2F/openwan_transcoding_jobs \
  -u guest:guest | python3 -m json.tool | grep messages

# 实时监控队列
watch -n 1 'curl -s http://localhost:15672/api/queues/%2F/openwan_transcoding_jobs -u guest:guest | python3 -m json.tool | grep -E "name|messages"'
```

### 监控S3上传

```bash
# 查看S3最新文件
aws s3 ls s3://video-bucket-843250590784/openwan/ \
  --recursive --human-readable | tail -10

# 监控预览文件生成
watch -n 2 'aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | grep preview | tail -5'
```

---

## 🔧 Worker管理命令

### 启动Worker

```bash
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
echo $! > /tmp/worker.pid
```

### 停止Worker

```bash
# 优雅停止（Ctrl+C或SIGTERM）
kill $(cat /tmp/worker.pid)

# 强制停止
kill -9 $(cat /tmp/worker.pid)

# 或者
pkill openwan-worker
```

### 重启Worker

```bash
pkill openwan-worker && sleep 2
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
```

### 查看Worker状态

```bash
# 检查进程
ps aux | grep openwan-worker

# 查看日志尾部
tail -50 /tmp/worker.log

# 搜索错误
grep -i "error\|failed" /tmp/worker.log | tail -20

# 搜索成功的任务
grep "✓ Job completed" /tmp/worker.log
```

---

## ⚠️ 已知问题和建议

### 问题1: S3文件下载失败

**当前状态**: Worker在处理S3存储的文件时路径解析有问题

**临时解决方案**: 
- 上传新文件测试是否能正常工作
- 如果新文件也失败，需要修复S3下载逻辑

**永久解决方案**:
- 检查 `internal/transcoding/service.go`
- 确保S3模式下正确下载文件
- 添加更详细的日志

### 问题2: 旧任务发送到DLQ

**状态**: 文件ID=22的任务已失败并进入死信队列

**处理方式**:
1. **忽略**: 这是旧任务，可以忽略
2. **手动处理**: 从DLQ取出重新处理
3. **清空DLQ**: 删除失败的消息

**建议**: 上传新文件测试新的转码流程

---

## ✅ 总结

### 已完成 ✅

1. ✅ **Worker已启动**: 4个worker线程运行中
2. ✅ **连接到RabbitMQ**: 4个消费者已连接
3. ✅ **FFmpeg已配置**: 路径正确
4. ✅ **S3配置已加载**: bucket和region正确
5. ✅ **等待新任务**: 队列为空，准备就绪

### 待验证 ⏳

1. ⏳ **S3文件下载**: 需要上传新文件测试
2. ⏳ **转码流程**: 需要完整测试
3. ⏳ **预览文件生成**: 需要验证

### 下一步 🎯

**立即推荐**:
1. 上传一个新的测试视频文件
2. 观察Worker日志
3. 验证预览文件生成

```bash
# 监控Worker
tail -f /tmp/worker.log

# 等待上传后查看转码过程
```

---

**Worker启动时间**: 2026-02-07 06:24  
**Worker PID**: 4025524  
**状态**: ✅ 运行中，等待任务  
**日志**: `/tmp/worker.log`
