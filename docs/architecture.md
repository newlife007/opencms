# OpenWan Architecture Documentation

## System Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           End Users                                  │
└────────────┬────────────────────────────────────────────────────────┘
             │
             │ HTTPS
             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Application Load Balancer (ALB)                   │
│                  - SSL Termination                                   │
│                  - Health Checks (/health, /ready)                   │
│                  - Connection Draining (300s)                        │
└──────┬──────────────────────────────────────────────────────────────┘
       │
       │ HTTP (Internal)
       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Frontend Layer (Vue.js)                           │
│  ┌─────────────────────────────────────────────────────────┐        │
│  │   Static Files served via Nginx or CloudFront CDN       │        │
│  │   - SPA Application (Vue 3 + Element Plus)              │        │
│  │   - Video.js Player with FLV.js                         │        │
│  │   - Responsive UI with i18n (Chinese/English)           │        │
│  └─────────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────────┘
       │
       │ REST API (JSON)
       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Backend API Layer (Go + Gin)                      │
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  API Instance │  │  API Instance │  │  API Instance │             │
│  │   (Stateless) │  │   (Stateless) │  │   (Stateless) │             │
│  │               │  │               │  │               │             │
│  │  - Auth       │  │  - Auth       │  │  - Auth       │             │
│  │  - Files      │  │  - Files      │  │  - Files      │             │
│  │  - Admin      │  │  - Admin      │  │  - Admin      │             │
│  │  - Search     │  │  - Search     │  │  - Search     │             │
│  │  - Catalog    │  │  - Catalog    │  │  - Catalog    │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │
│         │                 │                 │                        │
│         └─────────────────┴─────────────────┘                        │
│                           │                                          │
│                     Auto-Scaling                                     │
│                  Min: 2, Max: 20 replicas                            │
│                  Scale on CPU 70% / Memory 80%                       │
└───────────┬──────────────────────────────────────────────────────────┘
            │
            │
            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Service Layer Components                          │
│                                                                       │
│  ┌───────────────┐  ┌────────────────┐  ┌─────────────────┐        │
│  │   ACL Service │  │  File Service   │  │ Catalog Service │        │
│  │   - RBAC      │  │  - Upload       │  │ - Metadata      │        │
│  │   - Permissions│  │  - Download    │  │ - Dynamic Forms │        │
│  └───────────────┘  └────────────────┘  └─────────────────┘        │
│                                                                       │
│  ┌───────────────┐  ┌────────────────┐  ┌─────────────────┐        │
│  │Search Service │  │Transcoding Svc │  │  Storage Service│        │
│  │ - Sphinx      │  │  - Queue Jobs  │  │  - Local/S3     │        │
│  │ - Fallback DB │  │  - FFmpeg      │  │  - MD5 paths    │        │
│  └───────────────┘  └────────────────┘  └─────────────────┘        │
└───────────┬──────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Data & Infrastructure Layer                      │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │              Redis Cluster (Session + Cache)              │       │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐           │       │
│  │  │  Master  │────│  Slave 1 │    │  Slave 2 │           │       │
│  │  └──────────┘    └──────────┘    └──────────┘           │       │
│  │        │                                                  │       │
│  │        │  Sentinel for Auto-Failover                     │       │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐           │       │
│  │  │Sentinel 1│────│Sentinel 2│────│Sentinel 3│           │       │
│  │  └──────────┘    └──────────┘    └──────────┘           │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │              MySQL Database (RDS Multi-AZ)                │       │
│  │  ┌──────────┐                     ┌──────────┐           │       │
│  │  │  Master  │─────Replication────▶│  Slave   │           │       │
│  │  │ (Write)  │                     │  (Read)  │           │       │
│  │  └──────────┘                     └──────────┘           │       │
│  │       │                                  │                │       │
│  │       │   Auto-Failover (< 2 min)       │                │       │
│  │       │                                  │                │       │
│  │  Connection Pooling (Max 100 per instance)               │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │             RabbitMQ Cluster (Message Queue)              │       │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐           │       │
│  │  │  Node 1  │────│  Node 2  │────│  Node 3  │           │       │
│  │  └──────────┘    └──────────┘    └──────────┘           │       │
│  │                                                           │       │
│  │  Queues: transcoding_jobs, notifications, search_index   │       │
│  │  DLQ: dead_letter_queue (failed jobs)                    │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │                  AWS S3 (Media Storage)                   │       │
│  │  Bucket: openwan-media                                    │       │
│  │  - Versioning Enabled                                     │       │
│  │  - Server-Side Encryption                                 │       │
│  │  - Lifecycle Policies (Glacier after 90 days)            │       │
│  │  - Cross-Region Replication (DR)                         │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │              Sphinx Search Engine (CSFT)                  │       │
│  │  Main Index + Delta Index                                 │       │
│  │  Chinese Word Segmentation Support                        │       │
│  └──────────────────────────────────────────────────────────┘       │
└───────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────┐
│                     Transcoding Workers Layer                        │
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │   Worker 1   │  │   Worker 2   │  │   Worker N   │              │
│  │   FFmpeg     │  │   FFmpeg     │  │   FFmpeg     │              │
│  │   Consumer   │  │   Consumer   │  │   Consumer   │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
│                                                                       │
│  Auto-Scaling: Min 2, Max 10 based on Queue Depth                   │
│  Each worker processes jobs from RabbitMQ                            │
└───────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────┐
│                   Monitoring & Observability                         │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │                    Prometheus                             │       │
│  │  - Metrics Collection (/metrics endpoint)                 │       │
│  │  - Alert Manager                                          │       │
│  └──────────────────────────────────────────────────────────┘       │
│                          │                                           │
│                          ▼                                           │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │                      Grafana                              │       │
│  │  - Dashboards (System, Application, Business)            │       │
│  │  - Visualization                                          │       │
│  └──────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │              Centralized Logging                          │       │
│  │  - CloudWatch Logs / ELK Stack                            │       │
│  │  - Structured JSON logs with correlation IDs             │       │
│  │  - Log aggregation and search                            │       │
│  └──────────────────────────────────────────────────────────┘       │
└───────────────────────────────────────────────────────────────────────┘
```

## Component Details

### Frontend Layer (Vue.js)
- **Technology**: Vue 3, Element Plus UI, Video.js with FLV.js
- **Deployment**: Static files served via Nginx or CloudFront CDN
- **Features**:
  - Single Page Application (SPA)
  - Responsive design (Desktop 1920x1080 primary, tablet, mobile)
  - i18n support (Chinese, English)
  - Video player with FLV format support
  - File upload with drag-drop and progress tracking
  - Dynamic form builder for catalog metadata
  - RBAC-based UI rendering

### Backend API Layer (Go + Gin)
- **Technology**: Go 1.25.5, Gin web framework
- **Architecture**: Stateless microservices
- **Horizontal Scaling**: 
  - Minimum 2 replicas for HA
  - Maximum 20 replicas for peak load
  - Auto-scaling based on CPU (70%) and Memory (80%)
- **API Endpoints**:
  - `/api/v1/auth/*` - Authentication and session management
  - `/api/v1/files/*` - File upload, download, catalog, search
  - `/api/v1/admin/*` - Admin panel (users, groups, roles, permissions)
  - `/api/v1/categories/*` - Category management
  - `/api/v1/catalog/*` - Catalog metadata configuration
  - `/api/v1/search` - Search functionality
  - `/health` - Health check (200/503)
  - `/ready` - Readiness probe (200/503)
  - `/alive` - Liveness probe (200)
  - `/metrics` - Prometheus metrics

### Service Layer
- **ACL Service**: RBAC implementation with permission checks
- **File Service**: Upload, download, metadata management
- **Catalog Service**: Dynamic metadata field configuration
- **Search Service**: Sphinx integration with DB fallback
- **Transcoding Service**: FFmpeg job queue management
- **Storage Service**: Abstraction for local/S3 storage

### Data Layer

#### Redis Cluster (Session + Cache)
- **Purpose**: Distributed session management and caching
- **Architecture**: Master-Slave with Sentinel
- **Failover**: Automatic within 30 seconds via Sentinel
- **Use Cases**:
  - Session storage (externalized from app servers)
  - Cache for permissions, categories, catalog config
  - Cache hit rate target: 80%+

#### MySQL Database (RDS Multi-AZ)
- **Architecture**: Master-Slave replication
- **Failover**: Automatic within 2 minutes
- **Read-Write Splitting**: Writes to master, reads from slaves
- **Connection Pooling**: Max 100 connections per instance
- **Tables**: 
  - Core: ow_files, ow_users, ow_groups, ow_roles, ow_permissions
  - Relationships: ow_groups_has_roles, ow_roles_has_permissions
  - Metadata: ow_catalog, ow_categories, ow_levels

#### RabbitMQ Cluster (Message Queue)
- **Architecture**: 3-node cluster for HA
- **Queues**:
  - `transcoding_jobs` - FFmpeg transcoding tasks
  - `notifications` - User notifications
  - `search_indexing` - Sphinx delta indexing
- **DLQ**: Dead letter queue for failed jobs after max retries
- **Delivery**: At-least-once with idempotent processing

#### AWS S3 (Media Storage)
- **Bucket**: openwan-media
- **Features**:
  - Versioning enabled for data protection
  - Server-side encryption (SSE-S3)
  - Lifecycle policies (transition to Glacier after 90 days)
  - Cross-region replication for disaster recovery
- **File Organization**: MD5-based directory structure
  - `/data1/{md5(filename+time)}/{md5(content)}.{ext}`
  - Max 65,535 files per data directory

#### Sphinx Search Engine
- **Indexes**: Main index + Delta index
- **Features**:
  - Chinese word segmentation (CoreSeek)
  - Full-text search on file metadata
  - Access control filtering in queries
- **Indexing Strategy**:
  - Main index: Full data, rebuilt daily
  - Delta index: Recent changes (24 hours), rebuilt hourly

### Transcoding Workers Layer
- **Technology**: Go workers consuming from RabbitMQ
- **FFmpeg**: Media file transcoding to FLV format
- **Scaling**: 
  - Min 2, Max 10 workers
  - Scale based on queue depth (>100 messages triggers scale-up)
- **Job Processing**:
  - Download from S3 (if S3 storage mode)
  - Transcode with FFmpeg
  - Upload preview to S3
  - Update database with job status

### Monitoring & Observability

#### Prometheus
- **Metrics**:
  - HTTP request count, duration, status codes
  - Database query count, duration
  - Cache hit/miss rates
  - File upload/download counts
  - Transcoding job metrics
  - Queue depth and message rates
- **Alerting**: Alert Manager for anomaly detection

#### Grafana
- **Dashboards**:
  - System Health: CPU, memory, disk, network
  - Application Performance: Request latency, error rates
  - Business Metrics: Uploads, searches, active users

#### Centralized Logging
- **Technology**: CloudWatch Logs or ELK Stack
- **Format**: Structured JSON with correlation IDs
- **Fields**: timestamp, level, message, service, request_id, user_id

## Data Flow Diagrams

### File Upload Flow
```
User → Frontend → Backend API → Storage Service → S3/Local
                       │
                       ├─→ Database (file record)
                       │
                       └─→ RabbitMQ (transcoding job)
                                │
                                ▼
                          Transcoding Worker
                                │
                                ├─→ Download from S3
                                ├─→ FFmpeg transcode
                                ├─→ Upload preview to S3
                                └─→ Update database
```

### Authentication Flow
```
User → Frontend → POST /api/v1/auth/login → Backend API
                                                  │
                                                  ├─→ Database (verify credentials)
                                                  │
                                                  ├─→ Redis (create session)
                                                  │
                                                  └─→ Return token/session ID
                                                       │
                                                       ▼
                                              Frontend stores token
                                              (localStorage)
```

### Search Flow
```
User → Frontend → POST /api/v1/search → Backend API
                                             │
                                             ├─→ Sphinx (full-text search)
                                             │        │
                                             │        └─→ Fallback to MySQL if Sphinx unavailable
                                             │
                                             ├─→ ACL check (filter by user permissions)
                                             │
                                             └─→ Return filtered results
```

## High Availability Features

1. **No Single Points of Failure**:
   - Multiple API instances behind load balancer
   - Database replication with automatic failover
   - Redis Sentinel for cache HA
   - RabbitMQ cluster for queue HA
   - S3 for durable storage

2. **Auto-Scaling**:
   - API servers: Scale 2-20 based on load
   - Workers: Scale 2-10 based on queue depth
   - Graceful shutdown with connection draining

3. **Health Checks**:
   - `/health` endpoint checks all dependencies
   - Load balancer removes unhealthy instances within 60s
   - Readiness probes prevent traffic to initializing instances

4. **Failover Times**:
   - API instance failure: <60 seconds
   - Database failover: <2 minutes
   - Redis failover: <30 seconds
   - No downtime for rolling deployments

## Scalability Features

1. **Stateless Design**:
   - All session data in Redis
   - No local state in application servers
   - Any instance can serve any request

2. **Horizontal Scaling**:
   - Add more API instances for more capacity
   - Add more workers for more transcoding throughput
   - Database read replicas for read scaling

3. **Caching Strategy**:
   - Redis cache for frequently accessed data
   - 80%+ cache hit rate reduces database load
   - Cache invalidation on data updates

4. **CDN Integration**:
   - CloudFront for media file delivery
   - Reduced bandwidth costs
   - Improved global performance

## Security Architecture

1. **Authentication**:
   - Session-based or JWT token authentication
   - Secure password hashing (bcrypt)
   - Session timeout and renewal

2. **Authorization**:
   - Role-Based Access Control (RBAC)
   - Permission checks at service layer
   - Group-based category access restrictions

3. **Network Security**:
   - TLS termination at load balancer
   - Private subnets for databases and caches
   - Security groups restricting access

4. **Data Security**:
   - Encryption at rest (S3 SSE, RDS encryption)
   - Encryption in transit (TLS)
   - No secrets in code or logs

5. **Rate Limiting**:
   - Per-IP rate limits (100 rps)
   - Per-user rate limits (1000 rps)
   - Anonymous rate limits (10 rps)

## Deployment Architecture

### Kubernetes Deployment
```
┌─────────────────────────────────────────────┐
│           Kubernetes Cluster                 │
│                                              │
│  ┌────────────────────────────────────┐     │
│  │     API Deployment                  │     │
│  │  replicas: 2-20 (HPA)              │     │
│  │  resources:                         │     │
│  │    requests: 500m CPU, 512Mi RAM   │     │
│  │    limits: 2 CPU, 2Gi RAM          │     │
│  └────────────────────────────────────┘     │
│                                              │
│  ┌────────────────────────────────────┐     │
│  │     Worker Deployment               │     │
│  │  replicas: 2-10 (HPA)              │     │
│  │  resources:                         │     │
│  │    requests: 1 CPU, 1Gi RAM        │     │
│  │    limits: 4 CPU, 4Gi RAM          │     │
│  └────────────────────────────────────┘     │
│                                              │
│  ┌────────────────────────────────────┐     │
│  │     Services                        │     │
│  │  - api-service (ClusterIP)         │     │
│  │  - metrics-service (ClusterIP)     │     │
│  └────────────────────────────────────┘     │
│                                              │
│  ┌────────────────────────────────────┐     │
│  │     Ingress                         │     │
│  │  - TLS termination                  │     │
│  │  - Path routing                     │     │
│  └────────────────────────────────────┘     │
└─────────────────────────────────────────────┘
```

### AWS Infrastructure
- **VPC**: 3 availability zones
- **Subnets**: Public (load balancer), Private (app, workers), Data (database, cache)
- **RDS**: Multi-AZ MySQL 5.7+
- **ElastiCache**: Redis cluster mode
- **S3**: Media storage bucket with CloudFront distribution
- **ALB**: Application Load Balancer with SSL certificate
- **ECS/EKS**: Container orchestration

## Performance Characteristics

### Target SLAs
- **Availability**: 99.9% (max 43 minutes downtime/month)
- **API Response Time**: 
  - p50 < 100ms (read operations)
  - p95 < 500ms (read operations)
  - p99 < 1000ms (read operations)
- **Search**: < 200ms response time
- **File Upload**: 50MB file in < 30 seconds (good network)
- **Transcoding**: Start within 10 seconds, complete within 5 minutes (typical video)
- **Concurrent Users**: 500+ simultaneous users
- **Upload Concurrency**: 100+ concurrent uploads
- **Transcoding Concurrency**: 10+ concurrent jobs

### Capacity Planning
- **API Instances**: 1 instance per 100 concurrent users
- **Workers**: 1 worker per 2 concurrent transcoding jobs
- **Database**: 100 connections per instance, 1000 total
- **Redis**: 10k ops/sec per node
- **S3**: Unlimited storage, 3500 PUT/s, 5500 GET/s per prefix

## Disaster Recovery

### Backup Strategy
- **Database**: Automated daily backups, 7-day retention, point-in-time recovery
- **Redis**: RDB snapshots every 6 hours, AOF for durability
- **S3**: Versioning enabled, cross-region replication

### Recovery Procedures
- **RTO** (Recovery Time Objective): 1 hour
- **RPO** (Recovery Point Objective): 15 minutes

See [Disaster Recovery Runbook](./dr-runbook.md) for detailed procedures.

## Migration from Legacy PHP System

### Architecture Comparison
| Component | Legacy (PHP) | New (Go/Vue) |
|-----------|--------------|--------------|
| Frontend | PHP templates | Vue.js SPA |
| Backend | Monolithic PHP | Stateless Go microservices |
| Database | Single MySQL | MySQL with replication |
| Session | File-based | Redis distributed |
| Caching | None/File | Redis cluster |
| Storage | Local filesystem | S3 + Local |
| Scaling | Vertical only | Horizontal auto-scaling |
| HA | Single server | Multi-AZ, load balanced |

### Key Improvements
1. **Scalability**: Horizontal scaling vs vertical-only
2. **Availability**: 99.9% with redundancy vs single server
3. **Performance**: Sub-second response vs multi-second
4. **Maintainability**: Microservices vs monolith
5. **Modern Tech**: Go/Vue vs PHP5.x
6. **Cloud-Ready**: S3, RDS, ElastiCache integration

---

**Document Version**: 1.0  
**Last Updated**: 2024-02-01  
**Maintained By**: OpenWan Development Team
