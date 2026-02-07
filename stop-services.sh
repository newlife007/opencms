#!/bin/bash
#
# OpenWan 服务停止脚本
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  OpenWan 服务停止脚本${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# 停止服务
stopped=0

# 停止 API
if [ -f logs/api.pid ]; then
    API_PID=$(cat logs/api.pid)
    if ps -p $API_PID > /dev/null 2>&1; then
        echo -e "停止后端 API 服务 (PID: $API_PID)..."
        kill $API_PID
        sleep 1
        if ps -p $API_PID > /dev/null 2>&1; then
            kill -9 $API_PID 2>/dev/null || true
        fi
        echo -e "  ${GREEN}✓ 后端 API 已停止${NC}"
        stopped=$((stopped + 1))
    fi
    rm -f logs/api.pid
fi

# 停止 Worker #1
if [ -f logs/worker-1.pid ]; then
    WORKER1_PID=$(cat logs/worker-1.pid)
    if ps -p $WORKER1_PID > /dev/null 2>&1; then
        echo -e "停止 Worker #1 (PID: $WORKER1_PID)..."
        kill $WORKER1_PID
        sleep 1
        if ps -p $WORKER1_PID > /dev/null 2>&1; then
            kill -9 $WORKER1_PID 2>/dev/null || true
        fi
        echo -e "  ${GREEN}✓ Worker #1 已停止${NC}"
        stopped=$((stopped + 1))
    fi
    rm -f logs/worker-1.pid
fi

# 停止 Worker #2
if [ -f logs/worker-2.pid ]; then
    WORKER2_PID=$(cat logs/worker-2.pid)
    if ps -p $WORKER2_PID > /dev/null 2>&1; then
        echo -e "停止 Worker #2 (PID: $WORKER2_PID)..."
        kill $WORKER2_PID
        sleep 1
        if ps -p $WORKER2_PID > /dev/null 2>&1; then
            kill -9 $WORKER2_PID 2>/dev/null || true
        fi
        echo -e "  ${GREEN}✓ Worker #2 已停止${NC}"
        stopped=$((stopped + 1))
    fi
    rm -f logs/worker-2.pid
fi

# 清理其他可能的进程
pkill -f "bin/openwan" 2>/dev/null || true

if [ $stopped -eq 0 ]; then
    echo -e "${YELLOW}没有发现运行中的服务${NC}"
else
    echo -e "\n${GREEN}✓ 已停止 $stopped 个服务${NC}"
fi

echo ""
