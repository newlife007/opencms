#!/bin/bash
################################################################################
# OpenWan AWS状态查看脚本
################################################################################
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

AWS_REGION=${AWS_REGION:-ap-northeast-1}
PROJECT_NAME="openwan"
ENVIRONMENT=${ENVIRONMENT:-production}
STACK_PREFIX="${PROJECT_NAME}-${ENVIRONMENT}"

echo -e "${BLUE}OpenWan AWS部署状态${NC}"
echo "区域: ${AWS_REGION}"
echo "环境: ${ENVIRONMENT}"
echo ""

# 检查CloudFormation Stacks
echo -e "${BLUE}CloudFormation Stacks:${NC}"
for stack in vpc security-groups rds elasticache sqs ecs alb; do
    stack_name="${STACK_PREFIX}-${stack}"
    status=$(aws cloudformation describe-stacks \
        --stack-name ${stack_name} \
        --query 'Stacks[0].StackStatus' \
        --output text \
        --region ${AWS_REGION} 2>/dev/null || echo "NOT_FOUND")
    
    if [ "$status" = "CREATE_COMPLETE" ] || [ "$status" = "UPDATE_COMPLETE" ]; then
        echo -e "  ${stack}: ${GREEN}✓ ${status}${NC}"
    elif [ "$status" = "NOT_FOUND" ]; then
        echo -e "  ${stack}: ${YELLOW}✗ 不存在${NC}"
    else
        echo -e "  ${stack}: ${RED}${status}${NC}"
    fi
done

echo ""
echo -e "${BLUE}查看详细日志:${NC}"
echo "  aws logs tail /ecs/${STACK_PREFIX}-api --follow --region ${AWS_REGION}"
