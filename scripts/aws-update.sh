#!/bin/bash
################################################################################
# OpenWan AWS更新部署脚本  
################################################################################
set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'

AWS_REGION=${AWS_REGION:-ap-northeast-1}
PROJECT_NAME="openwan"
ENVIRONMENT=${ENVIRONMENT:-production}
STACK_PREFIX="${PROJECT_NAME}-${ENVIRONMENT}"

echo -e "${BLUE}更新OpenWan AWS部署${NC}"
echo ""

# 获取ECR信息
account_id=$(aws sts get-caller-identity --query Account --output text)
ECR_API_REPO="${account_id}.dkr.ecr.${AWS_REGION}.amazonaws.com/${PROJECT_NAME}-api"
ECR_WORKER_REPO="${account_id}.dkr.ecr.${AWS_REGION}.amazonaws.com/${PROJECT_NAME}-worker"

# 登录ECR
echo "登录ECR..."
aws ecr get-login-password --region ${AWS_REGION} | \
    docker login --username AWS --password-stdin ${ECR_API_REPO%%/*}

# 构建新镜像
echo "构建API镜像..."
docker build -t ${ECR_API_REPO}:latest -f Dockerfile .

echo "构建Worker镜像..."
docker build -t ${ECR_WORKER_REPO}:latest -f Dockerfile.worker .

# 推送镜像
echo "推送API镜像..."
docker push ${ECR_API_REPO}:latest

echo "推送Worker镜像..."
docker push ${ECR_WORKER_REPO}:latest

# 强制ECS更新服务
echo "更新ECS服务..."
aws ecs update-service \
    --cluster ${STACK_PREFIX}-cluster \
    --service ${STACK_PREFIX}-api \
    --force-new-deployment \
    --region ${AWS_REGION}

aws ecs update-service \
    --cluster ${STACK_PREFIX}-cluster \
    --service ${STACK_PREFIX}-worker \
    --force-new-deployment \
    --region ${AWS_REGION}

echo -e "${GREEN}✓ 更新完成！${NC}"
echo ""
echo "查看部署状态:"
echo "  ./scripts/aws-status.sh"
