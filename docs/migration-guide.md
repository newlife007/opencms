# Migration Guide: PHP OpenWan to Go/Vue System

## Table of Contents
1. [Overview](#overview)
2. [Pre-Migration Checklist](#pre-migration-checklist)
3. [Database Migration](#database-migration)
4. [File Storage Migration](#file-storage-migration)
5. [User Data Migration](#user-data-migration)
6. [Testing Migration](#testing-migration)
7. [Cutover Plan](#cutover-plan)
8. [Rollback Procedures](#rollback-procedures)
9. [Post-Migration Validation](#post-migration-validation)

## Overview

This guide provides step-by-step instructions for migrating from the legacy PHP5.x/QeePHP OpenWan system to the new Go/Vue architecture.

### Migration Goals
- Zero data loss
- Minimal downtime (target: <1 hour)
- Preserve all functionality
- Maintain user sessions (if possible)
- Validate data integrity

### Migration Timeline
- **Week 1-2**: Preparation and backup
- **Week 3**: Data migration to staging
- **Week 4**: Validation and testing
- **Week 5**: Production migration (scheduled maintenance window)
- **Week 6+**: Parallel operation and final cutover

## Pre-Migration Checklist

### 1. Backup Everything
```bash
# Backup legacy PHP application
cd /path/to/legacy/openwan
tar -czf ~/backups/openwan-php-$(date +%Y%m%d).tar.gz .

# Backup legacy database
mysqldump -h <legacy-db-host> -u <user> -p openwan_db \
  --single-transaction --routines --triggers \
  > ~/backups/openwan-db-$(date +%Y%m%d).sql

# Backup files directory
rsync -av /path/to/media/files/ ~/backups/media-files/
```

### 2. Document Current System
```bash
# Document database schema
mysqldump -h <legacy-db-host> -u <user> -p openwan_db \
  --no-data --routines --triggers \
  > ~/docs/legacy-schema.sql

# Count records in each table
mysql -h <legacy-db-host> -u <user> -p openwan_db -e "
  SELECT 'ow_users' as table_name, COUNT(*) as row_count FROM ow_users
  UNION ALL
  SELECT 'ow_groups', COUNT(*) FROM ow_groups
  UNION ALL
  SELECT 'ow_files', COUNT(*) FROM ow_files
  UNION ALL
  SELECT 'ow_categories', COUNT(*) FROM ow_categories
  UNION ALL
  SELECT 'ow_catalog', COUNT(*) FROM ow_catalog;
"

# Document file counts
find /path/to/media/files -type f | wc -l
du -sh /path/to/media/files
```

### 3. Verify New System Readiness
```bash
# Check new database is created
mysql -h <new-db-host> -u <user> -p -e "SHOW DATABASES LIKE 'openwan%';"

# Verify migrations are ready
cd /home/ec2-user/openwan
ls -l migrations/

# Test application builds
go build -o bin/openwan ./cmd/api
cd frontend && npm run build

# Check infrastructure is provisioned
kubectl get pods -l app=openwan-api
kubectl get svc
aws s3 ls s3://openwan-media/
```

## Database Migration

### Step 1: Export Legacy Database

```bash
# Set environment variables
export LEGACY_DB_HOST="legacy-mysql.example.com"
export LEGACY_DB_USER="openwan_user"
export LEGACY_DB_NAME="openwan_db"
export LEGACY_DB_PASSWORD="<password>"

# Export with specific tables
mysqldump -h $LEGACY_DB_HOST -u $LEGACY_DB_USER -p$LEGACY_DB_PASSWORD \
  $LEGACY_DB_NAME \
  ow_users ow_groups ow_roles ow_permissions ow_levels \
  ow_categories ow_catalog ow_files \
  ow_groups_has_roles ow_roles_has_permissions ow_groups_has_category \
  --single-transaction --skip-lock-tables --complete-insert \
  > ~/migration/legacy-data-export.sql

# Generate data checksums for validation
mysql -h $LEGACY_DB_HOST -u $LEGACY_DB_USER -p$LEGACY_DB_PASSWORD \
  $LEGACY_DB_NAME -e "
  SELECT 
    'users' as entity,
    COUNT(*) as count,
    MD5(GROUP_CONCAT(id ORDER BY id)) as checksum
  FROM ow_users
  UNION ALL
  SELECT 'files', COUNT(*), MD5(GROUP_CONCAT(id ORDER BY id)) FROM ow_files
  UNION ALL
  SELECT 'categories', COUNT(*), MD5(GROUP_CONCAT(id ORDER BY id)) FROM ow_categories;
" > ~/migration/legacy-checksums.txt
```

### Step 2: Run Schema Migration on New Database

```bash
# Set new database environment variables
export DB_HOST="new-mysql.example.com"
export DB_USER="openwan_user"
export DB_NAME="openwan_db"
export DB_PASSWORD="<new-password>"

# Run migrations to create schema
cd /home/ec2-user/openwan
# Using golang-migrate tool
migrate -path migrations -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:3306)/$DB_NAME" up

# Or using application
./bin/openwan migrate up

# Verify schema created
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "SHOW TABLES;"
```

### Step 3: Data Transformation and Import

```bash
# Use the data migration script
cd /home/ec2-user/openwan/scripts

# Set environment variables for both databases
export LEGACY_DB_HOST="legacy-mysql.example.com"
export LEGACY_DB_USER="openwan_user"
export LEGACY_DB_PASSWORD="<legacy-password>"
export LEGACY_DB_NAME="openwan_db"

export DB_HOST="new-mysql.example.com"
export DB_USER="openwan_user"
export DB_PASSWORD="<new-password>"
export DB_NAME="openwan_db"

# Run migration script
go run migrate_data.go 2>&1 | tee ~/migration/data-migration.log

# Expected output:
# Migrating users...
# Migrated 150 users
# Migrating groups...
# Migrated 25 groups
# ...
```

### Step 4: Validate Data Migration

```bash
# Compare record counts
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT 
    'users' as entity,
    COUNT(*) as count,
    MD5(GROUP_CONCAT(id ORDER BY id)) as checksum
  FROM ow_users
  UNION ALL
  SELECT 'files', COUNT(*), MD5(GROUP_CONCAT(id ORDER BY id)) FROM ow_files
  UNION ALL
  SELECT 'categories', COUNT(*), MD5(GROUP_CONCAT(id ORDER BY id)) FROM ow_categories;
" > ~/migration/new-checksums.txt

# Compare checksums
diff ~/migration/legacy-checksums.txt ~/migration/new-checksums.txt

# If checksums match, data is identical
# If not, investigate discrepancies

# Detailed validation queries
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME << EOF
-- Check for orphaned records
SELECT 'Files without categories' as issue, COUNT(*) as count
FROM ow_files f
LEFT JOIN ow_categories c ON f.category_id = c.id
WHERE f.category_id > 0 AND c.id IS NULL

UNION ALL

SELECT 'Users without groups', COUNT(*)
FROM ow_users u
LEFT JOIN ow_groups g ON u.group_id = g.id
WHERE u.group_id > 0 AND g.id IS NULL

UNION ALL

SELECT 'Invalid catalog_info JSON', COUNT(*)
FROM ow_files
WHERE catalog_info IS NOT NULL 
  AND catalog_info != ''
  AND JSON_VALID(catalog_info) = 0;
EOF
```

## File Storage Migration

### Option 1: Migrate to S3 (Recommended for Production)

```bash
# Configure AWS CLI
aws configure set region us-east-1

# Create S3 bucket (if not exists)
aws s3 mb s3://openwan-media

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket openwan-media \
  --versioning-configuration Status=Enabled

# Enable server-side encryption
aws s3api put-bucket-encryption \
  --bucket openwan-media \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Sync files to S3 maintaining directory structure
# Legacy structure: /data1/md5dir/md5filename.ext
cd /path/to/legacy/media/files

# Dry run first
aws s3 sync . s3://openwan-media/ --dryrun

# Actual sync with progress
aws s3 sync . s3://openwan-media/ \
  --storage-class STANDARD \
  --metadata-directive COPY \
  --size-only \
  | tee ~/migration/s3-sync.log

# Verify file count
aws s3 ls s3://openwan-media/ --recursive | wc -l
find . -type f | wc -l  # Should match

# Verify sample files integrity
md5sum /path/to/legacy/media/files/data1/abc123/file.mp4
aws s3api head-object --bucket openwan-media --key data1/abc123/file.mp4 \
  | jq -r '.ETag' | tr -d '"'  # ETag is MD5 for non-multipart uploads
```

### Option 2: Local Storage Migration

```bash
# Create storage directory for new system
mkdir -p /mnt/openwan-storage/{data1,data2,data3}

# Sync files maintaining structure
rsync -av --progress \
  /path/to/legacy/media/files/ \
  /mnt/openwan-storage/ \
  | tee ~/migration/files-sync.log

# Verify file counts
find /path/to/legacy/media/files -type f | wc -l
find /mnt/openwan-storage -type f | wc -l

# Verify sample files
md5sum /path/to/legacy/media/files/data1/abc123/file.mp4
md5sum /mnt/openwan-storage/data1/abc123/file.mp4

# Set proper permissions
chown -R openwan:openwan /mnt/openwan-storage
chmod -R 755 /mnt/openwan-storage
```

### Update Database File Paths (if needed)

```bash
# If file paths need updating in database
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME << EOF
-- Check current paths
SELECT path FROM ow_files LIMIT 10;

-- Update paths if needed (example: prepending new base path)
-- UPDATE ow_files SET path = CONCAT('/mnt/openwan-storage/', path);

-- For S3, paths might stay the same (just storage type changes in config)
EOF
```

## User Data Migration

### Password Migration

```bash
# Legacy system uses MD5 or custom hashing
# New system uses bcrypt

# Option 1: Force password reset for all users
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  UPDATE ow_users SET password_reset_required = 1;
"

# Option 2: Create password migration script (if you know legacy hash method)
# This allows users to login with old password once, then rehash to bcrypt
# Implement in application login handler:
# if isLegacyHash(user.Password) {
#   if verifyLegacyPassword(inputPassword, user.Password) {
#     newHash = bcrypt.GenerateFromPassword(inputPassword)
#     updateUserPassword(user.ID, newHash)
#     return success
#   }
# }
```

### Session Migration (Optional)

```bash
# Legacy PHP sessions (file-based)
# New Go sessions (Redis-based)

# Session migration is typically not performed due to:
# 1. Different session formats
# 2. Short TTL of sessions
# 3. Security considerations

# Recommended: Force all users to re-login
# Send email notification:
cat > ~/migration/session-migration-email.txt << 'EOF'
Subject: OpenWan System Upgrade - Please Login Again

Dear OpenWan User,

We have upgraded the OpenWan system to improve performance and security.
As part of this upgrade, all users will need to log in again.

Please visit https://openwan.example.com and log in with your existing credentials.

If you have any issues, please contact support.

Thank you,
OpenWan Team
EOF
```

## Testing Migration

### Step 1: Deploy to Staging

```bash
# Deploy new system to staging environment
cd /home/ec2-user/openwan

# Update staging configuration
cp configs/config.example.yaml configs/config.staging.yaml
# Edit config.staging.yaml with staging database and storage settings

# Deploy backend
kubectl apply -f k8s/deployment.yaml --namespace staging
kubectl apply -f k8s/service.yaml --namespace staging

# Build and deploy frontend
cd frontend
npm run build
# Upload dist/ to staging CDN or server
```

### Step 2: Run Test Suite

```bash
# Functional tests
cd /home/ec2-user/openwan/tests

# Test authentication
./test_auth.sh staging.openwan.example.com

# Test file operations
./test_files.sh staging.openwan.example.com

# Test search
./test_search.sh staging.openwan.example.com

# Test admin operations
./test_admin.sh staging.openwan.example.com
```

### Step 3: User Acceptance Testing

```
UAT Checklist:

1. Authentication & Authorization
   [ ] Admin user can login
   [ ] Regular user can login
   [ ] User sees correct permissions based on role
   [ ] Access to restricted files is enforced

2. File Upload
   [ ] Can upload video file
   [ ] Can upload audio file
   [ ] Can upload image file
   [ ] File type validation works
   [ ] File size limit enforced
   [ ] Upload progress shows correctly

3. File Cataloging
   [ ] Can view file details
   [ ] Can edit catalog metadata
   [ ] Dynamic form fields display correctly
   [ ] Save catalog changes works
   [ ] Required field validation works

4. File Search
   [ ] Search by title works
   [ ] Filter by file type works
   [ ] Filter by category works
   [ ] Filter by status works
   [ ] Search results respect user permissions
   [ ] Pagination works

5. File Download
   [ ] Can download published files
   [ ] Cannot download unpublished files (if no permission)
   [ ] Permission checks work correctly
   [ ] File streams correctly (not corrupt)

6. Video Playback
   [ ] Video player loads
   [ ] Can play FLV preview files
   [ ] Player controls work (play, pause, seek, volume)
   [ ] Video quality is acceptable

7. File Workflow
   [ ] Can submit file for review (status: new -> pending)
   [ ] Reviewer can approve (status: pending -> published)
   [ ] Reviewer can reject (status: pending -> rejected)
   [ ] Can delete file (status: -> deleted)

8. Admin Functions
   [ ] Can create/edit/delete users
   [ ] Can create/edit/delete groups
   [ ] Can create/edit/delete roles
   [ ] Can assign users to groups
   [ ] Can assign roles to groups
   [ ] Can assign permissions to roles
   [ ] Can manage categories (create/edit/delete/reorder)
   [ ] Can manage catalog configuration

9. Performance
   [ ] Page load times acceptable (<3 seconds)
   [ ] Search responds quickly (<1 second)
   [ ] File upload speed reasonable
   [ ] Video playback smooth (no buffering)

10. UI/UX
    [ ] All Chinese text displays correctly
    [ ] UI layout matches legacy system (or improvements noted)
    [ ] Navigation is intuitive
    [ ] Error messages are clear
    [ ] No console errors in browser
```

## Cutover Plan

### Phase 1: Pre-Cutover (1 week before)

```bash
# Week before cutover:

# 1. Final staging migration and testing
# Run full migration on staging with latest production data
# Perform complete UAT

# 2. Communication
# Send announcement email to all users:
cat > ~/migration/cutover-announcement.txt << 'EOF'
Subject: OpenWan System Upgrade Scheduled - [Date] [Time]

Dear OpenWan Users,

We will be upgrading the OpenWan media asset management system on:

Date: [Saturday, February 10, 2024]
Time: [02:00 AM - 04:00 AM UTC] (Expected 2-hour maintenance window)

During this time, the system will be unavailable.

What to expect:
- All data will be preserved
- You will need to log in again after the upgrade
- The system will be faster and more reliable
- New features: [list if any]

Please save your work and log out before the maintenance window.

Thank you for your patience.
OpenWan Team
EOF

# 3. Prepare rollback plan
# Document current production configuration
# Take final backups
# Have legacy system ready to restore if needed
```

### Phase 2: Cutover Execution (2-hour maintenance window)

```bash
#!/bin/bash
# cutover.sh - Execute production cutover

set -e  # Exit on error

echo "=== OpenWan Production Cutover Script ==="
echo "Start time: $(date)"

# Step 1: Put legacy system in maintenance mode (T+0:00)
echo "[T+0:00] Putting legacy system in maintenance mode..."
ssh legacy-server "sudo systemctl stop apache2"
# Or display maintenance page
ssh legacy-server "cp /var/www/html/maintenance.html /var/www/html/index.html"

# Step 2: Final legacy database backup (T+0:05)
echo "[T+0:05] Taking final legacy database backup..."
ssh legacy-db-server "mysqldump -u root -p<password> openwan_db > /backups/final-backup-$(date +%Y%m%d-%H%M%S).sql"

# Step 3: Run data migration (T+0:10)
echo "[T+0:10] Running data migration..."
export LEGACY_DB_HOST="legacy-mysql.example.com"
export DB_HOST="new-mysql.example.com"
# ... set other env vars
cd /home/ec2-user/openwan/scripts
go run migrate_data.go 2>&1 | tee ~/migration/production-migration-$(date +%Y%m%d-%H%M%S).log

# Step 4: Validate data migration (T+0:30)
echo "[T+0:30] Validating data migration..."
./validate_migration.sh
if [ $? -ne 0 ]; then
  echo "ERROR: Data validation failed!"
  echo "Initiating rollback..."
  ./rollback.sh
  exit 1
fi

# Step 5: Sync remaining files (T+0:40)
echo "[T+0:40] Syncing files to S3..."
# Only sync files modified in last 24 hours (delta)
find /legacy/media/files -type f -mtime -1 -print0 | \
  xargs -0 -I {} aws s3 cp {} s3://openwan-media/{} --quiet

# Step 6: Deploy new system (T+0:50)
echo "[T+0:50] Deploying new Go/Vue system..."
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl rollout status deployment/openwan-api

# Step 7: Update DNS (T+1:00)
echo "[T+1:00] Updating DNS to point to new system..."
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch file://dns-update-new-system.json

# Wait for DNS propagation
echo "Waiting for DNS propagation (60 seconds)..."
sleep 60

# Step 8: Smoke tests (T+1:05)
echo "[T+1:05] Running smoke tests..."
./smoke_tests.sh https://api.openwan.example.com
if [ $? -ne 0 ]; then
  echo "ERROR: Smoke tests failed!"
  echo "Initiating rollback..."
  ./rollback.sh
  exit 1
fi

# Step 9: Monitor (T+1:10)
echo "[T+1:10] Cutover complete. Monitoring..."
echo "Check Grafana dashboard: https://monitoring.openwan.example.com"
echo "Check application logs: kubectl logs -f deployment/openwan-api"
echo "Monitor for 30 minutes before declaring success."

echo "=== Cutover Script Complete ==="
echo "End time: $(date)"
```

### Phase 3: Post-Cutover Monitoring (First 24 hours)

```bash
# Monitor key metrics
watch -n 30 'kubectl get pods -l app=openwan-api'
watch -n 30 'curl -s https://api.openwan.example.com/health | jq'

# Check error logs
kubectl logs -f deployment/openwan-api --all-containers=true | grep -i error

# Monitor Grafana dashboards
# - Request rate
# - Error rate
# - Response time
# - Database connections
# - Cache hit rate

# User feedback
# Monitor support tickets
# Check user login success rate
# Watch for unusual patterns
```

## Rollback Procedures

### Scenario 1: Rollback Before DNS Update

```bash
#!/bin/bash
# rollback_pre_dns.sh

echo "=== Executing Rollback (Pre-DNS Update) ==="

# Step 1: Stop new system
kubectl scale deployment/openwan-api --replicas=0
kubectl scale deployment/openwan-worker --replicas=0

# Step 2: Restore legacy system
ssh legacy-server "sudo systemctl start apache2"
ssh legacy-server "rm /var/www/html/index.html"  # Remove maintenance page

# Step 3: Verify legacy system
curl https://legacy.openwan.example.com/health

echo "Rollback complete. Legacy system restored."
```

### Scenario 2: Rollback After DNS Update

```bash
#!/bin/bash
# rollback_post_dns.sh

echo "=== Executing Rollback (Post-DNS Update) ==="

# Step 1: Revert DNS
aws route53 change-resource-record-sets \
  --hosted-zone-id <zone-id> \
  --change-batch file://dns-update-legacy-system.json

# Wait for DNS propagation
echo "Waiting for DNS propagation..."
sleep 120

# Step 2: Stop new system
kubectl scale deployment/openwan-api --replicas=0

# Step 3: Check if any data was modified in new system
# If yes, need to sync back to legacy database
mysql -h new-db-host -u user -p new_db -e "
  SELECT COUNT(*) as new_records FROM ow_files 
  WHERE created_at > '<cutover-timestamp>';
"

# If new records exist, manually review and migrate back if needed
# This is complex and should be avoided by:
# - Thorough testing before cutover
# - Having rollback decision point before significant data changes

# Step 4: Restart legacy system
ssh legacy-server "sudo systemctl start apache2"

echo "Rollback complete. Legacy system active."
echo "Manual review required for any data created in new system."
```

### Rollback Decision Matrix

| Time Since Cutover | New Data Created | Rollback Complexity | Recommendation |
|-------------------|------------------|---------------------|----------------|
| < 30 minutes | No | Low | Safe to rollback |
| < 30 minutes | Yes (< 10 records) | Medium | Rollback and manually migrate |
| 30min - 2 hours | Yes (< 100 records) | Medium-High | Consider forward fix |
| > 2 hours | Yes (> 100 records) | Very High | Forward fix only |

## Post-Migration Validation

### Day 1: Intensive Monitoring

```bash
# Checklist for first 24 hours

# 1. System Health
curl https://api.openwan.example.com/health
# Expected: {"status": "healthy", "database": "connected", ...}

# 2. Error Rate
# Check Grafana: Error rate should be < 0.1%

# 3. Response Time
# Check Grafana: P95 response time should be < 500ms

# 4. User Activity
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT DATE(created_at) as date, COUNT(*) as login_count 
  FROM ow_user_sessions 
  WHERE created_at > NOW() - INTERVAL 24 HOUR 
  GROUP BY DATE(created_at);
"

# 5. File Operations
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT 
    COUNT(*) as uploads_last_24h 
  FROM ow_files 
  WHERE upload_at > NOW() - INTERVAL 24 HOUR;
"

# 6. Transcoding Jobs
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT status, COUNT(*) as count 
  FROM ow_transcode_jobs 
  WHERE created_at > NOW() - INTERVAL 24 HOUR 
  GROUP BY status;
"
```

### Week 1: Data Integrity Validation

```bash
# Run weekly validation script
#!/bin/bash
# weekly_validation.sh

echo "=== Week 1 Validation Report ==="

# 1. Record count comparison
echo "Record Counts:"
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT 'ow_users' as table_name, COUNT(*) as count FROM ow_users
  UNION ALL SELECT 'ow_files', COUNT(*) FROM ow_files
  UNION ALL SELECT 'ow_categories', COUNT(*) FROM ow_categories;
"

# 2. Check for data anomalies
echo "Data Anomalies:"
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  -- Files without categories
  SELECT 'Orphaned files' as issue, COUNT(*) as count FROM ow_files WHERE category_id NOT IN (SELECT id FROM ow_categories)
  UNION ALL
  -- Invalid JSON in catalog_info
  SELECT 'Invalid catalog JSON', COUNT(*) FROM ow_files WHERE catalog_info IS NOT NULL AND JSON_VALID(catalog_info) = 0
  UNION ALL
  -- Users without groups
  SELECT 'Orphaned users', COUNT(*) FROM ow_users WHERE group_id NOT IN (SELECT id FROM ow_groups);
"

# 3. File storage validation
echo "File Storage:"
if [ "$STORAGE_TYPE" == "s3" ]; then
  aws s3 ls s3://openwan-media/ --recursive | wc -l
else
  find /mnt/openwan-storage -type f | wc -l
fi

# 4. Performance metrics
echo "Performance (P95 response time last 7 days):"
# Query Prometheus or Grafana API
```

### Month 1: Legacy System Decommission

```bash
# After 30 days of successful operation

# 1. Final backup of legacy system
ssh legacy-server "mysqldump -u root -p<password> openwan_db > /archive/legacy-final-backup-$(date +%Y%m%d).sql"
tar -czf /archive/legacy-php-application-final.tar.gz /var/www/html/openwan

# 2. Upload to long-term storage
aws s3 cp /archive/legacy-final-backup-$(date +%Y%m%d).sql \
  s3://openwan-archives/legacy/ \
  --storage-class GLACIER

# 3. Shut down legacy servers
# After approval from stakeholders:
ssh legacy-server "sudo systemctl stop apache2 mysql"
aws ec2 stop-instances --instance-ids <legacy-instance-ids>

# 4. Keep legacy servers stopped for 90 days before termination
# In case urgent need to reference or restore
```

## Troubleshooting

### Issue: File Downloads Failing

```bash
# Check file paths in database
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT id, name, path FROM ow_files WHERE id = <failing-file-id>;
"

# Verify file exists
if [ "$STORAGE_TYPE" == "s3" ]; then
  aws s3 ls s3://openwan-media/<file-path>
else
  ls -lh /mnt/openwan-storage/<file-path>
fi

# Check application logs
kubectl logs deployment/openwan-api | grep "file-id=<failing-file-id>"
```

### Issue: Users Cannot Login

```bash
# Check user exists in new database
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT id, username, email FROM ow_users WHERE username = '<username>';
"

# Check password hash format
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT LENGTH(password) as hash_length FROM ow_users WHERE username = '<username>';
"
# bcrypt hashes are 60 characters

# Check Redis session storage
redis-cli -h <redis-endpoint> KEYS "session:*" | wc -l

# Reset user password
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  UPDATE ow_users SET password_reset_required = 1 WHERE username = '<username>';
"
```

### Issue: Search Not Working

```bash
# Check Sphinx status
curl http://<sphinx-host>:9312/

# Check if fallback to database search is working
kubectl logs deployment/openwan-api | grep -i "search.*database"

# Rebuild Sphinx indexes
ssh sphinx-server "indexer --all --rotate"

# Verify search in database directly
mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME -e "
  SELECT id, title FROM ow_files WHERE title LIKE '%test%' LIMIT 10;
"
```

## Appendix

### A. Migration Timeline Template

```
OpenWan Migration Timeline

Week 1-2: Preparation
  [ ] Complete pre-migration checklist
  [ ] Take full backups
  [ ] Document legacy system
  [ ] Verify new system readiness
  [ ] Create detailed migration scripts
  [ ] Dry run on staging

Week 3: Staging Migration
  [ ] Migrate data to staging
  [ ] Sync files to staging
  [ ] Deploy new system to staging
  [ ] Initial testing

Week 4: Testing & Validation
  [ ] Run automated test suite
  [ ] Conduct UAT with key users
  [ ] Performance testing
  [ ] Fix any identified issues
  [ ] Dry run cutover procedure

Week 5: Production Cutover
  [ ] Send user communications
  [ ] Schedule maintenance window
  [ ] Execute cutover plan
  [ ] Monitor intensively
  [ ] Declare success or rollback

Week 6+: Post-Migration
  [ ] Daily monitoring
  [ ] Weekly validation
  [ ] Gather user feedback
  [ ] Address any issues
  [ ] Plan legacy decommission
```

### B. Success Criteria

- [ ] All data migrated (100% record count match)
- [ ] All files accessible (100% file count match)
- [ ] All users can login
- [ ] All core functionality works (upload, download, search, catalog)
- [ ] Performance meets SLAs (p95 < 500ms)
- [ ] No critical bugs reported in first week
- [ ] User satisfaction > 80% (post-migration survey)
- [ ] Zero data loss
- [ ] Downtime within SLA (<1 hour)

### C. Key Contacts

| Role | Name | Email | Phone |
|------|------|-------|-------|
| Migration Lead | | | |
| Database Admin | | | |
| Application Lead | | | |
| Infrastructure Lead | | | |
| QA Lead | | | |
| Product Owner | | | |

---

**Document Version**: 1.0  
**Last Updated**: 2024-02-01  
**Maintained By**: OpenWan Migration Team
