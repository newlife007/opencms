#!/bin/bash
################################################################################
# OpenWan - 启动脚本
################################################################################
set -e
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}启动OpenWan服务...${NC}"
docker-compose up -d

echo ""
echo -e "${GREEN}✓ 服务启动成功！${NC}"
echo ""
echo "访问地址:"
echo "  前端: http://localhost:3000"
echo "  后端: http://localhost:8080"
echo ""
echo "查看日志: docker-compose logs -f"
