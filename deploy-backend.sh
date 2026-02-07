#!/bin/bash
# OpenWan Backend Deployment Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================="
echo "OpenWan Backend Deployment"
echo "========================================="
echo ""

# Step 1: Stop existing service
echo "Step 1: Stopping existing service..."
pkill -f "bin/openwan" 2>/dev/null || echo "  No existing service found"
sleep 2

# Step 2: Build backend
echo ""
echo "Step 2: Building backend..."
go build -o bin/openwan ./cmd/api
if [ $? -eq 0 ]; then
    echo "  ✓ Build successful"
else
    echo "  ✗ Build failed"
    exit 1
fi

# Step 3: Create logs directory
echo ""
echo "Step 3: Preparing logs..."
mkdir -p logs
touch logs/api.log

# Step 4: Start service
echo ""
echo "Step 4: Starting backend service..."
nohup ./bin/openwan > logs/api.log 2>&1 &
BACKEND_PID=$!
echo "  Backend PID: $BACKEND_PID"

# Step 5: Wait for service to start
echo ""
echo "Step 5: Waiting for service to start..."
sleep 3

# Step 6: Check service status
echo ""
echo "Step 6: Checking service status..."
if ps -p $BACKEND_PID > /dev/null; then
    echo "  ✓ Service is running (PID: $BACKEND_PID)"
else
    echo "  ✗ Service failed to start"
    echo ""
    echo "Last 20 lines of log:"
    tail -20 logs/api.log
    exit 1
fi

# Step 7: Test health endpoint
echo ""
echo "Step 7: Testing health endpoint..."
sleep 2
curl -s http://localhost:8080/health | grep -q "openwan-api" && echo "  ✓ Health endpoint responding" || echo "  ⚠ Health endpoint not responding (service may still be initializing)"

echo ""
echo "========================================="
echo "Deployment Complete!"
echo "========================================="
echo ""
echo "Service Information:"
echo "  PID: $BACKEND_PID"
echo "  Port: 8080"
echo "  Health: http://localhost:8080/health"
echo "  API: http://localhost:8080/api/v1/"
echo "  Logs: tail -f logs/api.log"
echo ""
echo "To stop: pkill -f 'bin/openwan'"
echo ""
