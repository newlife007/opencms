# Disaster Recovery Runbook

## Overview

This runbook provides step-by-step procedures for recovering the OpenWan media asset management system from various disaster scenarios.

**Recovery Time Objective (RTO)**: 1 hour  
**Recovery Point Objective (RPO)**: 15 minutes

## Emergency Contacts

| Role | Contact | Phone | Email |
|------|---------|-------|-------|
| On-Call Engineer | | | |
| Database Administrator | | | |
| DevOps Lead | | | |
| System Administrator | | | |
| Product Owner | | | |

## Pre-Requisites

### Access Requirements
- AWS Console access with admin privileges
- SSH access to EC2 instances
- Database credentials (stored in AWS Secrets Manager)
- S3 bucket access
- kubectl access to Kubernetes cluster (if applicable)

### Tools Required
- AWS CLI configured
- kubectl (for Kubernetes deployments)
- MySQL client
- Redis CLI
- ssh client

### Backup Locations
- **Database Backups**: RDS automated backups (7-day retention) or S3 bucket `s3://openwan-backups/mysql/`
- **Redis Snapshots**: S3 bucket `s3://openwan-backups/redis/`
- **Configuration Files**: S3 bucket `s3://openwan-backups/configs/`
- **Application Code**: Git repository (main branch)

## Disaster Scenarios

### Table of Contents
1. [Complete System Failure](#1-complete-system-failure)
2. [Database Failure](#2-database-failure)
3. [Redis Cache Failure](#3-redis-cache-failure)
4. [Application Server Failure](#4-application-server-failure)
5. [S3 Storage Failure](#5-s3-storage-failure)
6. [RabbitMQ Message Queue Failure](#6-rabbitmq-message-queue-failure)
7. [Network/Load Balancer Failure](#7-networkload-balancer-failure)
8. [Data Corruption or Accidental Deletion](#8-data-corruption-or-accidental-deletion)
9. [Security Breach](#9-security-breach)
10. [Regional Outage](#10-regional-outage)

---

## 1. Complete System Failure

### Symptoms
- All services unavailable
- Load balancer health checks failing
- No response from any component

### Impact
- Complete service outage
- Users cannot access system

### Recovery Procedure

#### Step 1: Assess Damage (5 minutes)
```bash
# Check AWS service status
aws health describe-events --filter eventTypeCategories=issue

# Check EC2 instances
aws ec2 describe-instances --filters "Name=tag:Project,Values=OpenWan"

# Check RDS status
aws rds describe-db-instances --db-instance-identifier openwan-db

# Check ALB status
aws elbv2 describe-load-balancers --names openwan-alb
```

#### Step 2: Provision New Infrastructure (15 minutes)
```bash
# If using Terraform
cd /path/to/terraform
terraform plan
terraform apply -auto-approve

# If using CloudFormation
aws cloudformation create-stack \
  --stack-name openwan-recovery \
  --template-body file://infrastructure.yaml \
  --parameters file://parameters.json
```

#### Step 3: Restore Database (20 minutes)
```bash
# List available backups
aws rds describe-db-snapshots \
  --db-instance-identifier openwan-db \
  --snapshot-type automated

# Restore from latest snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier openwan-db-restored \
  --db-snapshot-identifier <snapshot-id> \
  --db-instance-class db.r5.large \
  --multi-az

# Wait for restoration (check status)
aws rds describe-db-instances \
  --db-instance-identifier openwan-db-restored \
  --query 'DBInstances[0].DBInstanceStatus'

# Update DNS/connection strings to point to new database
```

#### Step 4: Restore Redis Cache (5 minutes)
```bash
# Create new Redis cluster
aws elasticache create-cache-cluster \
  --cache-cluster-id openwan-redis-restored \
  --cache-node-type cache.r5.large \
  --engine redis \
  --num-cache-nodes 1

# Restore from snapshot (if available)
aws elasticache restore-cache-cluster-from-snapshot \
  --cache-cluster-id openwan-redis-restored \
  --snapshot-name <snapshot-name>

# Note: Cache will rebuild automatically from database on first access
```

#### Step 5: Deploy Application (10 minutes)
```bash
# Deploy backend API
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Or using Docker Compose
docker-compose up -d

# Verify deployment
kubectl get pods -l app=openwan-api
kubectl get svc openwan-api
```

#### Step 6: Restore Media Files (if needed) (15 minutes)
```bash
# S3 should be intact, but if needed:
# List S3 versions (if versioning enabled)
aws s3api list-object-versions --bucket openwan-media

# Restore from cross-region replication (if configured)
aws s3 sync s3://openwan-media-backup s3://openwan-media --region us-west-2
```

#### Step 7: Verify System (5 minutes)
```bash
# Health check
curl https://api.openwan.example.com/health

# Test login
curl -X POST https://api.openwan.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<password>"}'

# Test file access
curl https://api.openwan.example.com/api/v1/files?page=1&limit=10

# Check monitoring
# Open Grafana dashboard and verify metrics
```

#### Step 8: Communication
```
# Notify stakeholders
Subject: OpenWan System Restored - All Services Operational

The OpenWan system has been fully restored following the outage at <time>.

- Root cause: <description>
- Duration: <X> hours
- Data loss: <none/X minutes of data>
- Status: All services operational

Post-incident review scheduled for <date/time>.
```

**Total Estimated Time**: 60 minutes  
**Data Loss**: Up to 15 minutes (last automated backup)

---

## 2. Database Failure

### Symptoms
- "Database connection failed" errors in logs
- Health check endpoint returning 503
- API requests timing out

### Impact
- Service unavailable
- No data reads or writes

### Recovery Procedure

#### Step 1: Check Database Status (2 minutes)
```bash
# Check RDS instance status
aws rds describe-db-instances \
  --db-instance-identifier openwan-db \
  --query 'DBInstances[0].DBInstanceStatus'

# Check for recent events
aws rds describe-events \
  --source-identifier openwan-db \
  --duration 60

# Try direct connection
mysql -h <rds-endpoint> -u <username> -p<password> -e "SELECT 1"
```

#### Step 2: Determine Failure Type

##### If RDS Multi-AZ Automatic Failover (2 minutes)
- Multi-AZ RDS automatically fails over to standby
- Wait for failover to complete (typically < 2 minutes)
- Application should automatically reconnect

```bash
# Monitor failover progress
watch -n 5 'aws rds describe-db-instances \
  --db-instance-identifier openwan-db \
  --query "DBInstances[0].[DBInstanceStatus,AvailabilityZone]"'

# Check application logs for reconnection
kubectl logs -f deployment/openwan-api --tail=50
```

##### If Database Corrupted (20 minutes)
```bash
# Restore from point-in-time
RESTORE_TIME=$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ)

aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier openwan-db \
  --target-db-instance-identifier openwan-db-restored \
  --restore-time $RESTORE_TIME \
  --db-instance-class db.r5.large \
  --multi-az

# Wait for restoration
aws rds wait db-instance-available \
  --db-instance-identifier openwan-db-restored

# Update application configuration
kubectl set env deployment/openwan-api \
  DB_HOST=<new-db-endpoint>

# Delete old database (after verification)
aws rds delete-db-instance \
  --db-instance-identifier openwan-db \
  --skip-final-snapshot
```

#### Step 3: Verify Recovery (3 minutes)
```bash
# Test database connection
mysql -h <rds-endpoint> -u <username> -p<password> \
  -e "SELECT COUNT(*) FROM ow_users; SELECT COUNT(*) FROM ow_files;"

# Check application logs
kubectl logs deployment/openwan-api --tail=100 | grep -i database

# Test API
curl https://api.openwan.example.com/health
```

**Total Estimated Time**: 
- Automatic failover: 2-5 minutes
- Point-in-time restore: 25 minutes

**Data Loss**: 
- Automatic failover: None
- Point-in-time restore: Up to RPO (15 minutes)

---

## 3. Redis Cache Failure

### Symptoms
- Slower API response times
- "Redis connection failed" warnings in logs
- Cache miss rate at 100%

### Impact
- Service still available but slower
- Increased database load
- Session loss (users logged out)

### Recovery Procedure

#### Step 1: Assess Impact (2 minutes)
```bash
# Check Redis status
redis-cli -h <redis-endpoint> -p 6379 ping

# Check ElastiCache status (if AWS)
aws elasticache describe-cache-clusters \
  --cache-cluster-id openwan-redis \
  --show-cache-node-info

# Check application logs
kubectl logs deployment/openwan-api | grep -i redis | tail -50
```

#### Step 2: Initiate Failover (if Sentinel) (2 minutes)
```bash
# If using Redis Sentinel, automatic failover should occur
# Monitor Sentinel
redis-cli -h <sentinel-endpoint> -p 26379 sentinel masters

# Force manual failover if needed
redis-cli -h <sentinel-endpoint> -p 26379 \
  sentinel failover openwan-master

# Update application with new master endpoint (if not using Sentinel discovery)
kubectl set env deployment/openwan-api \
  REDIS_HOST=<new-master-endpoint>
```

#### Step 3: Rebuild Cache (5 minutes)
```bash
# Cache will rebuild automatically, but can warm it up
curl https://api.openwan.example.com/api/v1/admin/cache/warm

# Or manually via script
kubectl exec -it deployment/openwan-api -- /bin/sh
> curl localhost:8080/internal/cache/rebuild
```

#### Step 4: Verify Recovery (2 minutes)
```bash
# Check cache hit rate in Grafana
# Should gradually increase to 80%+

# Test Redis connection
redis-cli -h <redis-endpoint> -p 6379 INFO stats

# Check application
kubectl logs deployment/openwan-api | grep -i "cache" | tail -20
```

**Total Estimated Time**: 10 minutes  
**Data Loss**: Session data only (users need to re-login)

---

## 4. Application Server Failure

### Symptoms
- Some API requests failing
- Load balancer marking instances as unhealthy
- "Connection refused" or timeout errors

### Impact
- Reduced capacity
- Possible service degradation
- No impact if multiple instances healthy

### Recovery Procedure

#### Step 1: Identify Failed Instances (2 minutes)
```bash
# Check load balancer targets
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>

# Check Kubernetes pods
kubectl get pods -l app=openwan-api
kubectl describe pods -l app=openwan-api | grep -A 10 "Events:"

# Check pod logs for errors
kubectl logs <pod-name> --tail=100
```

#### Step 2: Restart Failed Instances (3 minutes)
```bash
# Kubernetes: Delete pod (will auto-recreate)
kubectl delete pod <pod-name>

# Or force rollout restart
kubectl rollout restart deployment/openwan-api

# EC2: Restart instance
aws ec2 reboot-instances --instance-ids <instance-id>

# Or terminate (auto-scaling will launch replacement)
aws ec2 terminate-instances --instance-ids <instance-id>
```

#### Step 3: Check for Pattern (5 minutes)
```bash
# If multiple instances failing, check for:

# 1. Resource exhaustion
kubectl top pods -l app=openwan-api
kubectl describe nodes

# 2. Application errors
kubectl logs deployment/openwan-api --all-containers=true | grep -i error

# 3. Configuration issues
kubectl get configmap openwan-config -o yaml
kubectl get secret openwan-secrets -o yaml

# 4. Database connection pool exhaustion
# Check metrics in Grafana for connection pool usage
```

#### Step 4: Scale if Needed (2 minutes)
```bash
# Temporarily increase replicas
kubectl scale deployment/openwan-api --replicas=5

# Or adjust HPA
kubectl patch hpa openwan-api -p '{"spec":{"minReplicas":3}}'
```

**Total Estimated Time**: 5-10 minutes  
**Data Loss**: None (stateless services)

---

## 5. S3 Storage Failure

### Symptoms
- "Failed to upload file" errors
- "File not found" errors for downloads
- S3 API errors in logs

### Impact
- Cannot upload new files
- Cannot download existing files
- Transcoding jobs fail

### Recovery Procedure

#### Step 1: Check S3 Status (2 minutes)
```bash
# Check AWS service health
aws health describe-events \
  --filter eventTypeCategories=issue serviceCode=S3

# Test S3 access
aws s3 ls s3://openwan-media/ --recursive | head -20

# Check bucket permissions
aws s3api get-bucket-policy --bucket openwan-media
aws s3api get-bucket-acl --bucket openwan-media
```

#### Step 2: Verify IAM Permissions (3 minutes)
```bash
# Check IAM role/user permissions
aws iam get-role-policy \
  --role-name openwan-api-role \
  --policy-name S3Access

# Test with AWS CLI using same credentials
export AWS_ACCESS_KEY_ID=<from-app-config>
export AWS_SECRET_ACCESS_KEY=<from-app-config>
aws s3 ls s3://openwan-media/
```

#### Step 3: Fallback or Restore (10 minutes)

##### If Bucket Deleted
```bash
# Check for cross-region replication backup
aws s3 ls s3://openwan-media-backup/ --region us-west-2

# Restore from backup
aws s3 sync s3://openwan-media-backup s3://openwan-media \
  --region us-west-2 \
  --source-region us-west-2

# If no backup, recreate bucket
aws s3 mb s3://openwan-media --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket openwan-media \
  --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
  --bucket openwan-media \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'
```

##### If Files Accidentally Deleted
```bash
# Restore from version (if versioning enabled)
aws s3api list-object-versions \
  --bucket openwan-media \
  --prefix data1/

# Restore specific file
aws s3api copy-object \
  --bucket openwan-media \
  --copy-source openwan-media/<key>?versionId=<version-id> \
  --key <key>
```

#### Step 4: Switch to Local Storage (if needed) (5 minutes)
```bash
# Update application configuration temporarily
kubectl set env deployment/openwan-api STORAGE_TYPE=local

# Update worker configuration
kubectl set env deployment/openwan-worker STORAGE_TYPE=local

# Verify
kubectl logs deployment/openwan-api | grep -i "storage type"
```

**Total Estimated Time**: 15-20 minutes  
**Data Loss**: None (if versioning enabled or backup exists)

---

## 6. RabbitMQ Message Queue Failure

### Symptoms
- "Failed to publish message" errors
- Transcoding jobs not processing
- Queue depth at 0 or growing uncontrollably

### Impact
- New transcoding jobs not queued
- Existing jobs may be lost
- Workers idle or overloaded

### Recovery Procedure

#### Step 1: Check RabbitMQ Status (2 minutes)
```bash
# Check RabbitMQ management API
curl -u admin:<password> http://<rabbitmq-host>:15672/api/overview

# Check cluster status
curl -u admin:<password> http://<rabbitmq-host>:15672/api/cluster-nodes

# Check queues
curl -u admin:<password> http://<rabbitmq-host>:15672/api/queues

# Or via command line
kubectl exec -it rabbitmq-0 -- rabbitmqctl cluster_status
kubectl exec -it rabbitmq-0 -- rabbitmqctl list_queues
```

#### Step 2: Restart RabbitMQ (if hung) (5 minutes)
```bash
# Kubernetes StatefulSet
kubectl rollout restart statefulset/rabbitmq

# Wait for pods to be ready
kubectl rollout status statefulset/rabbitmq

# Or individual node
kubectl delete pod rabbitmq-0
kubectl wait --for=condition=ready pod/rabbitmq-0
```

#### Step 3: Recover Messages (10 minutes)

##### If Messages Lost
```bash
# Check for message persistence
kubectl exec -it rabbitmq-0 -- rabbitmqctl list_queues name durable messages

# Requeue failed jobs from database
kubectl exec -it deployment/openwan-api -- /bin/sh
> /app/scripts/requeue-failed-jobs.sh

# Or manually
mysql -h <db-host> -u <user> -p -e "
  SELECT id, file_id FROM ow_transcode_jobs 
  WHERE status = 'pending' AND created_at > NOW() - INTERVAL 1 HOUR
" | while read id file_id; do
  curl -X POST http://localhost:8080/api/v1/internal/transcode/requeue/$file_id
done
```

##### If Queue Corrupted
```bash
# Delete and recreate queue
kubectl exec -it rabbitmq-0 -- rabbitmqctl delete_queue transcoding_jobs

# Queue will be auto-recreated by application
# Restart API to recreate queues
kubectl rollout restart deployment/openwan-api
```

#### Step 4: Verify Recovery (3 minutes)
```bash
# Check queue depth
curl -u admin:<password> \
  http://<rabbitmq-host>:15672/api/queues/%2F/transcoding_jobs \
  | jq '.messages'

# Check worker logs
kubectl logs deployment/openwan-worker --tail=50 | grep -i "processing job"

# Verify new jobs being created
curl -X POST https://api.openwan.example.com/api/v1/files/<file-id>/transcode
```

**Total Estimated Time**: 15-20 minutes  
**Data Loss**: Unacknowledged messages in flight (should be minimal with persistent queues)

---

## 7. Network/Load Balancer Failure

### Symptoms
- Cannot reach application from internet
- DNS resolution fails or times out
- SSL/TLS errors

### Impact
- Complete external service outage
- Internal services may still function

### Recovery Procedure

#### Step 1: Check Load Balancer (2 minutes)
```bash
# Check ALB status
aws elbv2 describe-load-balancers --names openwan-alb

# Check target group health
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>

# Check listeners
aws elbv2 describe-listeners --load-balancer-arn <lb-arn>

# Check security groups
aws ec2 describe-security-groups --group-ids <sg-id>
```

#### Step 2: Check DNS (2 minutes)
```bash
# Verify DNS resolution
nslookup api.openwan.example.com
dig api.openwan.example.com

# Check Route53 (if using)
aws route53 list-resource-record-sets \
  --hosted-zone-id <zone-id> \
  | jq '.ResourceRecordSets[] | select(.Name=="api.openwan.example.com.")'
```

#### Step 3: Recreate Load Balancer (if needed) (10 minutes)
```bash
# Create new ALB
aws elbv2 create-load-balancer \
  --name openwan-alb-new \
  --subnets <subnet-1> <subnet-2> <subnet-3> \
  --security-groups <sg-id> \
  --scheme internet-facing \
  --type application

# Create target group
aws elbv2 create-target-group \
  --name openwan-targets-new \
  --protocol HTTP \
  --port 8080 \
  --vpc-id <vpc-id> \
  --health-check-path /health

# Register targets
aws elbv2 register-targets \
  --target-group-arn <new-tg-arn> \
  --targets Id=<instance-1> Id=<instance-2>

# Create listener
aws elbv2 create-listener \
  --load-balancer-arn <new-lb-arn> \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=<cert-arn> \
  --default-actions Type=forward,TargetGroupArn=<new-tg-arn>

# Update DNS
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch file://dns-update.json
```

#### Step 4: Verify Connectivity (2 minutes)
```bash
# Test load balancer directly
curl -I http://<lb-dns-name>/health

# Test via domain
curl -I https://api.openwan.example.com/health

# Test SSL certificate
openssl s_client -connect api.openwan.example.com:443 -servername api.openwan.example.com
```

**Total Estimated Time**: 15-20 minutes  
**Data Loss**: None

---

## 8. Data Corruption or Accidental Deletion

### Symptoms
- Reports of missing or incorrect data
- "Record not found" errors
- Unexpected data values

### Impact
- Data integrity compromised
- User trust impacted
- Possible compliance issues

### Recovery Procedure

#### Step 1: Assess Scope (5 minutes)
```bash
# Identify affected data
mysql -h <db-host> -u <user> -p openwan_db -e "
  SELECT COUNT(*) as total_files FROM ow_files;
  SELECT COUNT(*) as total_users FROM ow_users;
  SELECT COUNT(*) as deleted_files FROM ow_files WHERE status = 4;
"

# Check recent changes
mysql -h <db-host> -u <user> -p openwan_db -e "
  SELECT * FROM ow_files WHERE updated_at > NOW() - INTERVAL 1 HOUR ORDER BY updated_at DESC LIMIT 100;
"

# Check application logs for deletion operations
kubectl logs deployment/openwan-api | grep -i "delete\|remove" | tail -100
```

#### Step 2: Stop Further Damage (2 minutes)
```bash
# If ongoing, scale down to prevent more changes
kubectl scale deployment/openwan-api --replicas=0

# Or put in maintenance mode
kubectl set env deployment/openwan-api MAINTENANCE_MODE=true
kubectl rollout restart deployment/openwan-api
```

#### Step 3: Restore Data (20 minutes)

##### From Database Backup
```bash
# Create restore database
mysql -h <db-host> -u <user> -p -e "CREATE DATABASE openwan_restore;"

# Restore from backup
mysqldump_file="openwan-backup-2024-02-01.sql"
aws s3 cp s3://openwan-backups/mysql/$mysqldump_file /tmp/

mysql -h <db-host> -u <user> -p openwan_restore < /tmp/$mysqldump_file

# Identify correct records
mysql -h <db-host> -u <user> -p openwan_restore -e "
  SELECT * FROM ow_files WHERE id IN (123, 456, 789);
"

# Restore specific records
mysql -h <db-host> -u <user> -p openwan_db -e "
  INSERT INTO ow_files SELECT * FROM openwan_restore.ow_files WHERE id IN (123, 456, 789)
  ON DUPLICATE KEY UPDATE 
    title = VALUES(title),
    status = VALUES(status),
    catalog_info = VALUES(catalog_info);
"
```

##### From S3 File Versions
```bash
# List deleted files
aws s3api list-object-versions \
  --bucket openwan-media \
  --prefix data1/ \
  --query 'DeleteMarkers[?IsLatest==`true`]'

# Restore deleted file
aws s3api delete-object \
  --bucket openwan-media \
  --key data1/abc123/file.mp4 \
  --version-id <delete-marker-version-id>

# Or copy previous version
aws s3api copy-object \
  --bucket openwan-media \
  --copy-source openwan-media/data1/abc123/file.mp4?versionId=<good-version-id> \
  --key data1/abc123/file.mp4
```

#### Step 4: Verify Restoration (3 minutes)
```bash
# Verify record count
mysql -h <db-host> -u <user> -p openwan_db -e "
  SELECT COUNT(*) FROM ow_files;
  SELECT id, title, status FROM ow_files WHERE id IN (123, 456, 789);
"

# Verify files accessible
curl https://api.openwan.example.com/api/v1/files/123
curl https://api.openwan.example.com/api/v1/files/456
```

#### Step 5: Resume Normal Operations (2 minutes)
```bash
# Scale back up
kubectl scale deployment/openwan-api --replicas=3

# Or disable maintenance mode
kubectl set env deployment/openwan-api MAINTENANCE_MODE=false
kubectl rollout restart deployment/openwan-api
```

**Total Estimated Time**: 30 minutes  
**Data Loss**: Depends on backup age (up to RPO: 15 minutes)

---

## 9. Security Breach

### Symptoms
- Unauthorized access alerts
- Unusual traffic patterns
- Modified or deleted data without authorization
- AWS GuardDuty or Security Hub alerts

### Impact
- Data confidentiality compromised
- System integrity compromised
- Compliance implications

### Recovery Procedure

#### Step 1: Contain Breach (IMMEDIATE - 5 minutes)
```bash
# 1. Block suspicious IPs at load balancer
aws elbv2 modify-listener --listener-arn <listener-arn> --default-actions Type=fixed-response,FixedResponseConfig={StatusCode=403}

# 2. Revoke compromised credentials
aws iam delete-access-key --access-key-id <compromised-key> --user-name <user>

# 3. Rotate all secrets
aws secretsmanager rotate-secret --secret-id openwan/database/password
aws secretsmanager rotate-secret --secret-id openwan/api/jwt-secret

# 4. Disable compromised user accounts
mysql -h <db-host> -u <user> -p openwan_db -e "
  UPDATE ow_users SET status = 'suspended' WHERE username IN ('suspicious_user1', 'suspicious_user2');
"

# 5. Take system offline if severe
kubectl scale deployment/openwan-api --replicas=0
```

#### Step 2: Investigate (15 minutes)
```bash
# Check CloudTrail logs
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=EventName,AttributeValue=DeleteObject \
  --start-time $(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%S) \
  --max-results 100

# Check application logs
kubectl logs deployment/openwan-api --since=24h | grep -E "login|auth|admin|delete" > /tmp/security-audit.log

# Check database audit log
mysql -h <db-host> -u <user> -p openwan_db -e "
  SELECT * FROM ow_audit_log WHERE created_at > NOW() - INTERVAL 24 HOUR ORDER BY created_at DESC;
"

# Check for malware or backdoors
kubectl exec -it deployment/openwan-api -- /bin/sh
> find /app -type f -mtime -1  # Files modified in last 24 hours
> ps aux  # Check for suspicious processes
```

#### Step 3: Eradicate Threat (10 minutes)
```bash
# Rebuild all containers from known good images
kubectl set image deployment/openwan-api \
  openwan-api=<registry>/openwan-api:<known-good-tag>

kubectl rollout restart deployment/openwan-api
kubectl rollout restart deployment/openwan-worker

# Reset all user passwords (force password change)
mysql -h <db-host> -u <user> -p openwan_db -e "
  UPDATE ow_users SET password_reset_required = 1;
"

# Update security groups (allowlist only known IPs)
aws ec2 revoke-security-group-ingress --group-id <sg-id> --ip-permissions <current-rules>
aws ec2 authorize-security-group-ingress --group-id <sg-id> --ip-permissions <new-restricted-rules>
```

#### Step 4: Restore Clean State (15 minutes)
```bash
# Restore database from pre-breach backup
# Identify last known good backup time
aws rds describe-db-snapshots \
  --db-instance-identifier openwan-db \
  | jq '.DBSnapshots[] | select(.SnapshotCreateTime < "2024-02-01T10:00:00Z")'

# Restore
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier openwan-db-clean \
  --db-snapshot-identifier <clean-snapshot-id>

# Manually review and restore any legitimate changes made after backup
# This requires business logic understanding
```

#### Step 5: Strengthen Security (10 minutes)
```bash
# Enable MFA for all admin users
# Enable AWS CloudTrail logging
aws cloudtrail create-trail --name openwan-audit --s3-bucket-name openwan-audit-logs

# Enable VPC Flow Logs
aws ec2 create-flow-logs \
  --resource-type VPC \
  --resource-ids <vpc-id> \
  --traffic-type ALL \
  --log-destination-type s3 \
  --log-destination arn:aws:s3:::openwan-flow-logs

# Update WAF rules (if applicable)
# Add rate limiting, IP reputation, SQL injection protection

# Enable GuardDuty (if not already)
aws guardduty create-detector --enable
```

#### Step 6: Resume Operations (5 minutes)
```bash
# Scale back up
kubectl scale deployment/openwan-api --replicas=3

# Verify security
curl https://api.openwan.example.com/health

# Monitor closely
watch -n 5 'kubectl logs deployment/openwan-api --tail=20 | grep -i error'
```

#### Step 7: Post-Incident
- **Document**: Create detailed incident report
- **Notify**: Inform affected users if data was exposed
- **Compliance**: Report to authorities if required (GDPR, etc.)
- **Lessons Learned**: Conduct post-mortem meeting
- **Improve**: Implement additional security controls

**Total Estimated Time**: 60 minutes (contain and eradicate)  
**Data Loss**: Depends on breach scope and restore point

---

## 10. Regional Outage

### Symptoms
- All AWS services in region unavailable
- Cannot access EC2, RDS, S3, etc.
- AWS Service Health Dashboard showing regional issues

### Impact
- Complete service outage in primary region
- Requires failover to backup region

### Recovery Procedure

**Note**: This requires pre-configured multi-region setup with cross-region replication.

#### Step 1: Verify Regional Outage (5 minutes)
```bash
# Check AWS Service Health Dashboard
aws health describe-events --filter eventTypeCategories=issue

# Attempt to access services in affected region
aws ec2 describe-instances --region us-east-1
aws s3 ls s3://openwan-media --region us-east-1
aws rds describe-db-instances --region us-east-1
```

#### Step 2: Activate Backup Region (10 minutes)
```bash
# Switch to backup region
export AWS_DEFAULT_REGION=us-west-2

# Verify backup infrastructure exists
aws ec2 describe-instances --filters "Name=tag:Project,Values=OpenWan"
aws rds describe-db-instances --db-instance-identifier openwan-db-replica
aws s3 ls s3://openwan-media-replica

# Promote read replica to master
aws rds promote-read-replica \
  --db-instance-identifier openwan-db-replica-us-west-2

# Wait for promotion
aws rds wait db-instance-available \
  --db-instance-identifier openwan-db-replica-us-west-2
```

#### Step 3: Update DNS (5 minutes)
```bash
# Update Route53 to point to backup region
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch '{
    "Changes": [{
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "api.openwan.example.com",
        "Type": "CNAME",
        "TTL": 60,
        "ResourceRecords": [{"Value": "<backup-region-lb-dns>"}]
      }
    }]
  }'

# Wait for DNS propagation
watch -n 5 'dig api.openwan.example.com +short'
```

#### Step 4: Deploy Application in Backup Region (10 minutes)
```bash
# If not already deployed, deploy application
kubectl apply -f k8s/deployment.yaml --context backup-region-cluster
kubectl apply -f k8s/service.yaml --context backup-region-cluster
kubectl apply -f k8s/ingress.yaml --context backup-region-cluster

# Update configuration to use backup region resources
kubectl set env deployment/openwan-api \
  DB_HOST=<backup-region-db-endpoint> \
  REDIS_HOST=<backup-region-redis-endpoint> \
  S3_REGION=us-west-2 \
  S3_BUCKET=openwan-media-replica \
  --context backup-region-cluster
```

#### Step 5: Verify Service (5 minutes)
```bash
# Test health
curl https://api.openwan.example.com/health

# Test functionality
curl -X POST https://api.openwan.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"<password>"}'

# Verify files accessible from replica S3 bucket
curl https://api.openwan.example.com/api/v1/files/123
```

#### Step 6: Monitor and Communicate (ongoing)
```
Subject: OpenWan Service Running from Backup Region

Due to an AWS regional outage in us-east-1, the OpenWan service has been 
failed over to our backup region (us-west-2).

- Service Status: Operational
- Data Status: All data intact (replicated)
- Performance: Normal (slight latency increase possible)
- Action Required: None

We are monitoring the situation and will fail back to the primary region 
once AWS declares the outage resolved.
```

#### Step 7: Failback to Primary Region (when resolved)
```bash
# Wait for AWS to declare primary region healthy
# Verify primary region services operational
aws ec2 describe-instances --region us-east-1

# Replicate any new data from backup to primary
# This depends on your replication setup

# Update DNS back to primary region
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch <primary-region-update>

# Scale down backup region (keep as hot standby or cold standby)
kubectl scale deployment/openwan-api --replicas=1 --context backup-region-cluster
```

**Total Estimated Time**: 30-40 minutes (if pre-configured)  
**Data Loss**: Minimal (depends on replication lag, typically <5 minutes)

---

## Testing and Validation

### Disaster Recovery Drills

Conduct quarterly DR drills to validate procedures:

#### Q1: Database Failover Drill
- Simulate database failure
- Perform manual failover (RDS Multi-AZ)
- Measure RTO and RPO
- Document findings

#### Q2: Complete System Recovery Drill
- Take down entire system in non-production
- Restore from backups
- Measure recovery time
- Validate data integrity

#### Q3: Security Breach Simulation
- Simulate compromised credentials
- Practice containment procedures
- Review audit logs
- Test communication plan

#### Q4: Regional Failover Drill
- Fail over to backup region
- Test all functionality
- Measure performance
- Fail back to primary

### Validation Checklist

After each recovery, validate:

- [ ] All services healthy (`/health` returns 200)
- [ ] Database accessible and data intact
- [ ] Cache functioning (check hit rate)
- [ ] File upload works
- [ ] File download works
- [ ] Search functionality works
- [ ] Transcoding jobs processing
- [ ] User authentication works
- [ ] Admin panels accessible
- [ ] Monitoring and metrics collecting
- [ ] Logs aggregating
- [ ] Backups running

## Post-Incident Activities

### 1. Incident Report Template
```
Incident Report: <Incident Title>
Date: <YYYY-MM-DD>
Severity: <Critical/High/Medium/Low>
Duration: <X hours Y minutes>

Timeline:
- HH:MM - Incident detected
- HH:MM - Response team assembled
- HH:MM - Root cause identified
- HH:MM - Mitigation applied
- HH:MM - Service restored
- HH:MM - Incident closed

Root Cause:
<Detailed description>

Impact:
- Users affected: <count or percentage>
- Services affected: <list>
- Data loss: <none or description>
- Financial impact: <if applicable>

Response:
<What was done to resolve>

Prevention:
<How to prevent in future>

Action Items:
1. <Action> - Owner: <Name> - Due: <Date>
2. <Action> - Owner: <Name> - Due: <Date>
```

### 2. Blameless Post-Mortem
- Schedule within 48 hours of resolution
- Focus on process improvement, not blame
- Invite all stakeholders
- Document lessons learned
- Track action items to completion

### 3. Update Runbook
- Add any new procedures discovered
- Update contact information
- Update infrastructure details
- Version control changes

## Appendix

### A. Important Endpoints

| Service | Endpoint | Port | Health Check |
|---------|----------|------|--------------|
| API | https://api.openwan.example.com | 443 | /health |
| Frontend | https://openwan.example.com | 443 | / |
| Metrics | http://<internal-ip>:9090 | 9090 | /metrics |
| RabbitMQ Mgmt | http://<internal-ip>:15672 | 15672 | /api/overview |

### B. Configuration Files Locations

- Application Config: `/etc/openwan/config.yaml`
- Kubernetes Manifests: `k8s/`
- Terraform: `terraform/`
- Backup Scripts: `scripts/backup/`
- Environment Variables: `kubectl get secret openwan-secrets`

### C. Backup Schedule

| Item | Frequency | Retention | Location |
|------|-----------|-----------|----------|
| Database | Automated (RDS) | 7 days | AWS RDS |
| Database Manual | Daily | 30 days | S3 |
| Redis Snapshot | Every 6 hours | 24 hours | S3 |
| Config Files | On change | 90 days | S3 |
| Application Logs | Continuous | 30 days | CloudWatch |

### D. Monitoring Alerts

| Alert | Threshold | Severity | Action |
|-------|-----------|----------|--------|
| API Down | Health check fails 3x | Critical | Page on-call |
| Database Down | Connection fails | Critical | Page on-call |
| High Error Rate | >1% of requests | High | Notify team |
| High Response Time | p95 >1s | High | Investigate |
| Disk Usage High | >80% | Medium | Plan expansion |
| Queue Depth High | >1000 messages | Medium | Scale workers |

---

**Document Version**: 1.0  
**Last Updated**: 2024-02-01  
**Last Tested**: [Schedule quarterly DR drills]  
**Next Review Date**: 2024-05-01  
**Maintained By**: OpenWan DevOps Team

**Emergency Hotline**: [Add number]  
**Slack Channel**: #openwan-incidents  
**PagerDuty**: [Add integration]
