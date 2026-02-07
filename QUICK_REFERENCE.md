# OpenWan 本地完整服务 - 快速参考

## 🚀 服务已启动

### ✅ 运行中的服务
```
✓ 后端 API:    http://localhost:8080 (PID: 4156741)
✓ Worker #1:   PID: 4156782
✓ Worker #2:   PID: 4156793
✓ 前端 Web:    http://localhost (Nginx)
✓ MySQL:       127.0.0.1:3306 (14张表)
✓ Redis:       localhost:6379
✓ RabbitMQ:    localhost:5672
✓ 存储类型:    AWS S3
✓ S3 存储桶:   video-bucket-843250590784
```

---

## 🌐 访问应用

### 前端Web应用
```
URL: http://localhost
```

### 登录凭证
```
用户名: admin
密码:   admin123
```

---

## 📁 文件存储到S3

### 配置确认
文件上传后将直接存储到AWS S3：
```
存储桶: s3://video-bucket-843250590784/
路径前缀: openwan/
目录结构: openwan/data1/{md5_hash}/{file_md5}.{ext}
```

### 测试上传步骤
1. 浏览器访问 `http://localhost`
2. 登录 (admin / admin123)
3. 点击"文件管理" → "文件上传"
4. 拖放或选择文件
5. 填写分类、类型、标题
6. 点击"开始上传"
7. 观察上传进度

### 验证S3存储
```bash
# 列出最近上传的文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | tail -10

# 查看文件总数
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | wc -l

# 下载文件验证
aws s3 cp s3://video-bucket-843250590784/openwan/data1/{path} /tmp/test-download.txt
```

---

## 🎥 视频转码流程

### 转码任务流程
1. 用户上传视频/音频文件
2. 文件上传到S3
3. 后端创建转码任务并发送到RabbitMQ
4. Worker从队列获取任务
5. Worker从S3下载原文件
6. 使用FFmpeg转码为FLV预览格式
7. Worker上传转码文件到S3
8. 更新数据库状态

### 监控转码任务
```bash
# 查看Worker日志
tail -f logs/worker-1.log
tail -f logs/worker-2.log

# 查看RabbitMQ队列
curl -u guest:guest http://localhost:15672/api/queues
```

---

## 📊 管理命令

### 停止服务
```bash
cd /home/ec2-user/openwan
./stop-services.sh
```

### 重启服务
```bash
./stop-services.sh
./start-services.sh
```

### 查看日志
```bash
# API日志
tail -f logs/api.log

# Worker日志
tail -f logs/worker-1.log
tail -f logs/worker-2.log

# Nginx日志
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### 健康检查
```bash
# API健康
curl http://localhost:8080/health

# Ping测试
curl http://localhost:8080/api/v1/ping

# 前端访问
curl -I http://localhost
```

---

## 🗄️ 数据库访问

### 连接数据库
```bash
mysql -h 127.0.0.1 -u openwan -popenwan123 -D openwan_db
```

### 查看文件记录
```sql
-- 查看所有文件
SELECT id, title, type, status, size, path, upload_at FROM ow_files ORDER BY upload_at DESC LIMIT 10;

-- 查看最近上传
SELECT * FROM ow_files WHERE DATE(upload_at) = CURDATE();

-- 按类型统计
SELECT type, COUNT(*) as count FROM ow_files GROUP BY type;

-- 按状态统计
SELECT status, COUNT(*) as count FROM ow_files GROUP BY status;
```

---

## 🧪 测试脚本

### S3上传测试
```bash
cd /home/ec2-user/openwan
./test-s3-upload.sh
```

这个脚本会：
1. 验证服务状态
2. 检查S3连接
3. 创建测试文件
4. 提供详细的上传测试步骤
5. 显示验证命令

---

## 🔍 故障排查

### API服务无响应
```bash
# 检查进程
ps aux | grep openwan

# 查看日志
tail -100 logs/api.log

# 重启服务
./stop-services.sh && ./start-services.sh
```

### S3上传失败
```bash
# 检查AWS凭证
cat ~/.aws/credentials

# 测试S3访问
aws s3 ls s3://video-bucket-843250590784/

# 检查IAM权限
aws sts get-caller-identity

# 查看上传日志
tail -f logs/api.log | grep -i "upload\|s3"
```

### 转码任务失败
```bash
# 查看Worker日志
tail -100 logs/worker-1.log

# 检查RabbitMQ
sudo systemctl status rabbitmq-server

# 检查FFmpeg
/usr/local/bin/ffmpeg -version

# 重启Worker
pkill openwan-worker
./bin/openwan-worker > logs/worker-1.log 2>&1 &
```

### 前端无法访问
```bash
# 检查Nginx
sudo systemctl status nginx

# 重载Nginx
sudo systemctl reload nginx

# 查看Nginx错误日志
sudo tail -50 /var/log/nginx/error.log

# 检查前端构建
ls -lh /home/ec2-user/openwan/frontend/dist/
```

---

## 📚 相关文档

- **完整服务状态**: `docs/SERVICE_STATUS_S3.md`
- **API文档**: `docs/api.md`
- **部署指南**: `docs/deployment.md`
- **S3配置**: `docs/S3_SETUP.md`
- **国际化**: `docs/I18N_VERIFICATION_REPORT.md`

---

## 🎯 下一步操作

1. **测试文件上传**
   - 访问 http://localhost
   - 登录并上传测试文件
   - 验证文件已存储到S3

2. **测试视频转码**
   - 上传视频文件 (MP4/AVI/MOV)
   - 监控Worker日志
   - 验证预览文件生成
   - 验证转码文件上传到S3

3. **测试搜索功能**
   - 上传多个文件
   - 使用搜索功能查找
   - 验证结果准确性

4. **测试权限控制**
   - 创建不同角色用户
   - 测试文件访问权限
   - 验证RBAC功能

5. **性能测试**
   - 上传大文件 (>100MB)
   - 并发上传多个文件
   - 监控系统资源使用

---

## 💡 提示

- **清除浏览器缓存**: 如果前端显示旧版本，按 `Ctrl+Shift+R` 硬刷新
- **查看实时日志**: 使用 `tail -f` 命令监控日志文件
- **S3费用**: 注意S3存储和请求会产生费用，测试完成后可删除测试文件
- **Worker扩展**: 如需更多转码能力，可启动更多Worker实例
- **数据库备份**: 定期备份数据库以防数据丢失

---

**文档更新时间**: 2026-02-07 08:52 UTC
