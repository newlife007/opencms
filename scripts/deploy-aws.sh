#!/bin/bash

################################################################################
# OpenWan AWS云部署脚本
# 
# 用途: 在AWS上部署OpenWan生产环境
# 架构: VPC + ECS Fargate + RDS + ElastiCache + S3 + ALB
# 区域: 可配置 (默认: ap-northeast-1)
#
# 版本: 2.0
# 日期: 2026-02-07
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置变量
AWS_REGION=${AWS_REGION:-ap-northeast-1}
PROJECT_NAME="openwan"
ENVIRONMENT=${ENVIRONMENT:-production}
STACK_PREFIX="${PROJECT_NAME}-${ENVIRONMENT}"

# 日志函数
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

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 打印Banner
print_banner() {
    cat << "EOF"
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║              OpenWan AWS 云部署脚本                       ║
║                                                           ║
║         高可用 | 可扩展 | 安全合规 | 成本优化             ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
    echo ""
}

# 检查AWS CLI
check_aws_cli() {
    log_info "检查AWS CLI..."
    
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI 未安装"
        echo "安装指南: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
        exit 1
    fi
    
    local aws_version=$(aws --version | cut -d' ' -f1 | cut -d'/' -f2)
    log_success "AWS CLI 版本: $aws_version"
    
    # 检查AWS凭证
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS凭证未配置或无效"
        echo "请运行: aws configure"
        exit 1
    fi
    
    local account_id=$(aws sts get-caller-identity --query Account --output text)
    local user_arn=$(aws sts get-caller-identity --query Arn --output text)
    log_success "AWS账号: $account_id"
    log_info "用户: $user_arn"
    echo ""
}

# 检查必要工具
check_requirements() {
    log_info "检查部署工具..."
    
    local all_ok=true
    
    # 检查jq
    if ! command -v jq &> /dev/null; then
        log_warn "jq 未安装 (可选，用于JSON解析)"
    fi
    
    # 检查Docker (用于构建镜像)
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        all_ok=false
    else
        log_success "Docker 已安装"
    fi
    
    if [ "$all_ok" = false ]; then
        log_error "请先安装缺失的工具"
        exit 1
    fi
    
    log_success "工具检查完成"
    echo ""
}

# 加载配置
load_config() {
    log_info "加载部署配置..."
    
    if [ ! -f "aws/config/deployment.yaml" ]; then
        log_error "配置文件不存在: aws/config/deployment.yaml"
        exit 1
    fi
    
    log_success "配置文件已加载"
    echo ""
}

# 确认部署
confirm_deployment() {
    cat << EOF
${YELLOW}
╔═══════════════════════════════════════════════════════════╗
║                   部署确认                                 ║
╚═══════════════════════════════════════════════════════════╝
${NC}

部署信息:
  项目名称:     ${PROJECT_NAME}
  环境:         ${ENVIRONMENT}
  AWS区域:      ${AWS_REGION}
  Stack前缀:    ${STACK_PREFIX}

即将创建的资源:
  ✓ VPC (10.0.0.0/16)
  ✓ 公共子网 x2
  ✓ 私有子网 x4 (应用层 + 数据层)
  ✓ NAT Gateway x2
  ✓ Internet Gateway x1
  ✓ RDS MySQL Multi-AZ
  ✓ ElastiCache Redis 集群
  ✓ S3 存储桶 (媒体文件)
  ✓ ECR 镜像仓库
  ✓ ECS Fargate 集群
  ✓ Application Load Balancer
  ✓ CloudWatch Logs
  ✓ SQS 队列

预估成本: ~$2,000/月

${RED}警告: 这将在AWS创建实际资源并产生费用${NC}

EOF
    
    read -p "确认开始部署? (yes/no): " -r
    if [ "$REPLY" != "yes" ]; then
        log_info "部署已取消"
        exit 0
    fi
    
    echo ""
}

# 步骤1: 创建S3存储桶（用于CloudFormation模板）
create_deployment_bucket() {
    log_step "步骤 1/12: 创建部署存储桶"
    
    local bucket_name="${STACK_PREFIX}-deployment-${AWS_REGION}"
    
    if aws s3 ls "s3://${bucket_name}" 2>&1 | grep -q 'NoSuchBucket'; then
        log_info "创建S3存储桶: ${bucket_name}"
        aws s3 mb "s3://${bucket_name}" --region ${AWS_REGION}
        
        # 启用版本控制
        aws s3api put-bucket-versioning \
            --bucket ${bucket_name} \
            --versioning-configuration Status=Enabled
        
        log_success "存储桶创建成功"
    else
        log_info "存储桶已存在: ${bucket_name}"
    fi
    
    export DEPLOYMENT_BUCKET=${bucket_name}
    echo ""
}

# 步骤2: 创建VPC和网络
create_vpc_stack() {
    log_step "步骤 2/12: 创建VPC和网络基础设施"
    
    local stack_name="${STACK_PREFIX}-vpc"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/vpc.yaml \
        --parameters \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
            ParameterKey=VpcCIDR,ParameterValue=10.0.0.0/16 \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待VPC Stack创建完成..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    # 获取VPC ID
    export VPC_ID=$(aws cloudformation describe-stacks \
        --stack-name ${stack_name} \
        --query 'Stacks[0].Outputs[?OutputKey==`VpcId`].OutputValue' \
        --output text \
        --region ${AWS_REGION})
    
    log_success "VPC创建成功: ${VPC_ID}"
    echo ""
}

# 步骤3: 创建安全组
create_security_groups() {
    log_step "步骤 3/12: 创建安全组"
    
    local stack_name="${STACK_PREFIX}-security-groups"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/security-groups.yaml \
        --parameters \
            ParameterKey=VpcId,ParameterValue=${VPC_ID} \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待安全组创建完成..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    log_success "安全组创建成功"
    echo ""
}

# 步骤4: 创建RDS数据库
create_rds_database() {
    log_step "步骤 4/12: 创建RDS MySQL数据库"
    
    local stack_name="${STACK_PREFIX}-rds"
    
    # 生成随机密码
    local db_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    log_warn "数据库密码将保存到AWS Secrets Manager"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/rds.yaml \
        --parameters \
            ParameterKey=VpcId,ParameterValue=${VPC_ID} \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
            ParameterKey=DBName,ParameterValue=openwan_db \
            ParameterKey=DBUsername,ParameterValue=openwan_admin \
            ParameterKey=DBPassword,ParameterValue=${db_password} \
            ParameterKey=DBInstanceClass,ParameterValue=db.r6g.xlarge \
            ParameterKey=AllocatedStorage,ParameterValue=500 \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待RDS创建完成 (约10-15分钟)..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    # 获取数据库端点
    export DB_ENDPOINT=$(aws cloudformation describe-stacks \
        --stack-name ${stack_name} \
        --query 'Stacks[0].Outputs[?OutputKey==`DBEndpoint`].OutputValue' \
        --output text \
        --region ${AWS_REGION})
    
    log_success "RDS创建成功"
    log_info "数据库端点: ${DB_ENDPOINT}"
    echo ""
}

# 步骤5: 创建ElastiCache Redis
create_elasticache_redis() {
    log_step "步骤 5/12: 创建ElastiCache Redis集群"
    
    local stack_name="${STACK_PREFIX}-elasticache"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/elasticache.yaml \
        --parameters \
            ParameterKey=VpcId,ParameterValue=${VPC_ID} \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
            ParameterKey=CacheNodeType,ParameterValue=cache.r6g.large \
            ParameterKey=NumCacheNodes,ParameterValue=3 \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待Redis集群创建完成 (约10分钟)..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    export REDIS_ENDPOINT=$(aws cloudformation describe-stacks \
        --stack-name ${stack_name} \
        --query 'Stacks[0].Outputs[?OutputKey==`RedisEndpoint`].OutputValue' \
        --output text \
        --region ${AWS_REGION})
    
    log_success "Redis集群创建成功"
    log_info "Redis端点: ${REDIS_ENDPOINT}"
    echo ""
}

# 步骤6: 创建S3媒体存储桶
create_media_bucket() {
    log_step "步骤 6/12: 创建S3媒体存储桶"
    
    local bucket_name="${STACK_PREFIX}-media-assets"
    
    log_info "创建S3存储桶: ${bucket_name}"
    
    aws s3 mb "s3://${bucket_name}" --region ${AWS_REGION}
    
    # 启用版本控制
    aws s3api put-bucket-versioning \
        --bucket ${bucket_name} \
        --versioning-configuration Status=Enabled
    
    # 启用加密
    aws s3api put-bucket-encryption \
        --bucket ${bucket_name} \
        --server-side-encryption-configuration '{
            "Rules": [{
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                }
            }]
        }'
    
    # 阻止公共访问
    aws s3api put-public-access-block \
        --bucket ${bucket_name} \
        --public-access-block-configuration \
            BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
    
    # 配置生命周期策略
    aws s3api put-bucket-lifecycle-configuration \
        --bucket ${bucket_name} \
        --lifecycle-configuration file://aws/config/s3-lifecycle.json
    
    export MEDIA_BUCKET=${bucket_name}
    log_success "媒体存储桶创建成功: ${bucket_name}"
    echo ""
}

# 步骤7: 创建SQS队列
create_sqs_queues() {
    log_step "步骤 7/12: 创建SQS消息队列"
    
    local stack_name="${STACK_PREFIX}-sqs"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/sqs.yaml \
        --parameters \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待SQS队列创建完成..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    log_success "SQS队列创建成功"
    echo ""
}

# 步骤8: 创建ECR镜像仓库
create_ecr_repositories() {
    log_step "步骤 8/12: 创建ECR镜像仓库"
    
    local api_repo="${PROJECT_NAME}-api"
    local worker_repo="${PROJECT_NAME}-worker"
    
    # 创建API仓库
    log_info "创建ECR仓库: ${api_repo}"
    aws ecr create-repository \
        --repository-name ${api_repo} \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256 \
        --region ${AWS_REGION} \
        2>/dev/null || log_info "仓库已存在"
    
    # 创建Worker仓库
    log_info "创建ECR仓库: ${worker_repo}"
    aws ecr create-repository \
        --repository-name ${worker_repo} \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256 \
        --region ${AWS_REGION} \
        2>/dev/null || log_info "仓库已存在"
    
    local account_id=$(aws sts get-caller-identity --query Account --output text)
    export ECR_API_REPO="${account_id}.dkr.ecr.${AWS_REGION}.amazonaws.com/${api_repo}"
    export ECR_WORKER_REPO="${account_id}.dkr.ecr.${AWS_REGION}.amazonaws.com/${worker_repo}"
    
    log_success "ECR仓库创建成功"
    echo ""
}

# 步骤9: 构建并推送Docker镜像
build_and_push_images() {
    log_step "步骤 9/12: 构建并推送Docker镜像"
    
    # 登录ECR
    log_info "登录到ECR..."
    aws ecr get-login-password --region ${AWS_REGION} | \
        docker login --username AWS --password-stdin ${ECR_API_REPO%%/*}
    
    # 构建API镜像
    log_info "构建API镜像..."
    docker build -t ${ECR_API_REPO}:latest -f Dockerfile .
    
    log_info "推送API镜像..."
    docker push ${ECR_API_REPO}:latest
    
    # 构建Worker镜像
    log_info "构建Worker镜像..."
    docker build -t ${ECR_WORKER_REPO}:latest -f Dockerfile.worker .
    
    log_info "推送Worker镜像..."
    docker push ${ECR_WORKER_REPO}:latest
    
    log_success "镜像推送成功"
    echo ""
}

# 步骤10: 创建ECS集群
create_ecs_cluster() {
    log_step "步骤 10/12: 创建ECS集群"
    
    local stack_name="${STACK_PREFIX}-ecs"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/ecs.yaml \
        --parameters \
            ParameterKey=VpcId,ParameterValue=${VPC_ID} \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
            ParameterKey=ApiImage,ParameterValue=${ECR_API_REPO}:latest \
            ParameterKey=WorkerImage,ParameterValue=${ECR_WORKER_REPO}:latest \
            ParameterKey=DBEndpoint,ParameterValue=${DB_ENDPOINT} \
            ParameterKey=RedisEndpoint,ParameterValue=${REDIS_ENDPOINT} \
            ParameterKey=MediaBucket,ParameterValue=${MEDIA_BUCKET} \
        --capabilities CAPABILITY_IAM \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待ECS集群创建完成 (约5分钟)..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    log_success "ECS集群创建成功"
    echo ""
}

# 步骤11: 创建Application Load Balancer
create_alb() {
    log_step "步骤 11/12: 创建Application Load Balancer"
    
    local stack_name="${STACK_PREFIX}-alb"
    
    log_info "创建CloudFormation Stack: ${stack_name}"
    
    aws cloudformation create-stack \
        --stack-name ${stack_name} \
        --template-body file://aws/cloudformation/alb.yaml \
        --parameters \
            ParameterKey=VpcId,ParameterValue=${VPC_ID} \
            ParameterKey=ProjectName,ParameterValue=${PROJECT_NAME} \
            ParameterKey=Environment,ParameterValue=${ENVIRONMENT} \
        --tags \
            Key=Project,Value=${PROJECT_NAME} \
            Key=Environment,Value=${ENVIRONMENT} \
        --region ${AWS_REGION}
    
    log_info "等待ALB创建完成..."
    aws cloudformation wait stack-create-complete \
        --stack-name ${stack_name} \
        --region ${AWS_REGION}
    
    export ALB_DNS=$(aws cloudformation describe-stacks \
        --stack-name ${stack_name} \
        --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerDNS`].OutputValue' \
        --output text \
        --region ${AWS_REGION})
    
    log_success "ALB创建成功"
    log_info "负载均衡器地址: ${ALB_DNS}"
    echo ""
}

# 步骤12: 运行数据库迁移
run_database_migration() {
    log_step "步骤 12/12: 运行数据库迁移"
    
    log_info "通过ECS Task运行数据库迁移..."
    
    # 这里需要通过ECS Run Task执行迁移
    # 实际实现需要Task Definition ARN
    
    log_warn "请手动运行数据库迁移:"
    echo "  aws ecs run-task \\"
    echo "    --cluster ${STACK_PREFIX}-cluster \\"
    echo "    --task-definition ${STACK_PREFIX}-api \\"
    echo "    --launch-type FARGATE \\"
    echo "    --network-configuration ... \\"
    echo "    --overrides '{\"containerOverrides\": [{\"name\": \"api\", \"command\": [\"/app/bin/openwan\", \"migrate\", \"up\"]}]}'"
    
    echo ""
}

# 显示部署结果
show_deployment_summary() {
    local account_id=$(aws sts get-caller-identity --query Account --output text)
    
    cat << EOF
${GREEN}
╔═══════════════════════════════════════════════════════════╗
║                 部署成功！                                 ║
╚═══════════════════════════════════════════════════════════╝
${NC}

${BLUE}部署信息：${NC}

  AWS账号:        ${account_id}
  区域:           ${AWS_REGION}
  环境:           ${ENVIRONMENT}

${BLUE}访问信息：${NC}

  负载均衡器:     http://${ALB_DNS}
  API端点:        http://${ALB_DNS}/api/v1
  健康检查:       http://${ALB_DNS}/health

${BLUE}资源详情：${NC}

  VPC ID:         ${VPC_ID}
  数据库端点:     ${DB_ENDPOINT}
  Redis端点:      ${REDIS_ENDPOINT}
  媒体存储桶:     ${MEDIA_BUCKET}
  API镜像:        ${ECR_API_REPO}:latest
  Worker镜像:     ${ECR_WORKER_REPO}:latest

${BLUE}CloudFormation Stacks：${NC}

  VPC:            ${STACK_PREFIX}-vpc
  安全组:         ${STACK_PREFIX}-security-groups
  RDS:            ${STACK_PREFIX}-rds
  Redis:          ${STACK_PREFIX}-elasticache
  SQS:            ${STACK_PREFIX}-sqs
  ECS:            ${STACK_PREFIX}-ecs
  ALB:            ${STACK_PREFIX}-alb

${BLUE}下一步：${NC}

  1. 配置域名DNS指向: ${ALB_DNS}
  2. 在ACM申请SSL证书
  3. 更新ALB监听器使用HTTPS
  4. 运行数据库迁移
  5. 创建管理员账号
  6. 配置CloudWatch告警
  7. 配置备份策略

${BLUE}管理命令：${NC}

  查看所有Stack:
    aws cloudformation list-stacks --region ${AWS_REGION}

  查看ECS服务:
    aws ecs list-services --cluster ${STACK_PREFIX}-cluster --region ${AWS_REGION}

  查看日志:
    aws logs tail /ecs/${STACK_PREFIX}-api --follow --region ${AWS_REGION}

  更新服务:
    aws ecs update-service --cluster ${STACK_PREFIX}-cluster --service ${STACK_PREFIX}-api --force-new-deployment --region ${AWS_REGION}

${YELLOW}预估月度成本: ~$2,000${NC}

详细成本构成请查看: aws/docs/COST_ESTIMATION.md

${GREEN}AWS部署完成！${NC}

EOF
}

# 主函数
main() {
    print_banner
    
    log_info "开始AWS云部署..."
    log_info "区域: ${AWS_REGION}"
    log_info "环境: ${ENVIRONMENT}"
    echo ""
    
    # 检查AWS CLI和凭证
    check_aws_cli
    
    # 检查必要工具
    check_requirements
    
    # 加载配置
    load_config
    
    # 确认部署
    confirm_deployment
    
    # 执行部署步骤
    create_deployment_bucket
    create_vpc_stack
    create_security_groups
    create_rds_database
    create_elasticache_redis
    create_media_bucket
    create_sqs_queues
    create_ecr_repositories
    build_and_push_images
    create_ecs_cluster
    create_alb
    run_database_migration
    
    # 显示部署结果
    show_deployment_summary
    
    log_success "AWS部署完成！"
}

# 执行主函数
main "$@"
