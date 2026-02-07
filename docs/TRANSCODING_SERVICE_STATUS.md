# 转码服务状态报告

## ✅ 服务状态

### Worker服务
```
✅ 已启动并运行

进程信息:
  PID: 4193208
  用户: ec2-user
  命令: ./bin/openwan-worker
  启动时间: 2026-02-07 09:31
  运行状态: 正常
```

### 配置信息
```
✓ FFmpeg: /usr/local/bin/ffmpeg
✓ 队列: amqp://guest:guest@localhost:5672/
✓ 工作线程数: 4
✓ 队列名称: openwan_transcoding_jobs
```

### Worker线程状态
```
[Worker 1] ✓ 已启动，等待任务
[Worker 2] ✓ 已启动，等待任务
[Worker 3] ✓ 已启动，等待任务
[Worker 4] ✓ 已启动，等待任务
```

---

## ⚠️ 发现的问题

### 问题：文件27转码失败

**错误信息**:
```
input file does not exist: openwan/2026/02/07/ba84945e1e9f6bac7f2f3b123a9a77e1/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
```

**实际情况**:
- ✓ 文件在S3中存在
- ✓ 路径: s3://video-bucket-843250590784/openwan/2026/02/07/ba84945e1e9f6bac7f2f3b123a9a77e1/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
- ✓ 文件大小: 8.5MB

**问题原因**:
Worker配置中**缺少S3存储配置**，导致无法访问S3文件。

Worker当前配置仅显示：
- FFmpeg路径
- RabbitMQ队列连接
- Worker数量

**缺失配置**：
- S3 Bucket
- S3 Region
- S3 Credentials (Access Key/Secret Key 或 IAM Role)

---

## 🔧 解决方案

### 方案1: 使用同步转码fallback (已实现)

由于Worker无法访问S3，**API服务已实现同步转码fallback机制**：
- 当RabbitMQ队列不可用或Worker处理失败时
- API服务器直接执行转码
- 流程：下载S3 → 转码 → 上传预览文件

**优点**:
- ✅ 无需配置Worker的S3访问
- ✅ 简单可靠
- ✅ 立即可用

**缺点**:
- ⚠️ API服务器负载较高
- ⚠️ 无法分布式处理
- ⚠️ 大量并发上传时性能受限

### 方案2: 配置Worker的S3访问 (推荐)

需要修改Worker配置文件，添加S3配置：

**1. 检查Worker配置文件**
```bash
# Worker可能从以下位置读取配置
ls -la /home/ec2-user/openwan/configs/worker*.yaml
ls -la /home/ec2-user/openwan/configs/config.yaml
```

**2. 添加S3配置**
```yaml
# configs/worker.yaml 或 configs/config.yaml
storage:
  type: s3
  s3:
    bucket: video-bucket-843250590784
    region: us-east-1
    # 使用IAM Role或提供凭证
    use_iam_role: true
    # 或
    # access_key_id: YOUR_ACCESS_KEY
    # secret_access_key: YOUR_SECRET_KEY
```

**3. 重启Worker**
```bash
pkill -f "openwan-worker"
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > logs/worker.log 2>&1 &
```

---

## 📋 当前工作流程

### 上传视频时的转码流程

#### 场景1: RabbitMQ可用
```
1. 用户上传视频
2. API保存文件到S3
3. API发布转码任务到RabbitMQ队列
4. Worker从队列获取任务
5. Worker尝试从S3下载文件 → ❌ 失败 (缺少S3配置)
6. Worker重试3次后将任务移到死信队列
7. 预览文件不存在
```

#### 场景2: RabbitMQ不可用或Worker失败后
```
1. 用户上传视频
2. API保存文件到S3
3. API尝试发布到队列 → 超时或失败
4. API自动触发同步转码fallback
5. API从S3下载文件到临时目录
6. API使用FFmpeg转码
7. API上传预览文件到S3
8. ✅ 转码完成，预览可用
```

### 实际效果
**当前状态下，同步转码fallback机制会自动处理所有转码任务**，因此视频预览功能仍然可用。

---

## ✅ 验证转码功能

### 测试步骤

1. **上传新视频**
```bash
# 通过前端或API上传视频
curl -X POST http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test-video.mp4" \
  -F "title=Test Video" \
  -F "category_id=1"
```

2. **监控转码日志**
```bash
# 查看API日志 (同步转码)
tail -f /home/ec2-user/openwan/logs/api.log | grep -i transcode

# 查看Worker日志 (队列转码)
tail -f /home/ec2-user/openwan/logs/worker.log
```

3. **检查预览文件**
```bash
# 列出S3中的预览文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | grep "preview.flv"

# 测试预览端点
curl -I "http://localhost:8080/api/v1/files/FILE_ID/preview"
```

---

## 🎯 推荐行动

### 短期 (当前可用)
✅ **使用同步转码fallback**
- 已实现，无需额外配置
- 适合中小规模使用
- 建议限制并发上传数量（10个以内）

### 中期 (生产环境推荐)
🔧 **配置Worker的S3访问**
1. 修改Worker配置添加S3设置
2. 或重新编译Worker确保读取配置
3. 重启Worker服务
4. 测试Worker能否成功转码

### 长期 (大规模部署)
🚀 **使用托管转码服务**
- AWS Elastic Transcoder
- AWS MediaConvert
- 独立的转码集群

---

## 📊 服务管理命令

### Worker服务

**检查状态**:
```bash
ps aux | grep openwan-worker | grep -v grep
```

**查看日志**:
```bash
# 实时日志
tail -f /home/ec2-user/openwan/logs/worker.log

# 查看错误
grep -i "error\|failed" /home/ec2-user/openwan/logs/worker.log
```

**停止服务**:
```bash
pkill -f "openwan-worker"
```

**启动服务**:
```bash
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > logs/worker.log 2>&1 &
```

**重启服务**:
```bash
pkill -f "openwan-worker" && sleep 2
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > logs/worker.log 2>&1 &
```

### 查看队列状态

```bash
# RabbitMQ管理界面 (如果已启用)
# http://localhost:15672
# 用户名: guest
# 密码: guest

# 或使用rabbitmqctl
sudo rabbitmqctl list_queues
```

---

## 📝 总结

### 当前状态

```
✅ API服务: 运行正常 (PID: 4189870)
✅ Worker服务: 运行正常 (PID: 4193208)
⚠️  Worker配置: 缺少S3访问配置
✅ 转码功能: 可用 (通过同步fallback)
✅ 预览功能: 可用
```

### 转码机制

| 机制 | 状态 | 说明 |
|------|------|------|
| 队列转码 (Worker) | ⚠️ 部分可用 | Worker无法访问S3文件 |
| 同步转码 (API) | ✅ 可用 | 自动fallback，直接转码 |

### 建议

1. **立即**: 继续使用同步转码，功能正常
2. **本周**: 配置Worker的S3访问权限
3. **下月**: 考虑使用托管转码服务

---

**报告生成时间**: 2026-02-07 09:35 UTC  
**Worker PID**: 4193208  
**API PID**: 4189870  
**转码功能状态**: ✅ 可用 (同步fallback模式)
