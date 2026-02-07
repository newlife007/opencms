#!/bin/bash

################################################################################
# OpenWan AWS资源销毁脚本
# 
# 用途: 删除所有AWS上的OpenWan资源
# 警告: 这将删除所有数据，请谨慎操作
#
# 版本: 2.0
# 日期: 2026-02-07
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

AWS_REGION=${AWS_REGION:-ap-northeast-1}
PROJECT_NAME="openwan"
ENVIRONMENT=${ENVIRONMENT:-production}
STACK_PREFIX="${PROJECT_NAME}-${ENVIRONMENT}"

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印警告Banner
print_warning_banner() {
    cat << EOF
${RED}
╔═══════════════════════════════════════════════════════════╗
║                     ⚠️  警告  ⚠️                          ║
║                                                           ║
║              AWS资源销毁脚本                              ║
║                                                           ║
║   这将删除所有OpenWan相关的AWS资源和数据                  ║
║              操作不可逆！                                  ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
${NC}
EOF
}

# 列出要删除的资源
list_resources() {
    cat << EOF
${YELLOW}即将删除的资源:${NC}

CloudFormation Stacks:
  ✗ ${STACK_PREFIX}-alb
  ✗ ${STACK_PREFIX}-ecs
  ✗ ${STACK_PREFIX}-sqs
  ✗ ${STACK_PREFIX}-elasticache
  ✗ ${STACK_PREFIX}-rds
  ✗ ${STACK_PREFIX}-security-groups
  ✗ ${STACK_PREFIX}-vpc

S3存储桶:
  ✗ ${STACK_PREFIX}-media-assets
  ✗ ${STACK_PREFIX}-deployment-${AWS_REGION}

ECR仓库:
  ✗ ${PROJECT_NAME}-api
  ✗ ${PROJECT_NAME}-worker

CloudWatch日志组:
  ✗ /ecs/${STACK_PREFIX}-api
  ✗ /ecs/${STACK_PREFIX}-worker

EOF
}

# 三次确认
confirm_deletion() {
    print_warning_banner
    list_resources
    
    echo -e "${RED}═══════════════════════════════════════════════${NC}"
    echo -e "${RED}请输入环境名称以确认删除: ${ENVIRONMENT}${NC}"
    read -p "输入环境名称: " -r
    if [ "$REPLY" != "$ENVIRONMENT" ]; then
        log_info "取消删除"
        exit 0
    fi
    
    echo ""
    echo -e "${RED}第二次确认${NC}"
    read -p "确认删除所有资源？这将无法恢复 (yes/no): " -r
    if [ "$REPLY" != "yes" ]; then
        log_info "取消删除"
        exit 0
    fi
    
    echo ""
    echo -e "${RED}最后确认${NC}"
    read -p "你真的确定要删除吗？输入DELETE继续: " -r
    if [ "$REPLY" != "DELETE" ]; then
        log_info "取消删除"
        exit 0
    fi
    
    echo ""
    log_warn "开始删除资源..."
    sleep 3
}

# 删除CloudFormation Stack
delete_stack() {
    local stack_name=$1
    
    log_info "删除Stack: ${stack_name}"
    
    if aws cloudformation describe-stacks --stack-name ${stack_name} --region ${AWS_REGION} &> /dev/null; then
        aws cloudformation delete-stack \
            --stack-name ${stack_name} \
            --region ${AWS_REGION}
        
        log_info "等待Stack删除完成..."
        aws cloudformation wait stack-delete-complete \
            --stack-name ${stack_name} \
            --region ${AWS_REGION} || true
        
        log_success "Stack删除成功: ${stack_name}"
    else
        log_warn "Stack不存在: ${stack_name}"
    fi
    
    echo ""
}

# 清空并删除S3存储桶
delete_bucket() {
    local bucket_name=$1
    
    log_info "删除S3存储桶: ${bucket_name}"
    
    if aws s3 ls "s3://${bucket_name}" &> /dev/null; then
        # 清空存储桶
        log_info "清空存储桶..."
        aws s3 rm "s3://${bucket_name}" --recursive --region ${AWS_REGION}
        
        # 删除所有版本（如果启用了版本控制）
        aws s3api delete-objects \
            --bucket ${bucket_name} \
            --delete "$(aws s3api list-object-versions \
                --bucket ${bucket_name} \
                --output json \
                --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}' \
                --region ${AWS_REGION})" \
            --region ${AWS_REGION} 2>/dev/null || true
        
        # 删除存储桶
        aws s3 rb "s3://${bucket_name}" --region ${AWS_REGION}
        log_success "存储桶删除成功: ${bucket_name}"
    else
        log_warn "存储桶不存在: ${bucket_name}"
    fi
    
    echo ""
}

# 删除ECR仓库
delete_ecr_repo() {
    local repo_name=$1
    
    log_info "删除ECR仓库: ${repo_name}"
    
    aws ecr delete-repository \
        --repository-name ${repo_name} \
        --force \
        --region ${AWS_REGION} 2>/dev/null || log_warn "仓库不存在: ${repo_name}"
    
    log_success "ECR仓库删除成功: ${repo_name}"
    echo ""
}

# 删除CloudWatch日志组
delete_log_group() {
    local log_group=$1
    
    log_info "删除CloudWatch日志组: ${log_group}"
    
    aws logs delete-log-group \
        --log-group-name ${log_group} \
        --region ${AWS_REGION} 2>/dev/null || log_warn "日志组不存在: ${log_group}"
    
    log_success "日志组删除成功: ${log_group}"
    echo ""
}

# 主删除流程
main() {
    log_info "开始AWS资源销毁流程"
    log_info "区域: ${AWS_REGION}"
    log_info "环境: ${ENVIRONMENT}"
    echo ""
    
    # 三次确认
    confirm_deletion
    
    # 按依赖顺序删除Stack
    # 注意: 必须先删除依赖资源，后删除被依赖资源
    
    log_info "步骤1: 删除ALB"
    delete_stack "${STACK_PREFIX}-alb"
    
    log_info "步骤2: 删除ECS集群"
    delete_stack "${STACK_PREFIX}-ecs"
    
    log_info "步骤3: 删除SQS队列"
    delete_stack "${STACK_PREFIX}-sqs"
    
    log_info "步骤4: 删除ElastiCache"
    delete_stack "${STACK_PREFIX}-elasticache"
    
    log_info "步骤5: 删除RDS数据库"
    delete_stack "${STACK_PREFIX}-rds"
    
    log_info "步骤6: 删除安全组"
    delete_stack "${STACK_PREFIX}-security-groups"
    
    log_info "步骤7: 删除VPC"
    delete_stack "${STACK_PREFIX}-vpc"
    
    log_info "步骤8: 删除S3存储桶"
    delete_bucket "${STACK_PREFIX}-media-assets"
    delete_bucket "${STACK_PREFIX}-deployment-${AWS_REGION}"
    
    log_info "步骤9: 删除ECR仓库"
    delete_ecr_repo "${PROJECT_NAME}-api"
    delete_ecr_repo "${PROJECT_NAME}-worker"
    
    log_info "步骤10: 删除CloudWatch日志"
    delete_log_group "/ecs/${STACK_PREFIX}-api"
    delete_log_group "/ecs/${STACK_PREFIX}-worker"
    
    cat << EOF
${GREEN}
╔═══════════════════════════════════════════════════════════╗
║                 资源删除完成                               ║
╚═══════════════════════════════════════════════════════════╝
${NC}

所有OpenWan相关的AWS资源已被删除。

${YELLOW}注意事项:${NC}
  - 某些资源可能有延迟删除
  - 请检查AWS控制台确认所有资源已删除
  - 检查是否有孤立的资源产生费用

${BLUE}验证删除:${NC}
  aws cloudformation list-stacks --region ${AWS_REGION}
  aws s3 ls | grep ${PROJECT_NAME}
  aws ecr describe-repositories --region ${AWS_REGION}

${GREEN}清理完成！${NC}

EOF
}

# 执行主函数
main "$@"
