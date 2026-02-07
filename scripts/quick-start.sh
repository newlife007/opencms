#!/bin/bash

# Quick Start Script for Testing
# 快速启动脚本用于测试

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "======================================"
echo "OpenWan 快速启动 (测试模式)"
echo "======================================"
echo ""

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: Go未安装${NC}"
    echo "请先安装Go 1.25+"
    exit 1
fi

echo -e "${BLUE}[INFO]${NC} Go版本: $(go version)"

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    echo -e "${RED}错误: 配置文件不存在: configs/config.yaml${NC}"
    exit 1
fi

echo -e "${GREEN}[OK]${NC} 配置文件存在"

# 创建必要的目录
mkdir -p storage logs tmp

# 编译API服务
echo ""
echo -e "${BLUE}[BUILD]${NC} 编译API服务..."
go build -o bin/api cmd/api/main.go

if [ $? -eq 0 ]; then
    echo -e "${GREEN}[OK]${NC} API服务编译成功"
else
    echo -e "${RED}[FAIL]${NC} API服务编译失败"
    exit 1
fi

# 启动API服务
echo ""
echo -e "${BLUE}[START]${NC} 启动API服务..."
echo -e "${YELLOW}提示: 按 Ctrl+C 停止服务${NC}"
echo ""

# 设置环境变量
export GIN_MODE=debug

# 启动服务
./bin/api

# 清理
cleanup() {
    echo ""
    echo -e "${BLUE}[STOP]${NC} 停止服务..."
    pkill -f "./bin/api" 2>/dev/null || true
}

trap cleanup EXIT INT TERM
