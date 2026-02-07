# OpenWan AWS 云上部署架构设计

**版本**: v1.0  
**更新日期**: 2026-02-01  
**设计目标**: 高可用、安全、可观测、可扩展

---

## 📋 目录

1. [架构概览](#架构概览)
2. [网络架构](#网络架构)
3. [计算资源](#计算资源)
4. [存储方案](#存储方案)
5. [数据库方案](#数据库方案)
6. [缓存和队列](#缓存和队列)
7. [安全架构](#安全架构)
8. [可观测性](#可观测性)
9. [成本估算](#成本估算)
10. [部署脚本](#部署脚本)

---

## 架构概览

### 整体架构图

```
                                    Internet
                                       │
                                       ▼
                            ┌──────────────────┐
                            │   Route 53 DNS   │
                            └──────────────────┘
                                       │
                                       ▼
                            ┌──────────────────┐
                            │   CloudFront     │
                            │   (CDN + WAF)    │
                            └──────────────────┘
                                       │
                ┌──────────────────────┼──────────────────────┐
                │                      │                      │
                ▼                      ▼                      ▼
        ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
        │   ALB (AZ-A) │      │   ALB (AZ-B) │      │   ALB (AZ-C) │
        └──────────────┘      └──────────────┘      └──────────────┘
                │                      │                      │
    ┌───────────┴──────────────────────┴──────────────────────┴───────────┐
    │                          VPC (10.0.0.0/16)                          │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                     Public Subnets (DMZ)                      │  │
    │  │  ┌────────────┐    ┌────────────┐    ┌────────────┐        │  │
    │  │  │ NAT GW (A) │    │ NAT GW (B) │    │ NAT GW (C) │        │  │
    │  │  └────────────┘    └────────────┘    └────────────┘        │  │
    │  │  10.0.1.0/24       10.0.2.0/24       10.0.3.0/24            │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                   Application Subnets                        │  │
    │  │  ┌─────────────────────────────────────────────────────┐    │  │
    │  │  │            ECS/EKS Cluster (Auto Scaling)           │    │  │
    │  │  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐   │    │  │
    │  │  │  │Backend │  │Backend │  │Worker  │  │Worker  │   │    │  │
    │  │  │  │ (2-10) │  │ (2-10) │  │ (2-10) │  │ (2-10) │   │    │  │
    │  │  │  └────────┘  └────────┘  └────────┘  └────────┘   │    │  │
    │  │  └─────────────────────────────────────────────────────┘    │  │
    │  │  10.0.11.0/24      10.0.12.0/24      10.0.13.0/24           │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                      Data Subnets (Private)                  │  │
    │  │  ┌──────────┐    ┌──────────┐    ┌──────────┐              │  │
    │  │  │ RDS (M)  │───▶│ RDS (S1) │    │ RDS (S2) │              │  │
    │  │  │ Primary  │    │ Replica  │    │ Replica  │              │  │
    │  │  └──────────┘    └──────────┘    └──────────┘              │  │
    │  │                                                               │  │
    │  │  ┌──────────────────┐    ┌──────────────────┐              │  │
    │  │  │ ElastiCache      │    │ ElastiCache      │              │  │
    │  │  │ Redis (Primary)  │───▶│ Redis (Replica)  │              │  │
    │  │  └──────────────────┘    └──────────────────┘              │  │
    │  │                                                               │  │
    │  │  10.0.21.0/24      10.0.22.0/24      10.0.23.0/24           │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                    External Services                         │  │
    │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐            │  │
    │  │  │    S3      │  │   SQS      │  │ Secrets    │            │  │
    │  │  │  (Media)   │  │  (Queue)   │  │  Manager   │            │  │
    │  │  └────────────┘  └────────────┘  └────────────┘            │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                  Observability Layer                         │  │
    │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐            │  │
    │  │  │ CloudWatch │  │   X-Ray    │  │ CloudWatch │            │  │
    │  │  │   Logs     │  │  (Tracing) │  │  Metrics   │            │  │
    │  │  └────────────┘  └────────────┘  └────────────┘            │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    │                                                                      │
    │  ┌──────────────────────────────────────────────────────────────┐  │
    │  │                     Security Layer                           │  │
    │  │  ┌────────────┐  ┌────────────┐  ┌────────────┐            │  │
    │  │  │   WAF      │  │  Security  │  │   GuardDuty│            │  │
    │  │  │  Rules     │  │   Groups   │  │   (Threat) │            │  │
    │  │  └────────────┘  └────────────┘  └────────────┘            │  │
    │  │  ┌────────────┐  ┌────────────┐                             │  │
    │  │  │   IAM      │  │   KMS      │                             │  │
    │  │  │  Roles     │  │ (Encrypt)  │                             │  │
    │  │  └────────────┘  └────────────┘                             │  │
    │  └──────────────────────────────────────────────────────────────┘  │
    └──────────────────────────────────────────────────────────────────────┘
```

### 架构特点

✅ **高可用性**:
- 3个可用区 (AZ-A, AZ-B, AZ-C)
- RDS Multi-AZ 自动故障转移
- ElastiCache 集群模式
- ALB 跨AZ负载均衡
- Auto Scaling 自动扩缩容

✅ **安全性**:
- VPC 网络隔离
- 私有子网 (数据库、缓存)
- 公有子网 (NAT Gateway)
- Security Groups 精细控制
- WAF 防护
- KMS 加密
- Secrets Manager 密钥管理
- GuardDuty 威胁检测

✅ **可观测性**:
- CloudWatch Logs 集中日志
- CloudWatch Metrics 监控指标
- X-Ray 分布式追踪
- CloudWatch Alarms 告警
- CloudWatch Dashboard 可视化

✅ **可扩展性**:
- ECS/EKS 容器编排
- Auto Scaling (2-20实例)
- S3 无限存储
- ElastiCache 集群扩展
- RDS 读副本扩展

---

## 网络架构

### VPC 设计

**CIDR**: 10.0.0.0/16 (65,536 IPs)

#### 子网划分

| 子网类型 | 用途 | CIDR | 可用IP | AZ |
|---------|------|------|--------|-----|
| **Public Subnets** (DMZ) |
| public-subnet-a | NAT GW, Bastion | 10.0.1.0/24 | 251 | us-east-1a |
| public-subnet-b | NAT GW | 10.0.2.0/24 | 251 | us-east-1b |
| public-subnet-c | NAT GW | 10.0.3.0/24 | 251 | us-east-1c |
| **Application Subnets** (Private) |
| app-subnet-a | Backend/Worker | 10.0.11.0/24 | 251 | us-east-1a |
| app-subnet-b | Backend/Worker | 10.0.12.0/24 | 251 | us-east-1b |
| app-subnet-c | Backend/Worker | 10.0.13.0/24 | 251 | us-east-1c |
| **Data Subnets** (Private) |
| data-subnet-a | RDS, ElastiCache | 10.0.21.0/24 | 251 | us-east-1a |
| data-subnet-b | RDS, ElastiCache | 10.0.22.0/24 | 251 | us-east-1b |
| data-subnet-c | RDS, ElastiCache | 10.0.23.0/24 | 251 | us-east-1c |

#### 路由表

**Public Route Table**:
```
Destination     Target
10.0.0.0/16    local
0.0.0.0/0      igw-xxxx (Internet Gateway)
```

**Private Route Table (App Subnets)**:
```
Destination     Target
10.0.0.0/16    local
0.0.0.0/0      nat-xxxx (NAT Gateway)
```

**Private Route Table (Data Subnets)**:
```
Destination     Target
10.0.0.0/16    local
(No internet access)
```

### Security Groups

#### ALB Security Group
```
Inbound:
- 443 (HTTPS) from 0.0.0.0/0
- 80 (HTTP) from 0.0.0.0/0 (redirect to 443)

Outbound:
- 8080 (Backend) to Backend SG
```

#### Backend Security Group
```
Inbound:
- 8080 (HTTP) from ALB SG
- 9090 (Metrics) from Monitoring SG

Outbound:
- 3306 (MySQL) to RDS SG
- 6379 (Redis) to ElastiCache SG
- 443 (HTTPS) to 0.0.0.0/0 (AWS APIs, S3, SQS)
```

#### Worker Security Group
```
Inbound:
- None (workers don't accept incoming connections)

Outbound:
- 3306 (MySQL) to RDS SG
- 6379 (Redis) to ElastiCache SG
- 443 (HTTPS) to 0.0.0.0/0 (AWS APIs, S3, SQS)
```

#### RDS Security Group
```
Inbound:
- 3306 (MySQL) from Backend SG
- 3306 (MySQL) from Worker SG
- 3306 (MySQL) from Bastion SG (management)

Outbound:
- None
```

#### ElastiCache Security Group
```
Inbound:
- 6379 (Redis) from Backend SG
- 6379 (Redis) from Worker SG

Outbound:
- None
```

---

## 计算资源

### ECS Fargate (推荐) / EKS

**选择理由**:
- 无需管理EC2实例
- 自动扩缩容
- 按使用付费
- 高可用性

#### Backend Service

**Task Definition**:
```yaml
Family: openwan-backend
CPU: 2 vCPU (2048)
Memory: 4 GB (4096)
Container:
  Name: backend
  Image: <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/openwan-backend:latest
  Port: 8080, 9090
  Environment:
    - DB_HOST: <RDS_ENDPOINT>
    - REDIS_HOST: <ELASTICACHE_ENDPOINT>
    - S3_BUCKET: <BUCKET_NAME>
    - SQS_QUEUE_URL: <QUEUE_URL>
  Secrets:
    - DB_PASSWORD: arn:aws:secretsmanager:...
    - JWT_SECRET: arn:aws:secretsmanager:...
  Logging:
    LogDriver: awslogs
    Options:
      awslogs-group: /ecs/openwan-backend
      awslogs-region: us-east-1
      awslogs-stream-prefix: ecs
  HealthCheck:
    Command: ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"]
    Interval: 30
    Timeout: 5
    Retries: 3
```

**Service Configuration**:
```yaml
ServiceName: openwan-backend
DesiredCount: 2
LaunchType: FARGATE
NetworkMode: awsvpc
Subnets:
  - app-subnet-a
  - app-subnet-b
  - app-subnet-c
SecurityGroups:
  - backend-sg
LoadBalancers:
  - TargetGroupArn: <ALB_TARGET_GROUP>
    ContainerName: backend
    ContainerPort: 8080
AutoScaling:
  MinCapacity: 2
  MaxCapacity: 20
  TargetValue:
    CPUUtilization: 70%
    MemoryUtilization: 80%
```

#### Worker Service

**Task Definition**:
```yaml
Family: openwan-worker
CPU: 4 vCPU (4096)
Memory: 8 GB (8192)
Container:
  Name: worker
  Image: <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/openwan-worker:latest
  Environment:
    - DB_HOST: <RDS_ENDPOINT>
    - REDIS_HOST: <ELASTICACHE_ENDPOINT>
    - S3_BUCKET: <BUCKET_NAME>
    - SQS_QUEUE_URL: <QUEUE_URL>
    - FFMPEG_THREADS: 2
  Secrets:
    - DB_PASSWORD: arn:aws:secretsmanager:...
  Logging:
    LogDriver: awslogs
    Options:
      awslogs-group: /ecs/openwan-worker
      awslogs-region: us-east-1
```

**Service Configuration**:
```yaml
ServiceName: openwan-worker
DesiredCount: 2
LaunchType: FARGATE
NetworkMode: awsvpc
Subnets:
  - app-subnet-a
  - app-subnet-b
  - app-subnet-c
SecurityGroups:
  - worker-sg
AutoScaling:
  MinCapacity: 2
  MaxCapacity: 10
  TargetTrackingScaling:
    - Type: SQSQueueDepth
      TargetValue: 100
      ScaleInCooldown: 300
      ScaleOutCooldown: 60
```

### Alternative: EKS (Kubernetes)

如果需要更多控制和Kubernetes生态：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openwan-backend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: openwan-backend
  template:
    metadata:
      labels:
        app: openwan-backend
    spec:
      containers:
      - name: backend
        image: <ECR_IMAGE>
        ports:
        - containerPort: 8080
        - containerPort: 9090
        resources:
          requests:
            cpu: "2000m"
            memory: "4Gi"
          limits:
            cpu: "4000m"
            memory: "8Gi"
        livenessProbe:
          httpGet:
            path: /alive
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: openwan-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: openwan-backend
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 存储方案

### S3 存储设计

#### Bucket 结构

**Primary Bucket**: `openwan-media-<ACCOUNT_ID>-<REGION>`

```
openwan-media-123456789012-us-east-1/
├── original/           # 原始文件
│   ├── video/
│   ├── audio/
│   ├── image/
│   └── document/
├── preview/            # 转码后预览文件
│   └── flv/
└── thumbnails/         # 缩略图
```

#### Bucket 配置

**Versioning**: Enabled (保留30天历史版本)

**Lifecycle Policy**:
```json
{
  "Rules": [
    {
      "Id": "TransitionOldVersions",
      "Status": "Enabled",
      "NoncurrentVersionTransitions": [
        {
          "NoncurrentDays": 30,
          "StorageClass": "GLACIER"
        }
      ],
      "NoncurrentVersionExpiration": {
        "NoncurrentDays": 90
      }
    },
    {
      "Id": "DeleteIncompleteUploads",
      "Status": "Enabled",
      "AbortIncompleteMultipartUpload": {
        "DaysAfterInitiation": 7
      }
    },
    {
      "Id": "TransitionPreviewToIA",
      "Status": "Enabled",
      "Prefix": "preview/",
      "Transitions": [
        {
          "Days": 90,
          "StorageClass": "STANDARD_IA"
        }
      ]
    }
  ]
}
```

**Encryption**: SSE-S3 (AES-256) or SSE-KMS

**CORS Configuration**:
```json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["https://openwan.example.com"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3600
    }
  ]
}
```

**Bucket Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCloudFrontAccess",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::cloudfront:user/CloudFront Origin Access Identity <OAI_ID>"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::openwan-media-*/*"
    }
  ]
}
```

### CloudFront CDN

**Distribution Configuration**:
```yaml
Origins:
  - DomainName: openwan-media-123456789012-us-east-1.s3.amazonaws.com
    Id: S3-openwan-media
    S3OriginConfig:
      OriginAccessIdentity: origin-access-identity/cloudfront/<OAI_ID>

CacheBehaviors:
  - PathPattern: /preview/*
    AllowedMethods: [GET, HEAD, OPTIONS]
    CachedMethods: [GET, HEAD]
    TargetOriginId: S3-openwan-media
    ForwardedValues:
      QueryString: false
      Headers: [Origin, Access-Control-Request-Method, Access-Control-Request-Headers]
    ViewerProtocolPolicy: redirect-to-https
    MinTTL: 0
    DefaultTTL: 86400  # 1 day
    MaxTTL: 31536000   # 1 year
    Compress: true

  - PathPattern: /thumbnails/*
    AllowedMethods: [GET, HEAD, OPTIONS]
    TargetOriginId: S3-openwan-media
    ViewerProtocolPolicy: redirect-to-https
    DefaultTTL: 604800  # 7 days
    Compress: true

PriceClass: PriceClass_All
ViewerCertificate:
  AcmCertificateArn: arn:aws:acm:us-east-1:123456789012:certificate/<CERT_ID>
  SslSupportMethod: sni-only
  MinimumProtocolVersion: TLSv1.2_2021

WebACLId: arn:aws:wafv2:us-east-1:123456789012:global/webacl/<WAF_ID>
```

---

## 数据库方案

### RDS MySQL Multi-AZ

**Instance Specification**:
```yaml
DBInstanceClass: db.r6g.xlarge
  vCPU: 4
  RAM: 32 GB
  Network: Up to 10 Gbps

Engine: mysql
EngineVersion: 8.0.35

Storage:
  Type: gp3 (SSD)
  Allocated: 500 GB
  IOPS: 12000
  Throughput: 500 MB/s
  AutoScaling:
    Enabled: true
    MaxAllocatedStorage: 2000 GB

MultiAZ: true  # 自动故障转移

BackupRetentionPeriod: 7 days
PreferredBackupWindow: "03:00-04:00"
PreferredMaintenanceWindow: "mon:04:00-mon:05:00"

Encryption:
  Enabled: true
  KmsKeyId: arn:aws:kms:us-east-1:123456789012:key/<KEY_ID>

MonitoringInterval: 60  # Enhanced Monitoring
EnablePerformanceInsights: true

DBSubnetGroup:
  Name: openwan-db-subnet-group
  Subnets:
    - data-subnet-a
    - data-subnet-b
    - data-subnet-c

VPCSecurityGroups:
  - rds-sg
```

**Read Replicas** (可选，用于读写分离):
```yaml
ReadReplicas:
  - DBInstanceIdentifier: openwan-db-replica-1
    SourceDBInstanceIdentifier: openwan-db-primary
    AvailabilityZone: us-east-1b
    DBInstanceClass: db.r6g.large
  
  - DBInstanceIdentifier: openwan-db-replica-2
    SourceDBInstanceIdentifier: openwan-db-primary
    AvailabilityZone: us-east-1c
    DBInstanceClass: db.r6g.large
```

**Parameter Group**:
```ini
[mysqld]
# Connection
max_connections = 500
max_connect_errors = 1000
connect_timeout = 10

# Performance
innodb_buffer_pool_size = 24G  # 75% of RAM
innodb_log_file_size = 1G
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# Query Cache
query_cache_type = 1
query_cache_size = 256M

# Character Set
character_set_server = utf8mb4
collation_server = utf8mb4_unicode_ci

# Logging
slow_query_log = 1
slow_query_log_file = /rdsdbdata/log/mysql-slow.log
long_query_time = 2
log_queries_not_using_indexes = 1

# Replication
binlog_format = ROW
binlog_row_image = MINIMAL
```

---

## 缓存和队列

### ElastiCache Redis Cluster

**Cluster Configuration**:
```yaml
CacheClusterName: openwan-redis
Engine: redis
EngineVersion: 7.0

NodeType: cache.r6g.large
  vCPU: 2
  RAM: 13.07 GB
  Network: Up to 10 Gbps

ClusterMode: enabled  # 集群模式
NumNodeGroups: 3  # 3个分片
ReplicasPerNodeGroup: 2  # 每个分片2个副本

# 共3个主节点 + 6个副本节点 = 9个节点

MultiAZ: true
AutomaticFailover: true

CacheSubnetGroup:
  Name: openwan-redis-subnet-group
  Subnets:
    - data-subnet-a
    - data-subnet-b
    - data-subnet-c

SecurityGroups:
  - elasticache-sg

SnapshotRetentionLimit: 5
SnapshotWindow: "03:00-05:00"
PreferredMaintenanceWindow: "sun:05:00-sun:07:00"

AtRestEncryption: true
TransitEncryption: true
AuthToken: <STORED_IN_SECRETS_MANAGER>

NotificationTopicArn: arn:aws:sns:us-east-1:123456789012:openwan-redis-events
```

**Parameter Group**:
```ini
# Memory Management
maxmemory-policy allkeys-lru
maxmemory-samples 5

# Persistence (RDB)
save 900 1
save 300 10
save 60 10000

# AOF (for durability)
appendonly yes
appendfsync everysec

# Performance
tcp-keepalive 300
timeout 300

# Cluster
cluster-enabled yes
cluster-node-timeout 15000
```

### Amazon SQS

**Queue Configuration**:

**1. Transcoding Queue** (主队列):
```yaml
QueueName: openwan-transcoding-queue
DelaySeconds: 0
MessageRetentionPeriod: 1209600  # 14 days
VisibilityTimeout: 3600  # 1 hour (transcoding time)
ReceiveMessageWaitTimeSeconds: 20  # Long polling
MaximumMessageSize: 262144  # 256 KB

DeadLetterQueue:
  TargetArn: arn:aws:sqs:us-east-1:123456789012:openwan-transcoding-dlq
  MaxReceiveCount: 3

Encryption:
  KmsMasterKeyId: alias/aws/sqs
  KmsDataKeyReusePeriodSeconds: 300

Tags:
  Environment: production
  Application: openwan
```

**2. Dead Letter Queue** (DLQ):
```yaml
QueueName: openwan-transcoding-dlq
MessageRetentionPeriod: 1209600  # 14 days
```

**Queue Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowECSTasksToSendReceive",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::123456789012:role/openwan-ecs-task-role"
      },
      "Action": [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      "Resource": "arn:aws:sqs:us-east-1:123456789012:openwan-transcoding-queue"
    }
  ]
}
```

---

## 安全架构

### AWS WAF

**Web ACL Configuration**:
```yaml
Name: openwan-waf-webacl
Scope: CLOUDFRONT
DefaultAction: Allow

Rules:
  # AWS Managed Rules
  - Name: AWSManagedRulesCommonRuleSet
    Priority: 1
    ManagedRuleGroup:
      VendorName: AWS
      Name: AWSManagedRulesCommonRuleSet
    OverrideAction: None

  - Name: AWSManagedRulesKnownBadInputsRuleSet
    Priority: 2
    ManagedRuleGroup:
      VendorName: AWS
      Name: AWSManagedRulesKnownBadInputsRuleSet

  - Name: AWSManagedRulesSQLiRuleSet
    Priority: 3
    ManagedRuleGroup:
      VendorName: AWS
      Name: AWSManagedRulesSQLiRuleSet

  # Custom Rules
  - Name: RateLimitRule
    Priority: 10
    Statement:
      RateBasedStatement:
        Limit: 2000
        AggregateKeyType: IP
    Action: Block

  - Name: GeoBlockRule
    Priority: 11
    Statement:
      NotStatement:
        Statement:
          GeoMatchStatement:
            CountryCodes: [US, CN, JP, KR, SG]  # 允许的国家
    Action: Block

  - Name: IPReputationList
    Priority: 12
    Statement:
      ManagedRuleGroupStatement:
        VendorName: AWS
        Name: AWSManagedRulesAmazonIpReputationList
```

### IAM Roles

#### ECS Task Role (Backend)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::openwan-media-*",
        "arn:aws:s3:::openwan-media-*/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      "Resource": "arn:aws:sqs:*:*:openwan-*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:openwan/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments",
        "xray:PutTelemetryRecords"
      ],
      "Resource": "*"
    }
  ]
}
```

#### ECS Task Execution Role

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:*:*:secret:openwan/*"
    }
  ]
}
```

### Secrets Manager

**Secrets 存储**:

```bash
# Database Password
aws secretsmanager create-secret \
  --name openwan/database/password \
  --secret-string '{"password":"<SECURE_PASSWORD>"}'

# JWT Secret
aws secretsmanager create-secret \
  --name openwan/jwt/secret \
  --secret-string '{"secret":"<RANDOM_JWT_SECRET>"}'

# Redis Auth Token
aws secretsmanager create-secret \
  --name openwan/redis/auth-token \
  --secret-string '{"token":"<RANDOM_TOKEN>"}'

# S3 Encryption Key
aws secretsmanager create-secret \
  --name openwan/s3/encryption-key \
  --secret-string '{"key":"<ENCRYPTION_KEY>"}'
```

### KMS Encryption Keys

```yaml
# RDS Encryption Key
KeyAlias: alias/openwan-rds
Description: Encryption key for OpenWan RDS database
KeyPolicy:
  - Principal: rds.amazonaws.com
    Action: 
      - kms:Decrypt
      - kms:DescribeKey

# S3 Encryption Key
KeyAlias: alias/openwan-s3
Description: Encryption key for OpenWan S3 bucket
KeyPolicy:
  - Principal: s3.amazonaws.com

# ElastiCache Encryption Key
KeyAlias: alias/openwan-elasticache
Description: Encryption key for OpenWan ElastiCache
```

---

## 可观测性

### CloudWatch Logs

**Log Groups**:
```yaml
LogGroups:
  # Application Logs
  - Name: /ecs/openwan-backend
    RetentionInDays: 30
    KmsKeyId: arn:aws:kms:us-east-1:123456789012:key/<KEY_ID>

  - Name: /ecs/openwan-worker
    RetentionInDays: 30

  # Infrastructure Logs
  - Name: /aws/rds/instance/openwan-db/error
    RetentionInDays: 30

  - Name: /aws/elasticache/openwan-redis
    RetentionInDays: 30

  # Access Logs
  - Name: /aws/alb/openwan
    RetentionInDays: 90

  - Name: /aws/waf/openwan
    RetentionInDays: 90

  # CloudTrail Logs
  - Name: /aws/cloudtrail/openwan
    RetentionInDays: 365
```

**Log Insights Queries**:

```sql
-- Error Rate
fields @timestamp, @message
| filter @message like /ERROR/
| stats count() as error_count by bin(5m)

-- Slow Queries
fields @timestamp, duration, query
| filter operation = "database.query" and duration > 1000
| sort duration desc
| limit 20

-- API Response Times
fields @timestamp, api_endpoint, response_time
| stats avg(response_time) as avg_response, max(response_time) as max_response by api_endpoint
| sort avg_response desc
```

### CloudWatch Metrics

**Custom Metrics**:
```go
// In application code
import "github.com/aws/aws-sdk-go/service/cloudwatch"

// Publish custom metrics
cloudwatch.PutMetricData(&cloudwatch.PutMetricDataInput{
    Namespace: aws.String("OpenWan/Application"),
    MetricData: []*cloudwatch.MetricDatum{
        {
            MetricName: aws.String("FileUploadCount"),
            Value:      aws.Float64(1),
            Unit:       aws.String("Count"),
            Dimensions: []*cloudwatch.Dimension{
                {
                    Name:  aws.String("FileType"),
                    Value: aws.String("video"),
                },
            },
        },
    },
})
```

**Metric Alarms**:
```yaml
Alarms:
  # High CPU
  - AlarmName: openwan-backend-high-cpu
    MetricName: CPUUtilization
    Namespace: AWS/ECS
    Statistic: Average
    Period: 300
    EvaluationPeriods: 2
    Threshold: 80
    ComparisonOperator: GreaterThanThreshold
    TreatMissingData: notBreaching
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-alerts

  # High Memory
  - AlarmName: openwan-backend-high-memory
    MetricName: MemoryUtilization
    Namespace: AWS/ECS
    Threshold: 85
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-alerts

  # RDS Connection Count
  - AlarmName: openwan-db-connection-high
    MetricName: DatabaseConnections
    Namespace: AWS/RDS
    Threshold: 450  # 90% of max_connections
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-critical

  # ElastiCache Memory
  - AlarmName: openwan-redis-memory-high
    MetricName: DatabaseMemoryUsagePercentage
    Namespace: AWS/ElastiCache
    Threshold: 90
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-alerts

  # SQS Queue Depth
  - AlarmName: openwan-sqs-queue-deep
    MetricName: ApproximateNumberOfMessagesVisible
    Namespace: AWS/SQS
    Threshold: 1000
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-alerts

  # ALB 5xx Errors
  - AlarmName: openwan-alb-5xx-errors
    MetricName: HTTPCode_Target_5XX_Count
    Namespace: AWS/ApplicationELB
    Statistic: Sum
    Period: 60
    Threshold: 10
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-critical

  # WAF Blocked Requests
  - AlarmName: openwan-waf-blocked-high
    MetricName: BlockedRequests
    Namespace: AWS/WAFV2
    Threshold: 100
    Period: 300
    TreatMissingData: notBreaching
    AlarmActions:
      - arn:aws:sns:us-east-1:123456789012:openwan-security
```

### AWS X-Ray

**Instrumentation**:
```go
// In Go application
import (
    "github.com/aws/aws-xray-sdk-go/xray"
)

// Instrument HTTP handlers
http.Handle("/", xray.Handler(xray.NewFixedSegmentNamer("openwan-backend"), mux))

// Instrument AWS SDK clients
sess := session.Must(session.NewSession())
xray.AWS(sess.Client())

// Custom subsegments
ctx, seg := xray.BeginSubsegment(ctx, "database.query")
defer seg.Close(nil)
// ... database operation ...
```

**Sampling Rules**:
```json
{
  "version": 2,
  "rules": [
    {
      "description": "Sample all errors",
      "host": "*",
      "http_method": "*",
      "url_path": "*",
      "fixed_target": 1,
      "rate": 1.0,
      "attributes": {
        "error": "true"
      }
    },
    {
      "description": "Sample health checks at low rate",
      "host": "*",
      "http_method": "GET",
      "url_path": "/health",
      "fixed_target": 0,
      "rate": 0.01
    },
    {
      "description": "Sample 10% of other requests",
      "host": "*",
      "http_method": "*",
      "url_path": "*",
      "fixed_target": 1,
      "rate": 0.1
    }
  ],
  "default": {
    "fixed_target": 1,
    "rate": 0.05
  }
}
```

### CloudWatch Dashboard

**Dashboard JSON**:
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ECS", "CPUUtilization", {"stat": "Average"}],
          [".", "MemoryUtilization"]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "ECS Resource Utilization",
        "yAxis": {
          "left": {
            "min": 0,
            "max": 100
          }
        }
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/RDS", "CPUUtilization"],
          [".", "DatabaseConnections"],
          [".", "ReadLatency"],
          [".", "WriteLatency"]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "RDS Performance"
      }
    },
    {
      "type": "log",
      "properties": {
        "query": "SOURCE '/ecs/openwan-backend' | fields @timestamp, @message | filter @message like /ERROR/ | sort @timestamp desc | limit 20",
        "region": "us-east-1",
        "title": "Recent Errors"
      }
    }
  ]
}
```

---

## 成本估算

### 月度成本估算 (us-east-1)

| 服务 | 配置 | 月费用 (USD) |
|------|------|-------------|
| **Compute (ECS Fargate)** | | |
| Backend (2-10 tasks) | 2vCPU, 4GB x 2 (baseline) | ~$73 x 2 = $146 |
| Worker (2-10 tasks) | 4vCPU, 8GB x 2 (baseline) | ~$146 x 2 = $292 |
| Auto Scaling (average +50%) | | +$219 |
| **Subtotal Compute** | | **$657** |
| **Database (RDS)** | | |
| Primary (Multi-AZ) | db.r6g.xlarge | $585 |
| Storage (500GB gp3) | 500GB + IOPS | $115 |
| Backup Storage (average 1TB) | | $95 |
| **Subtotal Database** | | **$795** |
| **Cache (ElastiCache)** | | |
| Redis Cluster (9 nodes) | cache.r6g.large x 9 | $0.226/hr x 9 x 730 = $1,486 |
| **Subtotal Cache** | | **$1,486** |
| **Storage (S3 + CloudFront)** | | |
| S3 Standard (10TB) | $0.023/GB | $235 |
| S3 Requests (10M requests) | | $5 |
| CloudFront (1TB transfer) | $0.085/GB | $85 |
| **Subtotal Storage** | | **$325** |
| **Networking** | | |
| ALB (730 hrs + LCU) | | $23 + $22 = $45 |
| NAT Gateway (3 AZs) | $0.045/hr x 3 x 730 | $98 |
| Data Transfer (500GB out) | $0.09/GB | $45 |
| **Subtotal Networking** | | **$188** |
| **Queue (SQS)** | | |
| Standard Queue (50M requests) | $0.40/M | $20 |
| **Security & Monitoring** | | |
| WAF (10M requests) | | $6 |
| CloudWatch Logs (100GB) | | $50 |
| CloudWatch Metrics (custom) | | $15 |
| X-Ray (1M traces) | | $5 |
| Secrets Manager (10 secrets) | | $4 |
| GuardDuty | | $30 |
| **Subtotal Security/Monitoring** | | **$110** |
| **Other Services** | | |
| Route 53 (1 hosted zone) | | $0.50 |
| ACM (SSL certificate) | | FREE |
| ECR (100GB) | | $10 |
| **Subtotal Other** | | **$10.50** |
| | | |
| **TOTAL MONTHLY COST** | | **~$3,591.50** |

### 成本优化建议

1. **使用 Savings Plans**: 可节省 20-30%
2. **Reserved Instances**: RDS/ElastiCache 可节省 40-60%
3. **S3 Intelligent-Tiering**: 自动优化存储成本
4. **Spot Instances**: Worker tasks 使用 Spot 可节省 70%
5. **Auto Scaling**: 在低峰期自动缩容

**优化后预估**: ~$2,500/month

---

## 部署检查清单

### 前置准备

- [ ] AWS 账号创建
- [ ] IAM 用户配置（具有管理员权限）
- [ ] AWS CLI 安装和配置
- [ ] Terraform/CloudFormation 工具安装
- [ ] Docker 安装（构建镜像）
- [ ] 域名购买和 Route 53 托管
- [ ] SSL 证书申请（ACM）

### 部署步骤

1. [ ] 创建 VPC 和子网
2. [ ] 创建 Security Groups
3. [ ] 创建 RDS 数据库
4. [ ] 创建 ElastiCache Redis
5. [ ] 创建 S3 存储桶
6. [ ] 创建 SQS 队列
7. [ ] 创建 ECR 仓库
8. [ ] 构建并推送 Docker 镜像
9. [ ] 创建 Secrets Manager 密钥
10. [ ] 创建 IAM 角色
11. [ ] 创建 ECS 集群
12. [ ] 部署 ECS 任务和服务
13. [ ] 创建 ALB 和目标组
14. [ ] 配置 CloudFront
15. [ ] 配置 WAF
16. [ ] 配置 Route 53
17. [ ] 设置 CloudWatch 告警
18. [ ] 配置 X-Ray
19. [ ] 运行数据库迁移
20. [ ] 验证部署

---

**下一步**: 查看详细的部署脚本和Terraform代码
