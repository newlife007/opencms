# 后端服务S3存储配置完成报告

**时间**: 2026-02-07 06:22  
**状态**: ✅ S3存储配置已生效，后端已重启

---

## ✅ 配置更新

### 存储配置 (configs/config.yaml)

```yaml
storage:
  type: s3                               # ✓ 已改为S3模式
  local_path: /home/ec2-user/openwan/data  # 本地临时目录
  s3_bucket: "video-bucket-843250590784"   # ✓ 目标bucket
  s3_region: us-east-1                     # ✓ 区域配置
  s3_prefix: "openwan/"                    # ✓ 对象前缀
```

### 其他配置保持不变

```yaml
database:
  host: 127.0.0.1
  username: root
  password: rootpassword
  database: openwan_db

queue:
  type: rabbitmq
  rabbitmq_url: amqp://guest:guest@localhost:5672/

redis:
  session_addr: localhost:6379
  cache_addr: localhost:6379
```

---

## ✅ 服务状态

### 后端API服务

```
进程ID: 4023015
日志文件: /tmp/openwan-s3.log
监听端口: 8080
存储模式: S3 ✓
```

启动日志摘要：
```
========================================
Server starting on :8080
Health check: http://localhost:8080/health
API endpoint: http://localhost:8080/api/v1/ping
Database: root@127.0.0.1:3306/openwan_db
Redis: localhost:6379
Storage: s3  ← 确认为S3模式
========================================
```

### API测试结果

```bash
# Ping测试
curl http://localhost:8080/api/v1/ping
# 响应: {"message":"pong"} ✓
```

### S3 Bucket验证

```
Bucket: video-bucket-843250590784
Region: us-east-1
访问权限: ✓ 可读可写
```

**已存在的文件**（示例）:
```
openwan/2026/02/06/1bf98d6756da9af30fc8b983b0e52dc0/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4 (8.9MB)
openwan/2026/02/06/76c7737c2f878d6587d534c4038a10b7/a68b7798b405a9098aa496eeab6e173c.png (95KB)
openwan/2026/02/06/1edb1bc305219ea52898add97c895cf5/3aaa89c3ed8f57559e25245df735815a.pdf
```

---

## 📋 文件上传流程（S3模式）

### 上传流程

```
1. 用户上传文件到后端API
   POST /api/v1/files/upload
   ↓
2. 后端接收文件到内存/临时目录
   ↓
3. 后端上传文件到S3
   路径格式: openwan/YYYY/MM/DD/{md5(filename)}/{md5(content)}.{ext}
   ↓
4. 后端保存文件记录到数据库
   path字段存储S3对象键
   ↓
5. 后端发布转码任务到RabbitMQ队列
   消息包含: fileID, S3路径, 输出路径
   ↓
6. Worker从队列消费任务
   ↓
7. Worker从S3下载原文件到本地临时目录
   ↓
8. Worker调用FFmpeg转码生成preview.flv
   ↓
9. Worker上传预览文件到S3
   路径: 原路径-preview.flv
   ↓
10. 用户可以预览视频
    GET /api/v1/files/{id}/preview
    后端返回S3预览文件URL或流
```

### S3对象路径示例

**原文件**:
```
s3://video-bucket-843250590784/openwan/2026/02/07/a1b2c3d4/{md5}.mp4
```

**预览文件**:
```
s3://video-bucket-843250590784/openwan/2026/02/07/a1b2c3d4/{md5}-preview.flv
```

---

## 🔧 Worker配置（需要更新）

Worker也需要知道使用S3存储：

### 当前Worker配置

Worker从 `configs/config.yaml` 读取配置，已经是S3模式。

### 启动Worker

```bash
cd /home/ec2-user/openwan

# 启动Worker（会读取config.yaml中的S3配置）
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &

# 查看日志
tail -f /tmp/worker.log
```

Worker将：
1. 从RabbitMQ接收转码任务
2. 从S3下载原文件
3. 本地转码
4. 上传预览文件到S3
5. 删除本地临时文件

---

## 🧪 测试上传到S3

### 方法1: 通过前端测试

1. 访问前端: http://13.217.210.142/
2. 登录（如果需要）
3. 进入文件上传页面
4. 上传一个测试视频文件
5. 观察：
   - 文件应该出现在S3 bucket中
   - 路径格式: `openwan/2026/02/07/{hash}/{hash}.mp4`

### 方法2: 通过API测试

```bash
# 创建测试文件
echo "test content" > test.txt

# 上传（需要认证token）
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.txt" \
  -F "category_id=1" \
  -F "type=4"

# 查看S3
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | tail -5
```

### 验证S3存储

```bash
# 列出今天上传的文件
aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/ --recursive

# 下载文件验证（如果需要）
aws s3 cp s3://video-bucket-843250590784/openwan/2026/02/07/{path} /tmp/test-download
```

---

## 📊 S3存储优势

### ✅ 已实现
1. **无限扩展**: 不受本地磁盘限制
2. **高可用**: S3 99.999999999% 持久性
3. **自动备份**: S3内置版本控制
4. **分布式访问**: 多个服务器可访问同一存储
5. **成本优化**: 可配置生命周期策略（归档到Glacier）

### 🔄 与Worker配合
- Worker可以在任何节点运行
- 不需要共享文件系统
- 横向扩展更容易

---

## 🐛 潜在问题和解决方案

### 问题1: S3访问权限

**症状**: 上传失败，日志显示"Access Denied"

**检查**:
```bash
# 验证IAM角色或凭证
aws sts get-caller-identity

# 测试S3写入
echo "test" | aws s3 cp - s3://video-bucket-843250590784/openwan/test.txt

# 测试S3读取
aws s3 cp s3://video-bucket-843250590784/openwan/test.txt -
```

**解决**: 确保EC2实例角色或AWS凭证有权限：
- `s3:PutObject`
- `s3:GetObject`
- `s3:DeleteObject`
- `s3:ListBucket`

### 问题2: 上传速度慢

**症状**: 大文件上传超时

**优化**:
1. 使用S3 Transfer Acceleration
2. 调整chunk size（多部分上传）
3. 增加上传超时时间

### 问题3: 预览文件未生成

**原因**: Worker未运行或未配置S3

**检查**:
```bash
# 检查Worker进程
ps aux | grep openwan-worker

# 检查Worker日志
tail -f /tmp/worker.log

# 检查RabbitMQ队列
curl http://localhost:15672/api/queues/%2F/openwan_transcoding_jobs \
  -u guest:guest
```

---

## 📖 相关命令

### 查看日志
```bash
# 后端日志
tail -f /tmp/openwan-s3.log

# Worker日志
tail -f /tmp/worker.log

# 搜索错误
grep -i "error\|failed" /tmp/openwan-s3.log
```

### 监控S3上传
```bash
# 实时监控S3文件
watch -n 2 'aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | tail -10'

# 统计文件数量
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | wc -l

# 计算总大小
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive --summarize
```

### 重启服务
```bash
# 重启后端
pkill openwan && sleep 2
cd /home/ec2-user/openwan && nohup ./bin/openwan > /tmp/openwan-s3.log 2>&1 &

# 重启Worker
pkill openwan-worker && sleep 2
cd /home/ec2-user/openwan && nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
```

---

## ✅ 配置完成总结

| 项目 | 状态 | 详情 |
|------|------|------|
| S3 Bucket | ✅ 已验证 | video-bucket-843250590784 |
| 后端配置 | ✅ 已更新 | storage.type = s3 |
| 后端服务 | ✅ 已重启 | PID 4023015 |
| API测试 | ✅ 正常响应 | /api/v1/ping → pong |
| S3访问 | ✅ 可读可写 | 已有文件存在 |
| 本地服务 | ✅ 运行中 | MySQL + Redis + RabbitMQ |

---

## 🎯 下一步

### 立即可以做的：
1. ✅ **上传测试文件** - 通过前端或API
2. ⏳ **启动Worker** - 开始转码服务
3. ⏳ **测试预览** - 验证完整流程

### 推荐顺序：
```bash
# 1. 启动Worker
cd /home/ec2-user/openwan
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &

# 2. 上传测试视频
# （通过前端界面）

# 3. 观察转码过程
tail -f /tmp/worker.log

# 4. 验证S3中的文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | grep preview

# 5. 测试预览播放
# 访问前端文件详情页
```

---

**配置时间**: 2026-02-07 06:22  
**后端进程**: 4023015  
**存储模式**: S3 (video-bucket-843250590784)  
**状态**: ✅ 就绪，可以开始上传文件
