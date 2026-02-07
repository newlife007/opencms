#!/bin/bash

# 简化版测试报告
# 检查当前系统状态并生成报告

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "======================================"
echo "OpenWan 部署状态报告"
echo "======================================"
echo "生成时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 1. 环境检查
echo "==== 1. 环境信息 ===="
echo -n "Go版本: "
go version 2>/dev/null | awk '{print $3}' || echo -e "${RED}未安装${NC}"

echo -n "Node.js版本: "
node --version 2>/dev/null || echo -e "${RED}未安装${NC}"

echo -n "MySQL状态: "
if command -v mysql &> /dev/null; then
    mysql --version 2>/dev/null | awk '{print $5}' | tr -d ','
else
    echo -e "${RED}未安装${NC}"
fi

echo -n "Redis状态: "
if command -v redis-cli &> /dev/null; then
    echo -e "${GREEN}已安装${NC}"
else
    echo -e "${YELLOW}未安装 (可选)${NC}"
fi

# 2. 项目文件
echo ""
echo "==== 2. 项目文件 ===="
echo -n "后端代码: "
if [ -d "cmd" ] && [ -d "internal" ]; then
    go_files=$(find cmd internal -name "*.go" | wc -l)
    echo -e "${GREEN}✓${NC} ($go_files 个Go文件)"
else
    echo -e "${RED}✗${NC} 目录不完整"
fi

echo -n "前端代码: "
if [ -d "frontend/src" ]; then
    vue_files=$(find frontend/src -name "*.vue" 2>/dev/null | wc -l)
    echo -e "${GREEN}✓${NC} ($vue_files 个Vue文件)"
else
    echo -e "${RED}✗${NC} 目录不存在"
fi

echo -n "配置文件: "
if [ -f "configs/config.yaml" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
fi

# 3. 依赖状态
echo ""
echo "==== 3. 依赖状态 ===="
echo -n "Go模块: "
if [ -f "go.mod" ] && [ -f "go.sum" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${YELLOW}⚠${NC} 不完整"
fi

echo -n "前端依赖: "
if [ -d "frontend/node_modules" ]; then
    echo -e "${GREEN}✓ 已安装${NC}"
else
    echo -e "${YELLOW}⚠ 未安装 (需要npm install)${NC}"
fi

# 4. 目录权限
echo ""
echo "==== 4. 目录权限 ===="
for dir in storage logs tmp; do
    echo -n "$dir: "
    if [ -d "$dir" ]; then
        if [ -w "$dir" ]; then
            echo -e "${GREEN}✓ 可写${NC}"
        else
            echo -e "${YELLOW}⚠ 只读${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ 不存在${NC}"
        mkdir -p "$dir" 2>/dev/null && echo -e "  ${GREEN}已创建${NC}"
    fi
done

# 5. 端口检查
echo ""
echo "==== 5. 端口状态 ===="
check_port() {
    port=$1
    name=$2
    echo -n "$name (端口 $port): "
    if ss -tuln 2>/dev/null | grep -q ":$port " || netstat -tuln 2>/dev/null | grep -q ":$port "; then
        echo -e "${YELLOW}已占用${NC}"
    else
        echo -e "${GREEN}可用${NC}"
    fi
}

check_port 8080 "API服务"
check_port 5173 "前端(Vite)"
check_port 3306 "MySQL"
check_port 6379 "Redis"

# 6. 数据库连接测试
echo ""
echo "==== 6. 数据库测试 ===="
echo -n "连接测试: "
if mysql -h localhost -e "SELECT 1" &> /dev/null; then
    echo -e "${GREEN}✓ 成功${NC}"
    
    echo -n "openwan_db数据库: "
    if mysql -h localhost -e "USE openwan_db" &> /dev/null; then
        echo -e "${GREEN}✓ 存在${NC}"
        
        table_count=$(mysql -h localhost openwan_db -e "SHOW TABLES" 2>/dev/null | wc -l)
        echo "  表数量: $((table_count - 1))"
    else
        echo -e "${YELLOW}⚠ 不存在 (需要创建)${NC}"
    fi
else
    echo -e "${YELLOW}⚠ 无法连接 (可能需要密码)${NC}"
fi

# 7. 编译测试
echo ""
echo "==== 7. 编译测试 ===="
echo -n "Go代码编译: "
if go build -o /tmp/openwan_test_build cmd/api/main.go 2>/dev/null; then
    echo -e "${GREEN}✓ 成功${NC}"
    rm -f /tmp/openwan_test_build
else
    echo -e "${RED}✗ 失败${NC}"
    echo "  请运行: go build cmd/api/main.go 查看详细错误"
fi

# 8. 总结
echo ""
echo "======================================"
echo "总结"
echo "======================================"

if [ -d "cmd" ] && [ -d "frontend" ] && [ -f "configs/config.yaml" ]; then
    echo -e "${GREEN}✓ 项目结构完整${NC}"
else
    echo -e "${RED}✗ 项目结构不完整${NC}"
fi

echo ""
echo "建议的下一步操作:"
echo "  1. 确保MySQL正在运行"
echo "  2. 创建数据库: mysql -h localhost -e 'CREATE DATABASE openwan_db'"
echo "  3. 运行数据库迁移"
echo "  4. 安装前端依赖: cd frontend && npm install"
echo "  5. 编译并启动后端: go run cmd/api/main.go"
echo "  6. 启动前端开发服务器: cd frontend && npm run dev"

echo ""
echo "======================================"
