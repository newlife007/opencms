# OpenWan 部署指南

本文档提供OpenWan系统的详细部署说明，包括本地开发、Docker部署和Kubernetes生产部署。

## 目录

- [环境要求](#环境要求)
- [本地开发部署](#本地开发部署)
- [Docker部署](#docker部署)
- [Kubernetes生产部署](#kubernetes生产部署)
- [配置说明](#配置说明)
- [故障排查](#故障排查)

## 环境要求

### 最低配置
- CPU: 4核心
- 内存: 8GB
- 磁盘: 100GB
- 操作系统: Linux (Ubuntu 20.04+ / CentOS 7+) / macOS

### 推荐配置（生产环境）
- CPU: 8核心+
- 内存: 16GB+
- 磁盘: 500GB+ SSD
- 操作系统: Linux (Ubuntu 22.04 LTS)

### 软件要求
- Go 1.25.5+
- Node.js 18+
- MySQL 5.7+ / 8.0+
- Redis 6.0+
- RabbitMQ 3.9+ (可选，使用Amazon SQS则不需要)
- FFmpeg 4.0+ (with H.264, AAC codecs)
- Sphinx 2.2+ (可选，用于搜索功能)
- Docker 20.10+ (Docker部署)
- Kubernetes 1.24+ (K8s部署)

## 本地开发部署

### 1. 安装依赖

#### 安装Go
```bash
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 安装Node.js
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### 安装MySQL
```bash
sudo apt-get update
sudo apt-get install mysql-server
sudo mysql_secure_installation
```

#### 安装Redis
```bash
sudo apt-get install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

#### 安装FFmpeg
```bash
sudo apt-get install ffmpeg
ffmpeg -version  # 验证安装
```

#### 安装RabbitMQ (可选)
```bash
sudo apt-get install rabbitmq-server
sudo systemctl start rabbitmq-server
sudo systemctl enable rabbitmq-server
sudo rabbitmq-plugins enable rabbitmq_management
```

### 2. 数据库初始化

```bash
# 登录MySQL
mysql -u root -p

# 创建数据库
CREATE DATABASE openwan_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 创建用户
CREATE USER 'openwan'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON openwan_db.* TO 'openwan'@'localhost';
FLUSH PRIVILEGES;
exit;
```

### 3. 配置后端

```bash
cd /home/ec2-user/openwan

# 复制配置文件
cp configs/config.yaml.example configs/config.yaml

# 编辑配置
vim configs/config.yaml
```

配置示例：
```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 3306
  database: openwan_db
  username: openwan
  password: your_password

redis:
  host: localhost
  port: 6379

storage:
  type: local
  local_path: ./storage

ffmpeg:
  path: /usr/bin/ffmpeg
  worker_count: 4
```

### 4. 运行数据库迁移

```bash
# 安装golang-migrate
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 运行迁移
migrate -path ./migrations -database "mysql://openwan:your_password@tcp(localhost:3306)/openwan_db" up
```

### 5. 启动后端服务

```bash
# 安装依赖
go mod download

# 启动API服务
go run cmd/api/main.go

# 另一个终端启动Worker
go run cmd/worker/main.go
```

API服务将在 `http://localhost:8080` 运行

### 6. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 运行

### 7. 访问系统

打开浏览器访问：http://localhost:5173

默认管理员账号：
- 用户名: `admin`
- 密码: `admin123`

## Docker部署

### 1. 安装Docker

```bash
# Ubuntu
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. 配置环境变量

```bash
cp .env.example .env
vim .env
```

`.env` 文件示例：
```bash
# 数据库
MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=openwan_db
MYSQL_USER=openwan
MYSQL_PASSWORD=your_password

# Redis
REDIS_PASSWORD=your_redis_password

# RabbitMQ
RABBITMQ_DEFAULT_USER=admin
RABBITMQ_DEFAULT_PASS=admin

# 应用
API_PORT=8080
FRONTEND_PORT=80

# 存储
STORAGE_TYPE=local
```

### 3. 构建并启动

```bash
# 构建镜像
docker-compose build

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看运行状态
docker-compose ps
```

### 4. 初始化数据

```bash
# 运行数据库迁移
docker-compose exec api migrate -path /app/migrations -database "mysql://openwan:your_password@mysql:3306/openwan_db" up

# 或者使用应用内置迁移
docker-compose exec api ./api migrate up
```

### 5. 访问系统

- 前端: http://localhost
- API: http://localhost:8080
- RabbitMQ管理界面: http://localhost:15672

### 6. 常用命令

```bash
# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v

# 重启特定服务
docker-compose restart api

# 查看特定服务日志
docker-compose logs -f api

# 进入容器
docker-compose exec api sh

# 更新镜像
docker-compose pull
docker-compose up -d
```

## Kubernetes生产部署

### 1. 前置准备

```bash
# 安装kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# 安装helm (可选)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### 2. 创建命名空间

```bash
kubectl create namespace openwan
```

### 3. 创建配置和密钥

```bash
# 创建ConfigMap
kubectl create configmap openwan-config \
  --from-file=configs/config.yaml \
  -n openwan

# 创建Secrets
kubectl create secret generic openwan-secrets \
  --from-literal=db-password=your_password \
  --from-literal=redis-password=your_redis_password \
  --from-literal=s3-access-key=your_access_key \
  --from-literal=s3-secret-key=your_secret_key \
  -n openwan
```

### 4. 部署MySQL

```bash
# 使用StatefulSet部署MySQL
kubectl apply -f deployments/k8s/mysql-statefulset.yaml -n openwan

# 或使用云服务商的托管数据库 (推荐)
# - AWS RDS
# - Azure Database for MySQL
# - Google Cloud SQL
```

### 5. 部署Redis

```bash
kubectl apply -f deployments/k8s/redis-deployment.yaml -n openwan

# 或使用云服务商的托管Redis (推荐)
# - AWS ElastiCache
# - Azure Cache for Redis
# - Google Cloud Memorystore
```

### 6. 部署RabbitMQ

```bash
kubectl apply -f deployments/k8s/rabbitmq-deployment.yaml -n openwan

# 或使用云服务商的托管消息队列 (推荐)
# - Amazon SQS
# - Azure Service Bus
# - Google Cloud Pub/Sub
```

### 7. 部署应用

```bash
# 部署API服务
kubectl apply -f deployments/k8s/api-deployment.yaml -n openwan

# 部署Worker服务
kubectl apply -f deployments/k8s/worker-deployment.yaml -n openwan

# 部署前端
kubectl apply -f deployments/k8s/frontend-deployment.yaml -n openwan
```

### 8. 配置Ingress

```bash
# 安装Ingress Controller (如果没有)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# 部署Ingress
kubectl apply -f deployments/k8s/ingress.yaml -n openwan
```

### 9. 配置自动扩展

```bash
# 配置HPA
kubectl apply -f deployments/k8s/hpa.yaml -n openwan

# 配置Cluster Autoscaler (云环境)
# 参考云服务商文档
```

### 10. 验证部署

```bash
# 查看所有资源
kubectl get all -n openwan

# 查看Pod状态
kubectl get pods -n openwan

# 查看日志
kubectl logs -f deployment/openwan-api -n openwan

# 查看服务
kubectl get svc -n openwan

# 查看Ingress
kubectl get ingress -n openwan
```

### 11. 配置持久化存储

```bash
# 创建StorageClass (如果使用云存储)
kubectl apply -f deployments/k8s/storageclass.yaml

# 创建PVC
kubectl apply -f deployments/k8s/pvc.yaml -n openwan
```

### 12. 配置监控

```bash
# 安装Prometheus Operator
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --create-namespace

# 部署ServiceMonitor
kubectl apply -f deployments/k8s/servicemonitor.yaml -n openwan

# 访问Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```

## 配置说明

### 环境配置优先级

1. 环境变量
2. 配置文件 (config.yaml)
3. 默认值

### 关键配置项

#### 数据库配置
```yaml
database:
  host: localhost
  port: 3306
  database: openwan_db
  username: openwan
  password: ${DB_PASSWORD}  # 支持环境变量
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600s
```

#### Redis配置
```yaml
redis:
  session:
    host: localhost
    port: 6379
    password: ${REDIS_PASSWORD}
    db: 0
  cache:
    host: localhost
    port: 6379
    password: ${REDIS_PASSWORD}
    db: 1
```

#### 存储配置
```yaml
storage:
  type: s3  # local or s3
  local:
    path: ./storage
    max_files_per_dir: 65535
  s3:
    bucket: your-bucket
    region: us-east-1
    access_key: ${S3_ACCESS_KEY}
    secret_key: ${S3_SECRET_KEY}
    endpoint: ""  # 自定义endpoint (MinIO等)
```

#### FFmpeg配置
```yaml
ffmpeg:
  path: /usr/bin/ffmpeg
  worker_count: 4
  timeout: 3600s
  preview_params: "-y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240"
```

## 故障排查

### 常见问题

#### 1. 数据库连接失败

**症状**: 启动时报错 "failed to connect to database"

**解决方案**:
```bash
# 检查MySQL是否运行
systemctl status mysql

# 检查连接参数
mysql -h localhost -u openwan -p

# 检查防火墙
sudo ufw allow 3306
```

#### 2. Redis连接失败

**症状**: 启动时报错 "failed to connect to redis"

**解决方案**:
```bash
# 检查Redis是否运行
systemctl status redis

# 测试连接
redis-cli ping

# 检查配置
cat /etc/redis/redis.conf | grep bind
```

#### 3. FFmpeg转码失败

**症状**: 文件上传成功但转码失败

**解决方案**:
```bash
# 检查FFmpeg安装
ffmpeg -version

# 检查codecs
ffmpeg -codecs | grep h264
ffmpeg -codecs | grep aac

# 手动测试转码
ffmpeg -i input.mp4 -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 output.flv
```

#### 4. 文件上传失败

**症状**: 上传文件时返回错误

**解决方案**:
```bash
# 检查存储目录权限
ls -la ./storage
chmod 755 ./storage

# 检查磁盘空间
df -h

# 检查Nginx配置 (如果使用Nginx)
# client_max_body_size 500M;
```

#### 5. Pod无法启动 (K8s)

**症状**: Pod状态为 CrashLoopBackOff

**解决方案**:
```bash
# 查看Pod日志
kubectl logs <pod-name> -n openwan

# 查看Pod事件
kubectl describe pod <pod-name> -n openwan

# 检查ConfigMap
kubectl get configmap -n openwan

# 检查Secrets
kubectl get secrets -n openwan
```

### 性能优化

#### 数据库优化
```sql
-- 添加索引
CREATE INDEX idx_files_status ON ow_files(status);
CREATE INDEX idx_files_type ON ow_files(type);
CREATE INDEX idx_files_category_id ON ow_files(category_id);
CREATE INDEX idx_files_upload_at ON ow_files(upload_at);

-- 优化配置
SET GLOBAL max_connections = 500;
SET GLOBAL innodb_buffer_pool_size = 4G;
```

#### Redis优化
```bash
# /etc/redis/redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

#### 应用优化
```yaml
# 增加worker数量
ffmpeg:
  worker_count: 8

# 调整连接池
database:
  max_open_conns: 200
  max_idle_conns: 50
```

## 备份与恢复

### 数据库备份

```bash
# 备份
mysqldump -u openwan -p openwan_db > backup_$(date +%Y%m%d).sql

# 恢复
mysql -u openwan -p openwan_db < backup_20240101.sql
```

### 文件备份

```bash
# 本地存储备份
tar -czf storage_backup_$(date +%Y%m%d).tar.gz ./storage

# S3备份 (如果使用S3)
aws s3 sync s3://your-bucket s3://backup-bucket --storage-class GLACIER
```

## 升级指南

### 滚动更新 (K8s)

```bash
# 更新镜像
kubectl set image deployment/openwan-api api=openwan/api:v2.0.0 -n openwan

# 查看更新状态
kubectl rollout status deployment/openwan-api -n openwan

# 回滚
kubectl rollout undo deployment/openwan-api -n openwan
```

### 蓝绿部署

```bash
# 部署新版本
kubectl apply -f deployments/k8s/api-deployment-v2.yaml -n openwan

# 切换流量
kubectl patch service openwan-api -p '{"spec":{"selector":{"version":"v2.0.0"}}}' -n openwan

# 验证后删除旧版本
kubectl delete deployment openwan-api-v1 -n openwan
```

## 安全建议

1. **使用HTTPS**: 配置SSL证书
2. **定期更新**: 保持系统和依赖最新
3. **最小权限**: 使用专用账号运行服务
4. **网络隔离**: 使用VPC和安全组
5. **日志审计**: 启用审计日志
6. **备份策略**: 定期备份并测试恢复
7. **监控告警**: 配置监控和告警规则
8. **密钥管理**: 使用密钥管理服务 (KMS)

## 联系支持

如有部署问题，请：
1. 查看日志文件
2. 搜索GitHub Issues
3. 提交新Issue并附上日志

---

最后更新: 2024-02-01
