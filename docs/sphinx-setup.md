# Sphinx Full-Text Search Setup Guide

## Overview

This guide provides comprehensive instructions for setting up and configuring Sphinx search engine for the OpenWan Media Asset Management System. Sphinx provides high-performance full-text search with Chinese language support.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Index Management](#index-management)
5. [Integration with Go Backend](#integration-with-go-backend)
6. [Maintenance](#maintenance)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

- MySQL database with OpenWan schema installed
- Linux server with at least 2GB RAM
- Root or sudo access for installation
- Go 1.25+ for indexer tool

## Installation

### Option 1: Install from Package Repository (Ubuntu/Debian)

```bash
# Add Sphinx repository
wget -qO - https://sphinxsearch.com/files/gpg.key | sudo apt-key add -
echo "deb http://sphinxsearch.com/repo/ubuntu/ focal main" | sudo tee /etc/apt/sources.list.d/sphinx.list

# Update and install
sudo apt-get update
sudo apt-get install sphinxsearch

# Verify installation
searchd --version
indexer --version
```

### Option 2: Install from Source

```bash
# Install dependencies
sudo apt-get install build-essential libmysqlclient-dev libexpat1-dev

# Download and compile Sphinx
cd /tmp
wget http://sphinxsearch.com/files/sphinx-3.4.1-bin.tar.gz
tar -xzf sphinx-3.4.1-bin.tar.gz
cd sphinx-3.4.1

./configure --prefix=/usr/local/sphinx --with-mysql
make
sudo make install
```

### Option 3: CoreSeek (Chinese-optimized Sphinx)

For better Chinese word segmentation, consider CoreSeek:

```bash
# Download CoreSeek
wget http://www.coreseek.cn/uploads/csft/4.0/coreseek-4.1-beta.tar.gz
tar -xzf coreseek-4.1-beta.tar.gz
cd coreseek-4.1-beta

# Compile mmseg (Chinese tokenizer)
cd mmseg-3.2.14
./bootstrap
./configure --prefix=/usr/local/mmseg3
make
sudo make install

# Compile CoreSeek
cd ../csft-4.1
./configure --prefix=/usr/local/coreseek --with-mysql --with-mmseg=/usr/local/mmseg3
make
sudo make install
```

## Configuration

### 1. Copy Configuration File

```bash
# Copy Sphinx configuration to system location
sudo cp /home/ec2-user/openwan/configs/sphinx.conf /etc/sphinx/sphinx.conf

# Or use custom location
export SPHINX_CONF=/home/ec2-user/openwan/configs/sphinx.conf
```

### 2. Update Database Credentials

Edit `/etc/sphinx/sphinx.conf`:

```conf
source main
{
    sql_host    = localhost      # Your MySQL host
    sql_user    = root           # MySQL user
    sql_pass    = your_password  # MySQL password
    sql_db      = openwan_db     # Database name
    sql_port    = 3306           # MySQL port
}
```

### 3. Create Data Directories

```bash
# Create Sphinx data directories
sudo mkdir -p /var/data/sphinx/openwan
sudo mkdir -p /var/log/sphinx
sudo mkdir -p /var/run/sphinx

# Set permissions
sudo chown -R sphinx:sphinx /var/data/sphinx
sudo chown -R sphinx:sphinx /var/log/sphinx
sudo chown -R sphinx:sphinx /var/run/sphinx
```

### 4. Update Paths in Configuration

Edit the `path` directives in sphinx.conf to match your system:

```conf
index main
{
    path = /var/data/sphinx/openwan/main  # Update this path
}

index delta
{
    path = /var/data/sphinx/openwan/delta  # Update this path
}
```

## Index Management

### Initial Index Build

Build all indexes for the first time:

```bash
# Build all indexes
sudo indexer --config /etc/sphinx/sphinx.conf --all

# Or use the custom indexer tool
cd /home/ec2-user/openwan
go run cmd/indexer/main.go --config /etc/sphinx/sphinx.conf --all
```

### Delta Index Updates (Incremental)

For frequent updates without rebuilding the entire index:

```bash
# Update delta index (runs quickly)
indexer --config /etc/sphinx/sphinx.conf --rotate delta

# Or use the indexer tool
go run cmd/indexer/main.go --config /etc/sphinx/sphinx.conf --delta
```

### Merge Delta into Main

Periodically merge the delta index into the main index:

```bash
# Merge delta into main (do this daily or weekly)
indexer --config /etc/sphinx/sphinx.conf --merge main delta --rotate
```

### Set Up Cron Jobs

Automate index updates with cron:

```bash
# Edit crontab
sudo crontab -e

# Add these lines:
# Update delta index every 10 minutes
*/10 * * * * /usr/bin/indexer --config /etc/sphinx/sphinx.conf --rotate delta >> /var/log/sphinx/indexer.log 2>&1

# Merge delta into main index daily at 2 AM
0 2 * * * /usr/bin/indexer --config /etc/sphinx/sphinx.conf --merge main delta --rotate >> /var/log/sphinx/indexer.log 2>&1

# Rebuild all indexes weekly on Sunday at 3 AM
0 3 * * 0 /usr/bin/indexer --config /etc/sphinx/sphinx.conf --all --rotate >> /var/log/sphinx/indexer.log 2>&1
```

## Starting Sphinx Daemon

### Start searchd Manually

```bash
# Start Sphinx search daemon
sudo searchd --config /etc/sphinx/sphinx.conf

# Check if running
ps aux | grep searchd
```

### Create systemd Service

Create `/etc/systemd/system/sphinx.service`:

```ini
[Unit]
Description=Sphinx Search Daemon
After=network.target mysql.service

[Service]
Type=forking
User=sphinx
Group=sphinx
ExecStart=/usr/bin/searchd --config /etc/sphinx/sphinx.conf
ExecStop=/usr/bin/searchd --config /etc/sphinx/sphinx.conf --stopwait
PIDFile=/var/run/sphinx/searchd.pid
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable sphinx
sudo systemctl start sphinx
sudo systemctl status sphinx
```

## Integration with Go Backend

### 1. Configure Go Application

Update your `.env` or `configs/config.yaml`:

```yaml
sphinx:
  host: localhost
  port: 9306                    # SphinxQL port
  main_index: main              # Main index name
  delta_index: delta            # Delta index name
  timeout: 5s
```

### 2. Connection String

The Go application connects to Sphinx via MySQL protocol:

```go
// Connection string format
sphinxDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
    "",              // No username for Sphinx
    "",              // No password for Sphinx
    "localhost",     // Sphinx host
    9306,            // SphinxQL port (not MySQL port!)
)
```

### 3. Test Connection

```bash
# Test SphinxQL connection
mysql -h localhost -P 9306

# In MySQL prompt, test query:
SELECT * FROM main LIMIT 10;
SHOW INDEX main STATUS;
```

## Maintenance

### Monitor Index Status

Check index statistics:

```bash
# Using indexer tool
go run cmd/indexer/main.go --status

# Using mysql client
mysql -h localhost -P 9306 -e "SHOW INDEX main STATUS"
```

### Optimize Indexes

Periodically optimize indexes for better performance:

```bash
# Optimize main index
mysql -h localhost -P 9306 -e "OPTIMIZE INDEX main"
```

### Backup Indexes

```bash
# Stop searchd
sudo systemctl stop sphinx

# Backup index files
sudo tar -czf sphinx-backup-$(date +%Y%m%d).tar.gz /var/data/sphinx/

# Restart searchd
sudo systemctl start sphinx
```

### Log Rotation

Configure log rotation in `/etc/logrotate.d/sphinx`:

```
/var/log/sphinx/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    missingok
    postrotate
        systemctl reload sphinx > /dev/null 2>&1 || true
    endscript
}
```

## Performance Tuning

### Memory Settings

Adjust in `sphinx.conf`:

```conf
indexer
{
    mem_limit = 512M     # Increase for faster indexing
}

searchd
{
    max_matches = 10000  # Maximum results to return
    qcache_max_bytes = 32M  # Query cache size
}
```

### MySQL Configuration

Optimize MySQL for Sphinx indexing:

```ini
[mysqld]
# Increase query cache
query_cache_size = 32M
query_cache_type = 1

# Increase buffer pool
innodb_buffer_pool_size = 1G
```

## Troubleshooting

### Sphinx Not Starting

```bash
# Check logs
sudo tail -f /var/log/sphinx/searchd.log

# Common issues:
# 1. Port already in use
sudo netstat -tulpn | grep 9306

# 2. Permission issues
sudo chown -R sphinx:sphinx /var/data/sphinx /var/log/sphinx /var/run/sphinx

# 3. Config syntax error
indexer --config /etc/sphinx/sphinx.conf --parse-only
```

### Search Not Returning Results

```bash
# 1. Check if indexes exist
ls -lh /var/data/sphinx/openwan/

# 2. Rebuild indexes
sudo indexer --config /etc/sphinx/sphinx.conf --all --rotate

# 3. Check index status
mysql -h localhost -P 9306 -e "SHOW INDEX main STATUS"

# 4. Test simple query
mysql -h localhost -P 9306 -e "SELECT * FROM main WHERE MATCH('test') LIMIT 10"
```

### Indexer Fails

```bash
# Check indexer log
sudo tail -f /var/log/sphinx/indexer.log

# Common issues:
# 1. MySQL connection failed - check credentials
# 2. Out of memory - increase mem_limit in config
# 3. Disk space - check available space
df -h /var/data/sphinx
```

### Chinese Text Not Tokenized

If using standard Sphinx (not CoreSeek):

```conf
# Enable NGram for Chinese in sphinx.conf
index main
{
    ngram_len = 1
    ngram_chars = U+3000..U+2FA1F
}
```

Or switch to CoreSeek for better Chinese support.

## API Endpoints

The Go backend exposes these search endpoints:

### POST /api/v1/search

Search for files with filters:

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "q": "视频",
    "type": 1,
    "status": 2,
    "page": 1,
    "page_size": 20,
    "sort_by": "relevance"
  }'
```

### GET /api/v1/admin/search/status

Get index status (admin only):

```bash
curl http://localhost:8080/api/v1/admin/search/status \
  -H "Authorization: Bearer <token>"
```

### POST /api/v1/admin/search/reindex

Trigger index rebuild (admin only):

```bash
curl -X POST http://localhost:8080/api/v1/admin/search/reindex \
  -H "Authorization: Bearer <token>"
```

## Best Practices

1. **Incremental Updates**: Use delta indexing for frequent updates instead of rebuilding all indexes
2. **Regular Merging**: Merge delta into main index daily to keep main index up-to-date
3. **Monitor Performance**: Track query times and adjust configuration as needed
4. **Backup Regularly**: Backup index files before major updates
5. **Use Distributed Index**: The `openwan` distributed index searches both main and delta automatically
6. **Query Cache**: Enable query cache for frequently repeated searches
7. **Access Control**: Ensure search results respect user permissions (level and group access)

## References

- [Sphinx Official Documentation](http://sphinxsearch.com/docs/)
- [SphinxQL Reference](http://sphinxsearch.com/docs/current/sphinxql-reference.html)
- [CoreSeek Documentation](http://www.coreseek.cn/) (Chinese)
- [OpenWan Database Schema](./database-schema.md)

## Support

For issues specific to OpenWan integration:
- Check application logs: `/home/ec2-user/openwan/logs/`
- Review search service code: `internal/service/search_service.go`
- Verify Sphinx connection in config: `configs/config.yaml`
