#!/bin/bash
#
# OpenWan 服务启动脚本
# 启动完整的后端服务（API + Worker）与 S3 存储集成
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  OpenWan 服务启动脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 检查依赖服务
echo -e "${YELLOW}[1/6] 检查依赖服务...${NC}"

check_service() {
    local service=$1
    local port=$2
    local name=$3
    
    if pgrep -f "$service" > /dev/null; then
        echo -e "  ✓ $name 运行中"
        return 0
    else
        echo -e "  ${RED}✗ $name 未运行${NC}"
        return 1
    fi
}

check_service "mysqld" "3306" "MySQL"
check_service "redis-server" "6379" "Redis"
check_service "rabbitmq" "5672" "RabbitMQ"

# 2. 检查数据库连接
echo -e "\n${YELLOW}[2/6] 检查数据库连接...${NC}"
if mysql -h 127.0.0.1 -u openwan -popenwan123 -e "USE openwan_db;" 2>/dev/null; then
    TABLE_COUNT=$(mysql -h 127.0.0.1 -u openwan -popenwan123 -D openwan_db -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'openwan_db';" -sN)
    echo -e "  ✓ 数据库 openwan_db 连接成功 ($TABLE_COUNT 张表)"
else
    echo -e "  ${RED}✗ 数据库连接失败${NC}"
    exit 1
fi

# 3. 检查 AWS S3 配置
echo -e "\n${YELLOW}[3/6] 检查 AWS S3 配置...${NC}"
if [ -f ~/.aws/credentials ]; then
    echo -e "  ✓ AWS 凭证文件存在"
    
    # 从配置读取S3桶名
    S3_BUCKET=$(grep s3_bucket configs/config.yaml | awk '{print $2}' | tr -d '"')
    S3_REGION=$(grep s3_region configs/config.yaml | awk '{print $2}')
    
    echo -e "  ✓ S3 存储桶: $S3_BUCKET"
    echo -e "  ✓ AWS 区域: $S3_REGION"
    
    # 测试S3访问
    if aws s3 ls "s3://$S3_BUCKET/" --region "$S3_REGION" > /dev/null 2>&1; then
        echo -e "  ✓ S3 存储桶访问正常"
    else
        echo -e "  ${YELLOW}⚠ S3 存储桶访问测试失败，但将继续启动${NC}"
    fi
else
    echo -e "  ${RED}✗ AWS 凭证文件不存在${NC}"
    exit 1
fi

# 4. 检查二进制文件
echo -e "\n${YELLOW}[4/6] 检查编译文件...${NC}"
if [ ! -f bin/openwan ]; then
    echo -e "  ${RED}✗ 后端服务文件不存在，正在编译...${NC}"
    go build -o bin/openwan ./cmd/api
fi
echo -e "  ✓ 后端服务: bin/openwan ($(ls -lh bin/openwan | awk '{print $5}'))"

if [ ! -f bin/openwan-worker ]; then
    echo -e "  ${RED}✗ Worker服务文件不存在，正在编译...${NC}"
    go build -o bin/openwan-worker ./cmd/worker
fi
echo -e "  ✓ Worker服务: bin/openwan-worker ($(ls -lh bin/openwan-worker | awk '{print $5}'))"

# 5. 停止已有进程
echo -e "\n${YELLOW}[5/6] 停止已有服务进程...${NC}"
if pgrep -f "bin/openwan" > /dev/null; then
    echo -e "  停止旧的后端进程..."
    pkill -f "bin/openwan" || true
    sleep 1
fi

if pgrep -f "bin/openwan-worker" > /dev/null; then
    echo -e "  停止旧的Worker进程..."
    pkill -f "bin/openwan-worker" || true
    sleep 1
fi

# 6. 启动服务
echo -e "\n${YELLOW}[6/6] 启动服务...${NC}"

# 创建日志目录
mkdir -p logs

# 设置AWS环境变量（从AWS凭证文件读取）
echo -e "  设置AWS环境变量..."
if [ -f ~/.aws/credentials ]; then
    export AWS_ACCESS_KEY_ID=$(grep aws_access_key_id ~/.aws/credentials | head -1 | awk '{print $3}')
    export AWS_SECRET_ACCESS_KEY=$(grep aws_secret_access_key ~/.aws/credentials | head -1 | awk '{print $3}')
    export AWS_DEFAULT_REGION=$(grep region ~/.aws/config | head -1 | awk '{print $3}')
    echo -e "  ✓ AWS凭证已导出到环境变量"
fi

# 启动后端 API 服务
echo -e "  启动后端 API 服务..."
nohup ./bin/openwan > logs/api.log 2>&1 &
API_PID=$!
echo $API_PID > logs/api.pid
echo -e "  ✓ 后端 API 服务已启动 (PID: $API_PID)"

# 等待API启动
sleep 2

# 检查API是否启动成功
if ps -p $API_PID > /dev/null; then
    echo -e "  ✓ 后端 API 服务运行中"
    
    # 等待健康检查
    echo -e "  等待健康检查..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓ 健康检查通过${NC}"
            break
        fi
        if [ $i -eq 10 ]; then
            echo -e "  ${YELLOW}⚠ 健康检查超时，但服务已启动${NC}"
        fi
        sleep 1
    done
else
    echo -e "  ${RED}✗ 后端 API 服务启动失败${NC}"
    echo -e "  查看日志: tail -f logs/api.log"
    exit 1
fi

# 启动 Worker 服务（2个实例）
echo -e "\n  启动 Worker 服务 (2个实例)..."
nohup ./bin/openwan-worker > logs/worker-1.log 2>&1 &
WORKER1_PID=$!
echo $WORKER1_PID > logs/worker-1.pid
echo -e "  ✓ Worker #1 已启动 (PID: $WORKER1_PID)"

sleep 1

nohup ./bin/openwan-worker > logs/worker-2.log 2>&1 &
WORKER2_PID=$!
echo $WORKER2_PID > logs/worker-2.pid
echo -e "  ✓ Worker #2 已启动 (PID: $WORKER2_PID)"

# 总结
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "服务状态:"
echo -e "  • 后端 API:    http://localhost:8080 (PID: $API_PID)"
echo -e "  • Worker #1:   PID: $WORKER1_PID"
echo -e "  • Worker #2:   PID: $WORKER2_PID"
echo -e "  • 前端 Web:    http://localhost (Nginx)"
echo -e "  • 存储类型:    AWS S3"
echo -e "  • S3 存储桶:   $S3_BUCKET"
echo ""
echo -e "查看日志:"
echo -e "  • API 日志:     tail -f logs/api.log"
echo -e "  • Worker 日志:  tail -f logs/worker-1.log"
echo -e "  • Worker 日志:  tail -f logs/worker-2.log"
echo ""
echo -e "停止服务:"
echo -e "  • 运行: ./stop-services.sh"
echo -e "  • 或手动: kill $API_PID $WORKER1_PID $WORKER2_PID"
echo ""
echo -e "API端点测试:"
echo -e "  • 健康检查:     curl http://localhost:8080/health"
echo -e "  • 登录测试:     curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo -e "                    -H 'Content-Type: application/json' \\"
echo -e "                    -d '{\"username\":\"admin\",\"password\":\"admin\"}'"
echo ""
echo -e "${GREEN}✓ 启动完成！现在可以通过浏览器访问 http://localhost${NC}"
echo ""
