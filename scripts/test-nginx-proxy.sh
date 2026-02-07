#!/bin/bash

# OpenWan Nginx 代理测试脚本
# ========================================

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================"
echo "OpenWan Nginx 代理测试"
echo "======================================"
echo ""

# 获取IP
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "localhost")
echo "测试服务器: $PUBLIC_IP"
echo ""

# 测试计数
PASS=0
FAIL=0

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    echo -n "测试 $name ... "
    
    if curl -s --connect-timeout 3 "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAIL++))
    fi
}

# Nginx本地测试
echo -e "${BLUE}[本地Nginx测试]${NC}"
test_endpoint "健康检查 (/health)" "http://localhost/health"
test_endpoint "API Ping (/api/v1/ping)" "http://localhost/api/v1/ping"
test_endpoint "前端首页 (/)" "http://localhost/"
echo ""

# 外网Nginx测试
if [ "$PUBLIC_IP" != "localhost" ]; then
    echo -e "${BLUE}[外网Nginx测试]${NC}"
    test_endpoint "外网健康检查" "http://$PUBLIC_IP/health"
    test_endpoint "外网API" "http://$PUBLIC_IP/api/v1/ping"
    test_endpoint "外网前端" "http://$PUBLIC_IP/"
    echo ""
fi

# 检查Nginx状态
echo -e "${BLUE}[Nginx服务状态]${NC}"
if systemctl is-active --quiet nginx; then
    echo -e "Nginx服务: ${GREEN}运行中${NC}"
else
    echo -e "Nginx服务: ${RED}未运行${NC}"
fi
echo ""

# 检查端口
echo -e "${BLUE}[端口监听状态]${NC}"
netstat -tuln | grep -E ":(80|3000|8080)" | while read line; do
    if echo "$line" | grep -q ":80 "; then
        echo -e "Nginx (80): ${GREEN}监听中${NC}"
    elif echo "$line" | grep -q ":3000"; then
        echo -e "前端 (3000): ${GREEN}监听中${NC}"
    elif echo "$line" | grep -q ":8080"; then
        echo -e "API (8080): ${GREEN}监听中${NC}"
    fi
done
echo ""

# 显示访问地址
echo "======================================"
echo -e "${BLUE}访问地址${NC}"
echo "======================================"
echo ""
if [ "$PUBLIC_IP" != "localhost" ]; then
    echo "【通过Nginx访问（推荐）】"
    echo "前端: http://$PUBLIC_IP/"
    echo "API:  http://$PUBLIC_IP/api/v1/xxx"
    echo ""
    echo "【直接访问（备用）】"
    echo "前端: http://$PUBLIC_IP:3000"
    echo "API:  http://$PUBLIC_IP:8080"
    echo ""
fi

# 测试结果
echo "======================================"
echo -e "${BLUE}测试结果${NC}"
echo "======================================"
echo ""
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    echo ""
    echo "现在可以访问："
    echo "http://$PUBLIC_IP/"
else
    echo -e "${YELLOW}⚠ 部分测试失败${NC}"
    echo ""
    echo "请检查："
    echo "1. Nginx服务是否运行"
    echo "2. 前端和API服务是否启动"
    echo "3. AWS安全组是否开放端口80"
fi

echo ""
echo "======================================"
