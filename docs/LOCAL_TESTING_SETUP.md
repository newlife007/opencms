# 本地测试环境配置指南

**目标**: 在本地部署所有服务进行功能测试  
**时间**: 2026-02-07 06:20

---

## ✅ 已完成的配置

### 1. 本地服务（Docker Compose）

所有依赖服务已启动并运行：

```bash
# 查看服务状态
sudo docker-compose ps
```

| 服务 | 状态 | 端口 |
|------|------|------|
| MySQL | ✅ Running (healthy) | 3306 |
| Redis | ✅ Running (healthy) | 6379 |
| RabbitMQ | ✅ Running (healthy) | 5672, 15672 |

### 2. 配置文件更新

`configs/config.yaml` 已更新为本地配置：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  database: openwan_db
  username: root
  password: rootpassword  # Docker MySQL密码
  max_conns: 100

storage:
  type: local  # 改为本地存储
  local_path: /home/ec2-user/openwan/data

redis:
  session_addr: localhost:6379
  cache_addr: localhost:6379
  password: ""
  db: 0

queue:
  type: rabbitmq  # 改回RabbitMQ
  rabbitmq_url: amqp://guest:guest@localhost:5672/
  queues:
    transcoding: openwan_transcoding_jobs
    notifications: openwan_notifications
```

### 3. 存储目录

```bash
/home/ec2-user/openwan/data/  # 本地文件存储
```

---

## 🔧 快速启动指南

### 步骤1: 验证Docker服务

```bash
cd /home/ec2-user/openwan

# 启动所有本地服务
sudo docker-compose up -d mysql redis rabbitmq

# 检查健康状态
sudo docker-compose ps

# 查看日志（如果有问题）
sudo docker-compose logs mysql
sudo docker-compose logs redis
sudo docker-compose logs rabbitmq
```

### 步骤2: 初始化数据库

```bash
# 运行数据库迁移
./bin/migrate -path migrations -database "mysql://root:rootpassword@tcp(127.0.0.1:3306)/openwan_db" up

# 或者使用Docker exec
sudo docker exec -i openwan-mysql-1 mysql -uroot -prootpassword openwan_db < migrations/000001_init_schema.up.sql
```

### 步骤3: 启动后端API

```bash
cd /home/ec2-user/openwan

# 停止旧进程
pkill openwan

# 启动新进程
nohup ./bin/openwan > /tmp/openwan-local.log 2>&1 &

# 查看日志
tail -f /tmp/openwan-local.log

# 测试健康检查
curl http://localhost:8080/health
```

### 步骤4: 启动转码Worker

```bash
cd /home/ec2-user/openwan

# 启动worker
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &

# 查看日志
tail -f /tmp/worker.log
```

### 步骤5: 访问前端

前端已通过Nginx部署：
```
http://13.217.210.142/
```

---

## 📋 测试流程

### 1. 测试后端API

```bash
# 健康检查
curl http://localhost:8080/health

# Ping测试
curl http://localhost:8080/api/v1/ping

# 查看路由（如果有debug endpoint）
curl http://localhost:8080/routes
```

### 2. 测试文件上传

通过前端或curl上传测试文件：

```bash
# 创建测试视频（使用FFmpeg）
ffmpeg -f lavfi -i testsrc=duration=10:size=320x240:rate=30 -pix_fmt yuv420p test-video.mp4

# 上传文件（需要先登录获取token）
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test-video.mp4" \
  -F "category_id=1" \
  -F "type=1"
```

### 3. 观察转码过程

```bash
# 终端1: 监控Worker日志
tail -f /tmp/worker.log

# 终端2: 监控RabbitMQ队列
# 访问 http://localhost:15672
# 用户名: guest, 密码: guest

# 终端3: 监控FFmpeg进程
watch -n 1 'ps aux | grep ffmpeg'

# 终端4: 监控存储目录
watch -n 1 'ls -lh /home/ec2-user/openwan/data/'
```

### 4. 测试视频预览

```bash
# 假设文件ID=1
curl -I http://localhost:8080/api/v1/files/1/preview

# 应该返回200 OK（如果转码完成）
# 或404（如果转码未完成或文件不存在）
```

---

## 🐛 常见问题排查

### 问题1: 后端无法连接数据库

**症状**:
```
Failed to initialize database: connection refused
```

**解决**:
```bash
# 检查MySQL是否运行
sudo docker ps | grep mysql

# 检查MySQL日志
sudo docker logs openwan-mysql-1

# 测试连接
mysql -h 127.0.0.1 -u root -prootpassword -e "SHOW DATABASES;"
```

### 问题2: Worker无法连接RabbitMQ

**症状**:
```
Failed to connect to message queue
```

**解决**:
```bash
# 检查RabbitMQ状态
sudo docker ps | grep rabbitmq

# 测试连接
curl http://localhost:15672/api/overview -u guest:guest

# 检查RabbitMQ日志
sudo docker logs openwan-rabbitmq-1
```

### 问题3: FFmpeg转码失败

**症状**:
```
Worker日志显示: FFmpeg failed with exit code 1
```

**解决**:
```bash
# 检查FFmpeg是否安装
which ffmpeg
ffmpeg -version

# 手动测试转码
ffmpeg -i /path/to/test.mp4 -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 /tmp/test-preview.flv

# 检查磁盘空间
df -h /tmp
```

### 问题4: 预览文件404

**可能原因**:
1. 文件未上传
2. 转码未完成
3. Worker未运行
4. 存储路径错误

**排查步骤**:
```bash
# 1. 检查文件是否在数据库
mysql -h 127.0.0.1 -u root -prootpassword openwan_db -e "SELECT id, name, path, type FROM ow_files LIMIT 10;"

# 2. 检查存储目录
ls -lR /home/ec2-user/openwan/data/ | grep -E "\.mp4|\.flv"

# 3. 检查Worker是否运行
ps aux | grep openwan-worker

# 4. 检查RabbitMQ队列
# 访问 http://localhost:15672 查看消息数量
```

---

## 🧪 完整测试场景

### 场景1: 上传并转码视频

```bash
# 1. 创建测试视频
ffmpeg -f lavfi -i testsrc=duration=10:size=320x240:rate=30 test.mp4

# 2. 通过前端上传（或使用API）

# 3. 观察转码过程
# Worker日志应该显示:
#   - Received transcode job for file ID=X
#   - Starting FFmpeg transcoding
#   - Transcoding completed
#   - Preview file saved

# 4. 验证预览文件
curl -I http://localhost:8080/api/v1/files/X/preview
# 应该返回200 OK

# 5. 在浏览器中测试播放器
# 打开 http://13.217.210.142/files/X
# 测试时间轴拖拽功能
```

### 场景2: 测试播放器修复

```bash
# 1. 确保有可用的视频文件（上面已上传）

# 2. 在浏览器打开
http://13.217.210.142/files/X

# 3. 硬刷新清除缓存
Ctrl + Shift + R

# 4. 打开开发者工具（F12）
# Console应该显示:
#   - Video player ready
#   - SeekBar enabled for dragging
#   - Video metadata loaded

# 5. 测试拖拽
# - 点击进度条 → 跳转 ✓
# - 拖拽时间球 → 流畅 ✓
# - 光标变化 → pointer/grab/grabbing ✓
```

---

## 📊 监控面板

### RabbitMQ管理界面
```
URL: http://localhost:15672
用户名: guest
密码: guest
```

查看：
- Queues: 队列深度，消息速率
- Connections: Worker连接状态
- Channels: 活跃通道

### MySQL数据库
```bash
# 查看文件表
mysql -h 127.0.0.1 -u root -prootpassword openwan_db \
  -e "SELECT id, title, type, status, path FROM ow_files ORDER BY id DESC LIMIT 10;"

# 查看转码任务（如果有表）
mysql -h 127.0.0.1 -u root -prootpassword openwan_db \
  -e "SELECT * FROM ow_transcode_jobs ORDER BY id DESC LIMIT 10;"
```

---

## 🚀 自动化启动脚本

创建 `start-local.sh`:

```bash
#!/bin/bash
set -e

echo "========================================="
echo "OpenWan本地测试环境启动脚本"
echo "========================================="

# 1. 启动Docker服务
echo "1. 启动Docker服务..."
sudo docker-compose up -d mysql redis rabbitmq
sleep 5

# 2. 检查服务健康
echo "2. 检查服务健康..."
sudo docker-compose ps

# 3. 停止旧进程
echo "3. 停止旧进程..."
pkill openwan || true
pkill openwan-worker || true
sleep 2

# 4. 启动后端
echo "4. 启动后端API..."
cd /home/ec2-user/openwan
nohup ./bin/openwan > /tmp/openwan-local.log 2>&1 &
echo "后端PID: $!"
sleep 3

# 5. 启动Worker
echo "5. 启动转码Worker..."
nohup ./bin/openwan-worker > /tmp/worker.log 2>&1 &
echo "Worker PID: $!"
sleep 2

# 6. 验证服务
echo "6. 验证服务..."
echo "后端健康检查:"
curl -s http://localhost:8080/health | head -20

echo ""
echo "========================================="
echo "✓ 所有服务已启动"
echo "========================================="
echo "后端日志: tail -f /tmp/openwan-local.log"
echo "Worker日志: tail -f /tmp/worker.log"
echo "前端地址: http://13.217.210.142/"
echo "RabbitMQ管理: http://localhost:15672 (guest/guest)"
echo "========================================="
```

使用方法：
```bash
chmod +x start-local.sh
./start-local.sh
```

---

## 📖 下一步

1. **初始化数据库**: 运行迁移创建表结构
2. **创建测试用户**: 通过API或直接SQL插入
3. **上传测试文件**: 验证上传和转码流程
4. **测试播放器**: 验证拖拽功能修复

---

**配置完成时间**: 2026-02-07 06:30  
**当前状态**: 本地服务已启动，等待数据库初始化和功能测试  
**文档路径**: `/home/ec2-user/openwan/docs/LOCAL_TESTING_SETUP.md`
