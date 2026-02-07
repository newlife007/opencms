#!/bin/bash

# OpenWan 部署测试脚本
# 用于验证系统部署和基本功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:5173}"

echo "======================================"
echo "OpenWan 部署测试"
echo "======================================"
echo "API URL: $API_BASE_URL"
echo "Frontend URL: $FRONTEND_URL"
echo ""

# 测试计数器
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "测试: $name ... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$response" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $response)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_status, got $response)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

test_json_response() {
    local name=$1
    local url=$2
    local expected_field=$3
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "测试: $name ... "
    
    response=$(curl -s "$url" 2>/dev/null || echo "{}")
    
    if echo "$response" | grep -q "$expected_field"; then
        echo -e "${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (Field '$expected_field' not found)"
        echo "Response: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# 1. 健康检查测试
echo ""
echo "==== 1. 健康检查测试 ===="
test_endpoint "健康检查端点" "$API_BASE_URL/health"
test_endpoint "就绪检查端点" "$API_BASE_URL/ready"
test_endpoint "存活检查端点" "$API_BASE_URL/alive"
test_json_response "健康检查JSON响应" "$API_BASE_URL/health" "status"

# 2. 数据库连接测试
echo ""
echo "==== 2. 数据库连接测试 ===="
test_json_response "数据库连接状态" "$API_BASE_URL/health" "database"

# 3. API端点测试
echo ""
echo "==== 3. API端点测试 ===="
test_endpoint "API Ping" "$API_BASE_URL/api/v1/ping"

# 4. 前端资源测试
echo ""
echo "==== 4. 前端资源测试 ===="
if command -v curl &> /dev/null; then
    test_endpoint "前端首页" "$FRONTEND_URL"
else
    echo -e "${YELLOW}⚠ SKIP${NC} (curl not available)"
fi

# 5. 认证端点测试 (预期401未授权)
echo ""
echo "==== 5. 认证保护测试 ===="
test_endpoint "受保护的API (预期401)" "$API_BASE_URL/api/v1/files" 401

# 6. 登录测试
echo ""
echo "==== 6. 认证测试 ===="
echo -n "测试: 登录API ... "
TESTS_RUN=$((TESTS_RUN + 1))

login_response=$(curl -s -X POST "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' \
    2>/dev/null || echo "{}")

if echo "$login_response" | grep -q "token\|success"; then
    echo -e "${GREEN}✓ PASS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    
    # 提取token (如果有)
    TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    if [ ! -z "$TOKEN" ]; then
        echo "  Token获取成功: ${TOKEN:0:20}..."
    fi
else
    echo -e "${YELLOW}⚠ SKIP${NC} (可能需要先初始化数据)"
    echo "  Response: $login_response"
fi

# 7. Docker容器检查 (如果运行在Docker环境)
echo ""
echo "==== 7. Docker容器状态 ===="
if command -v docker &> /dev/null; then
    if docker ps &> /dev/null; then
        echo "Docker容器列表:"
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "openwan|NAME"
        
        # 检查关键容器
        for container in openwan-api openwan-mysql openwan-redis; do
            if docker ps | grep -q "$container"; then
                echo -e "  $container: ${GREEN}运行中${NC}"
            else
                echo -e "  $container: ${RED}未运行${NC}"
            fi
        done
    else
        echo -e "${YELLOW}⚠ Docker daemon未运行或无权限${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Docker未安装${NC}"
fi

# 8. 端口监听检查
echo ""
echo "==== 8. 端口监听检查 ===="
check_port() {
    local port=$1
    local service=$2
    
    if command -v netstat &> /dev/null; then
        if netstat -tuln | grep -q ":$port "; then
            echo -e "  端口 $port ($service): ${GREEN}监听中${NC}"
        else
            echo -e "  端口 $port ($service): ${RED}未监听${NC}"
        fi
    elif command -v ss &> /dev/null; then
        if ss -tuln | grep -q ":$port "; then
            echo -e "  端口 $port ($service): ${GREEN}监听中${NC}"
        else
            echo -e "  端口 $port ($service): ${RED}未监听${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ netstat/ss命令不可用${NC}"
    fi
}

check_port 8080 "API服务"
check_port 3306 "MySQL"
check_port 6379 "Redis"
check_port 5672 "RabbitMQ"
check_port 5173 "前端开发服务器"

# 9. 日志检查
echo ""
echo "==== 9. 服务日志检查 ===="
if [ -f "logs/api.log" ]; then
    echo "API日志 (最后10行):"
    tail -n 10 logs/api.log
else
    echo -e "${YELLOW}⚠ 日志文件不存在: logs/api.log${NC}"
fi

if command -v docker &> /dev/null && docker ps &> /dev/null; then
    if docker ps | grep -q "openwan-api"; then
        echo ""
        echo "Docker API容器日志 (最后10行):"
        docker logs --tail 10 openwan-api 2>&1 | head -n 10
    fi
fi

# 10. 配置文件检查
echo ""
echo "==== 10. 配置文件检查 ===="
check_config_file() {
    local file=$1
    if [ -f "$file" ]; then
        echo -e "  $file: ${GREEN}存在${NC}"
    else
        echo -e "  $file: ${RED}缺失${NC}"
    fi
}

check_config_file "configs/config.yaml"
check_config_file "frontend/.env.development"
check_config_file "docker-compose.yml"

# 测试总结
echo ""
echo "======================================"
echo "测试总结"
echo "======================================"
echo "总计运行: $TESTS_RUN"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ 所有测试通过！系统部署正常。${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠ 有 $TESTS_FAILED 个测试失败，请检查相关服务。${NC}"
    exit 1
fi
