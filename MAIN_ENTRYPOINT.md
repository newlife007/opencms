# OpenWan Main Entry Point

## Main Application File

**The official main entry point is: `./main.go` in the project root directory.**

This file provides:
- ✅ Complete initialization of all services
- ✅ Graceful shutdown with signal handling (SIGTERM, SIGINT)
- ✅ Session management with Redis
- ✅ Environment variable configuration
- ✅ Comprehensive startup logging
- ✅ Error handling and cleanup

## Usage

### Running Directly

```bash
# With defaults (localhost database and Redis)
./main.go

# With environment variables
DB_HOST=192.168.1.100 \
DB_USER=openwan \
DB_PASSWORD=secret123 \
REDIS_ADDR=192.168.1.101:6379 \
LOCAL_STORAGE_PATH=./data \
PORT=8080 \
./main.go
```

### Build and Run

```bash
# Build
go build -o bin/openwan ./main.go

# Run with environment variables
LOCAL_STORAGE_PATH=./data ./bin/openwan

# Or with all environment variables from .env file
export $(cat .env | xargs) && ./bin/openwan
```

### Docker

The Dockerfile has been updated to use `./main.go`:

```bash
docker build -t openwan-api .
docker run -p 8080:8080 \
  -e DB_HOST=mysql-host \
  -e REDIS_ADDR=redis-host:6379 \
  -e LOCAL_STORAGE_PATH=/app/data \
  openwan-api
```

## Environment Variables

The main.go supports these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DB_HOST` | `127.0.0.1` | MySQL database host |
| `DB_USER` | `openwan` | Database username |
| `DB_PASSWORD` | `openwan123` | Database password |
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `STORAGE_TYPE` | `local` | Storage type (local/s3) |
| `LOCAL_STORAGE_PATH` | `./data` | Local storage path |
| `S3_BUCKET` | - | S3 bucket name (if using S3) |
| `S3_REGION` | - | AWS region |
| `S3_ACCESS_KEY_ID` | - | AWS access key |
| `S3_SECRET_ACCESS_KEY` | - | AWS secret key |
| `ALLOWED_ORIGINS` | - | Additional CORS origins |

See `.env.example` for complete list.

## Old Entry Points (Deprecated)

- ❌ `./cmd/api/main_db.go` - Backed up to `main_db.go.backup`
  - Missing graceful shutdown
  - Less comprehensive startup logging
  - No signal handling

**Do not use the old entry points. Use `./main.go` only.**

## Service Status

Check service is running:

```bash
ps aux | grep openwan
curl http://localhost:8080/health
```

Stop service (graceful):

```bash
# Send SIGTERM for graceful shutdown
pkill -TERM -f "bin/openwan"

# Or force kill
pkill -9 -f "bin/openwan"
```

## Startup Output Example

```
========================================
OpenWan Media Asset Management System
API Server with Session Authentication
Version: 1.0.0
Go Version: 1.25.5
========================================

Initializing database connection...
✓ Database connected
Initializing Redis session store...
✓ Redis session store connected
Initializing storage service...
✓ Storage service initialized
Initializing repositories...
✓ Repositories initialized
Initializing services...
✓ Services initialized
✓ Router configured

========================================
Server starting on :8080
Health check: http://localhost:8080/health
API endpoint: http://localhost:8080/api/v1/ping
Database: openwan@127.0.0.1:3306/openwan_db
Redis: localhost:6379
Storage: local
Press Ctrl+C to stop
========================================

Server started on :8080
```

## Integration with Frontend

The frontend (Vue.js) should be configured to use the API at:

- Development: `http://localhost:8080/api/v1`
- Production: `/api/v1` (relative URL when behind same domain/proxy)

The API server includes CORS support for:
- `http://localhost:3000` (frontend dev server)
- Custom origins via `ALLOWED_ORIGINS` environment variable

## Health Checks

The main.go provides health check endpoints:

- `GET /health` - Full health check (database, Redis, storage)
- `GET /ready` - Readiness probe (for Kubernetes)
- `GET /alive` - Liveness probe (for Kubernetes)

## Graceful Shutdown

The main.go handles shutdown signals properly:

1. Receives SIGTERM or SIGINT (Ctrl+C)
2. Stops accepting new requests
3. Waits up to 5 seconds for in-flight requests to complete
4. Closes Redis connections
5. Closes database connections
6. Exits cleanly

This ensures no data loss or connection leaks during deployment or restart.
