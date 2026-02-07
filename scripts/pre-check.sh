#!/bin/bash

# OpenWan 环境预检查脚本
# 在部署前检查所有依赖和配置

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ERRORS=0
WARNINGS=0

echo "======================================"
echo "OpenWan 环境预检查"
echo "======================================"
echo ""

check_command() {
    local cmd=$1
    local required=$2
    local name=$3
    
    echo -n "检查 $name ... "
    
    if command -v $cmd &> /dev/null; then
        if [ "$cmd" == "go" ]; then
            version=$(go version 2>&1)
        else
            version=$($cmd --version 2>&1 | head -n 1 || echo "已安装")
        fi
        echo -e "${GREEN}✓${NC} $version"
        return 0
    else
        if [ "$required" == "required" ]; then
            echo -e "${RED}✗ 未安装 (必需)${NC}"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${YELLOW}⚠ 未安装 (可选)${NC}"
            WARNINGS=$((WARNINGS + 1))
        fi
        return 1
    fi
}

check_port() {
    local port=$1
    local service=$2
    
    echo -n "检查端口 $port ($service) ... "
    
    if command -v netstat &> /dev/null; then
        if netstat -tuln 2>/dev/null | grep -q ":$port "; then
            echo -e "${YELLOW}⚠ 已被占用${NC}"
            WARNINGS=$((WARNINGS + 1))
        else
            echo -e "${GREEN}✓ 可用${NC}"
        fi
    elif command -v ss &> /dev/null; then
        if ss -tuln 2>/dev/null | grep -q ":$port "; then
            echo -e "${YELLOW}⚠ 已被占用${NC}"
            WARNINGS=$((WARNINGS + 1))
        else
            echo -e "${GREEN}✓ 可用${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ 无法检查${NC}"
    fi
}

check_file() {
    local file=$1
    local required=$2
    
    echo -n "检查文件 $file ... "
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓ 存在${NC}"
        return 0
    else
        if [ "$required" == "required" ]; then
            echo -e "${RED}✗ 不存在 (必需)${NC}"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${YELLOW}⚠ 不存在 (可选)${NC}"
            WARNINGS=$((WARNINGS + 1))
        fi
        return 1
    fi
}

check_directory() {
    local dir=$1
    local writable=$2
    
    echo -n "检查目录 $dir ... "
    
    if [ -d "$dir" ]; then
        if [ "$writable" == "writable" ]; then
            if [ -w "$dir" ]; then
                echo -e "${GREEN}✓ 存在且可写${NC}"
                return 0
            else
                echo -e "${YELLOW}⚠ 存在但不可写${NC}"
                WARNINGS=$((WARNINGS + 1))
                return 1
            fi
        else
            echo -e "${GREEN}✓ 存在${NC}"
            return 0
        fi
    else
        echo -e "${YELLOW}⚠ 不存在 (将自动创建)${NC}"
        WARNINGS=$((WARNINGS + 1))
        return 1
    fi
}

# 1. 检查必需的命令
echo "==== 1. 检查必需的软件 ===="
check_command go required "Go"
check_command node required "Node.js"
check_command npm required "npm"

echo ""
echo "==== 2. 检查可选的软件 ===="
check_command mysql optional "MySQL"
check_command redis-cli optional "Redis"
check_command ffmpeg optional "FFmpeg"
check_command docker optional "Docker"

# 3. 检查端口
echo ""
echo "==== 3. 检查端口可用性 ===="
check_port 8080 "API服务"
check_port 3000 "前端（备用）"
check_port 5173 "前端（Vite默认）"
check_port 3306 "MySQL"
check_port 6379 "Redis"
check_port 5672 "RabbitMQ"

# 4. 检查配置文件
echo ""
echo "==== 4. 检查配置文件 ===="
check_file "configs/config.yaml" required
check_file "go.mod" required
check_file "go.sum" required
check_file "frontend/package.json" required
check_file ".gitignore" optional
check_file "docker-compose.yaml" optional

# 5. 检查目录
echo ""
echo "==== 5. 检查目录结构 ===="
check_directory "cmd" ""
check_directory "internal" ""
check_directory "pkg" ""
check_directory "frontend" ""
check_directory "configs" ""
check_directory "storage" writable
check_directory "logs" writable
check_directory "tmp" writable

# 6. 检查Go模块
echo ""
echo "==== 6. 检查Go依赖 ===="
echo -n "检查go.mod ... "
if grep -q "module" go.mod 2>/dev/null; then
    module_name=$(grep "^module" go.mod | awk '{print $2}')
    echo -e "${GREEN}✓${NC} $module_name"
else
    echo -e "${RED}✗ go.mod格式错误${NC}"
    ERRORS=$((ERRORS + 1))
fi

# 7. 检查前端依赖
echo ""
echo "==== 7. 检查前端依赖 ===="
if [ -d "frontend/node_modules" ]; then
    echo -e "前端依赖: ${GREEN}✓ 已安装${NC}"
else
    echo -e "前端依赖: ${YELLOW}⚠ 未安装 (需要运行 npm install)${NC}"
    WARNINGS=$((WARNINGS + 1))
fi

# 8. 检查数据库配置
echo ""
echo "==== 8. 检查数据库配置 ===="
if [ -f "configs/config.yaml" ]; then
    echo -n "数据库配置 ... "
    if grep -q "database:" configs/config.yaml; then
        db_host=$(grep "host:" configs/config.yaml | head -1 | awk '{print $2}')
        db_name=$(grep "database:" configs/config.yaml | grep -v "^database:" | head -1 | awk '{print $2}')
        echo -e "${GREEN}✓${NC} (Host: $db_host, DB: $db_name)"
    else
        echo -e "${RED}✗ 配置不完整${NC}"
        ERRORS=$((ERRORS + 1))
    fi
fi

# 9. 检查存储配置
echo ""
echo "==== 9. 检查存储配置 ===="
if [ -f "configs/config.yaml" ]; then
    echo -n "存储配置 ... "
    if grep -q "storage:" configs/config.yaml; then
        storage_type=$(grep "type:" configs/config.yaml | head -1 | awk '{print $2}')
        echo -e "${GREEN}✓${NC} (Type: $storage_type)"
        
        if [ "$storage_type" == "local" ]; then
            storage_path=$(grep "local_path:" configs/config.yaml | awk '{print $2}')
            if [ -d "$storage_path" ] || mkdir -p "$storage_path" 2>/dev/null; then
                echo -e "  存储路径: ${GREEN}✓${NC} $storage_path"
            else
                echo -e "  存储路径: ${RED}✗${NC} 无法创建 $storage_path"
                ERRORS=$((ERRORS + 1))
            fi
        fi
    fi
fi

# 10. 系统资源检查
echo ""
echo "==== 10. 检查系统资源 ===="

# 检查磁盘空间
echo -n "磁盘空间 ... "
available_space=$(df -h . | awk 'NR==2 {print $4}')
echo -e "${GREEN}✓${NC} 可用: $available_space"

# 检查内存
if command -v free &> /dev/null; then
    echo -n "系统内存 ... "
    total_mem=$(free -h | awk 'NR==2 {print $2}')
    available_mem=$(free -h | awk 'NR==2 {print $7}')
    echo -e "${GREEN}✓${NC} 总计: $total_mem, 可用: $available_mem"
fi

# 总结
echo ""
echo "======================================"
echo "预检查总结"
echo "======================================"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有检查通过！环境已就绪。${NC}"
    echo ""
    echo "下一步:"
    echo "  1. 确保MySQL和Redis正在运行"
    echo "  2. 运行数据库迁移: cd /home/ec2-user/openwan && migrate -path ./migrations -database 'mysql://user:pass@tcp(localhost:3306)/openwan_db' up"
    echo "  3. 启动服务: ./scripts/quick-start.sh"
    echo "  4. 运行测试: ./scripts/test-deployment.sh"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ 有 $WARNINGS 个警告${NC}"
    echo "环境基本就绪，但建议解决警告项。"
    exit 0
else
    echo -e "${RED}✗ 发现 $ERRORS 个错误, $WARNINGS 个警告${NC}"
    echo "请解决上述错误后再继续。"
    exit 1
fi
