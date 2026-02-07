#!/bin/bash

# OpenWan 测试环境快速操作指南
# ========================================

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "======================================"
echo "OpenWan 快速操作指南"
echo "======================================"
echo ""

# 检查服务状态
echo -e "${BLUE}[服务状态]${NC}"
echo ""

# 检查API
API_PID=$(cat /tmp/openwan_api.pid 2>/dev/null)
if ps -p $API_PID > /dev/null 2>&1; then
    echo -e "API服务: ${GREEN}运行中${NC} (PID: $API_PID)"
    echo "  URL: http://localhost:8080"
    echo "  健康检查: curl http://localhost:8080/health"
else
    echo -e "API服务: ${RED}未运行${NC}"
    echo "  启动: cd /home/ec2-user/openwan && ./bin/api > logs/api.log 2>&1 &"
fi
echo ""

# 检查前端
FRONTEND_PID=$(cat /tmp/openwan_frontend.pid 2>/dev/null)
if ps -p $FRONTEND_PID > /dev/null 2>&1; then
    echo -e "前端服务: ${GREEN}运行中${NC} (PID: $FRONTEND_PID)"
    echo "  URL: http://localhost:3000"
else
    echo -e "前端服务: ${RED}未运行${NC}"
    echo "  启动: cd /home/ec2-user/openwan/frontend && npm run dev > ../logs/frontend.log 2>&1 &"
fi
echo ""

# 可用命令
echo -e "${BLUE}[可用命令]${NC}"
echo ""
cat << 'EOF'
1. 查看API日志（实时）:
   tail -f logs/api.log

2. 查看前端日志（实时）:
   tail -f logs/frontend.log

3. 测试API健康:
   curl http://localhost:8080/health | jq .

4. 测试登录:
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin123"}' | jq .

5. 获取文件列表:
   curl http://localhost:8080/api/v1/files | jq .

6. 运行部署测试:
   ./scripts/test-deployment.sh

7. 运行功能测试:
   ./scripts/test-functionality.sh

8. 重启API服务:
   kill $(cat /tmp/openwan_api.pid)
   ./bin/api > logs/api.log 2>&1 &
   echo $! > /tmp/openwan_api.pid

9. 重启前端服务:
   kill $(cat /tmp/openwan_frontend.pid)
   cd frontend && npm run dev > ../logs/frontend.log 2>&1 &
   echo $! > /tmp/openwan_frontend.pid
   cd ..

10. 停止所有服务:
    kill $(cat /tmp/openwan_api.pid) 2>/dev/null
    kill $(cat /tmp/openwan_frontend.pid) 2>/dev/null
EOF

echo ""
echo "======================================"
echo -e "${YELLOW}提示: 使用 -h 参数查看详细帮助${NC}"
echo "======================================"
