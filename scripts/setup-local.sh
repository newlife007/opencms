#!/bin/bash

################################################################################
# OpenWan 本地开发环境一键部署脚本
# 
# 用途: 在本地快速搭建OpenWan开发环境
# 要求: Docker, Docker Compose, Git
# 用法: ./scripts/setup-local.sh
#
# 版本: 2.0
# 日期: 2026-02-07
################################################################################

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 打印Banner
print_banner() {
    cat << "EOF"
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ___                __        __             ____  __   ║
║  / _ \ ___  ___ ___/ / /    /\ \ \__ _ _ __  |___ \/ /   ║
║ | | | / _ \/ _ \/ __/ / /___/  \/ / _` | '_ \   __) | |  ║
║ | |_| | (_) |  __/ /  \___/ /\  / (_| | | | | / __/| |  ║
║  \___/ \___/\___/_/      \_\ \/ \__,_|_| |_||_____|_|   ║
║                                                           ║
║     媒体资产管理系统 - 本地开发环境部署脚本               ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
    echo ""
}

# 检查命令是否存在
check_command() {
    local cmd=$1
    local install_hint=$2
    
    if ! command -v $cmd &> /dev/null; then
        log_error "$cmd 未安装"
        if [ -n "$install_hint" ]; then
            echo "  安装提示: $install_hint"
        fi
        return 1
    else
        log_success "$cmd 已安装"
        return 0
    fi
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    local all_ok=true
    
    # 检查Docker
    if ! check_command "docker" "请访问 https://docs.docker.com/get-docker/"; then
        all_ok=false
    else
        # 检查Docker版本
        local docker_version=$(docker --version | grep -oP '\d+\.\d+' | head -1)
        log_info "Docker 版本: $docker_version"
    fi
    
    # 检查Docker Compose
    if ! check_command "docker-compose" "请访问 https://docs.docker.com/compose/install/"; then
        all_ok=false
    else
        local compose_version=$(docker-compose --version | grep -oP '\d+\.\d+' | head -1)
        log_info "Docker Compose 版本: $compose_version"
    fi
    
    # 检查Git
    if ! check_command "git" "请使用包管理器安装: apt-get install git 或 yum install git"; then
        all_ok=false
    fi
    
    # 检查磁盘空间（至少需要20GB）
    local available_space=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
    if [ "$available_space" -lt 20 ]; then
        log_error "磁盘空间不足，至少需要20GB，当前可用: ${available_space}GB"
        all_ok=false
    else
        log_success "磁盘空间充足: ${available_space}GB"
    fi
    
    # 检查内存（至少需要8GB）
    local total_mem=$(free -g | grep Mem | awk '{print $2}')
    if [ "$total_mem" -lt 8 ]; then
        log_warn "内存较少，建议至少8GB，当前: ${total_mem}GB"
    else
        log_success "内存充足: ${total_mem}GB"
    fi
    
    if [ "$all_ok" = false ]; then
        log_error "系统要求检查失败，请先安装缺失的组件"
        exit 1
    fi
    
    log_success "系统要求检查通过"
    echo ""
}

# 停止已运行的容器
stop_existing_containers() {
    log_info "停止已运行的OpenWan容器..."
    
    if docker-compose ps -q | grep -q .; then
        docker-compose down
        log_success "已停止旧容器"
    else
        log_info "没有运行中的容器"
    fi
    echo ""
}

# 清理旧数据（可选）
cleanup_old_data() {
    read -p "是否清理旧数据？这将删除数据库和上传的文件 (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_warn "清理旧数据..."
        
        # 删除Docker volumes
        docker-compose down -v
        
        # 删除本地数据目录
        if [ -d "./storage/uploads" ]; then
            rm -rf ./storage/uploads/*
            log_success "已清理上传文件"
        fi
        
        log_success "数据清理完成"
    else
        log_info "跳过数据清理"
    fi
    echo ""
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    # 数据目录
    mkdir -p storage/uploads/data1
    mkdir -p storage/uploads/temp
    mkdir -p storage/logs
    
    # 数据库数据目录
    mkdir -p data/mysql
    mkdir -p data/redis
    mkdir -p data/rabbitmq
    mkdir -p data/sphinx
    
    # 日志目录
    mkdir -p logs/api
    mkdir -p logs/worker
    mkdir -p logs/nginx
    
    log_success "目录创建完成"
    echo ""
}

# 生成.env文件
generate_env_file() {
    log_info "生成环境配置文件..."
    
    if [ -f ".env" ]; then
        log_warn ".env 文件已存在"
        read -p "是否覆盖现有配置？(y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "跳过.env文件生成"
            echo ""
            return
        fi
    fi
    
    # 生成随机密码
    local db_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    local redis_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    local rabbitmq_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    local jwt_secret=$(openssl rand -base64 64 | tr -d "=+/" | cut -c1-50)
    
    cat > .env << EOF
# OpenWan 本地开发环境配置
# 生成时间: $(date)

# 应用配置
APP_NAME=OpenWan
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_DATABASE=openwan_db
DB_USERNAME=openwan
DB_PASSWORD=${db_password}

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=${redis_password}
REDIS_DB=0

# RabbitMQ配置
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=openwan
RABBITMQ_PASSWORD=${rabbitmq_password}
RABBITMQ_VHOST=/

# Sphinx配置
SPHINX_HOST=sphinx
SPHINX_PORT=9306

# 存储配置（本地开发使用本地存储）
STORAGE_TYPE=local
STORAGE_PATH=/app/storage/uploads

# FFmpeg配置
FFMPEG_PATH=/usr/local/bin/ffmpeg
FFMPEG_PARAMS=-y -c:v libx264 -c:a aac -b:a 56k -ar 22050 -r 15 -b:v 500k -s 320x240 -preset fast -profile:v baseline

# JWT配置
JWT_SECRET=${jwt_secret}
JWT_EXPIRE=24h

# 日志配置
LOG_LEVEL=debug
LOG_FORMAT=json

# 前端配置
FRONTEND_PORT=3000
VITE_API_BASE_URL=http://localhost:8080/api/v1
EOF
    
    log_success ".env 文件生成完成"
    log_info "数据库密码: ${db_password}"
    log_info "Redis密码: ${redis_password}"
    log_warn "请妥善保管这些密码！"
    echo ""
}

# 拉取Docker镜像
pull_docker_images() {
    log_info "拉取Docker镜像..."
    
    docker-compose pull
    
    log_success "镜像拉取完成"
    echo ""
}

# 构建应用镜像
build_application() {
    log_info "构建应用镜像..."
    
    # 构建后端
    log_info "构建Go后端..."
    docker-compose build api
    
    # 构建Worker
    log_info "构建Worker服务..."
    docker-compose build worker
    
    log_success "应用构建完成"
    echo ""
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 启动基础设施服务
    log_info "启动基础服务（MySQL, Redis, RabbitMQ）..."
    docker-compose up -d mysql redis rabbitmq
    
    # 等待MySQL就绪
    log_info "等待MySQL启动..."
    local retries=30
    while [ $retries -gt 0 ]; do
        if docker-compose exec -T mysql mysqladmin ping -h localhost -u root -p${db_password} &> /dev/null; then
            log_success "MySQL已就绪"
            break
        fi
        retries=$((retries - 1))
        if [ $retries -eq 0 ]; then
            log_error "MySQL启动超时"
            exit 1
        fi
        echo -n "."
        sleep 2
    done
    echo ""
    
    # 运行数据库迁移
    log_info "运行数据库迁移..."
    docker-compose run --rm api /app/bin/openwan migrate up
    
    # 启动应用服务
    log_info "启动应用服务..."
    docker-compose up -d api worker
    
    # 启动前端（如果需要）
    if [ -d "frontend" ]; then
        log_info "启动前端服务..."
        docker-compose up -d frontend
    fi
    
    log_success "所有服务启动完成"
    echo ""
}

# 初始化数据
initialize_data() {
    log_info "初始化默认数据..."
    
    # 创建管理员账号
    log_info "创建管理员账号..."
    docker-compose exec -T api /app/bin/openwan admin create \
        --username admin \
        --password admin123 \
        --email admin@openwan.local \
        || log_warn "管理员账号可能已存在"
    
    # 创建默认分类
    log_info "创建默认分类..."
    docker-compose exec -T api /app/bin/openwan category seed \
        || log_warn "分类可能已存在"
    
    log_success "数据初始化完成"
    echo ""
}

# 显示访问信息
show_access_info() {
    local ip=$(hostname -I | awk '{print $1}')
    
    cat << EOF
${GREEN}
╔═══════════════════════════════════════════════════════════╗
║                 部署成功！                                 ║
╚═══════════════════════════════════════════════════════════╝
${NC}

${BLUE}访问信息：${NC}

  📱 前端地址:    http://localhost:3000
                  http://${ip}:3000

  🔌 后端API:     http://localhost:8080
                  http://${ip}:8080

  📊 健康检查:    http://localhost:8080/health

  🔐 管理员账号:
     用户名: admin
     密码:   admin123

${BLUE}服务端口：${NC}

  MySQL:      3306
  Redis:      6379
  RabbitMQ:   5672 (管理界面: 15672)
  Sphinx:     9306

${BLUE}常用命令：${NC}

  查看日志:       docker-compose logs -f [service]
  重启服务:       docker-compose restart [service]
  停止所有服务:   docker-compose down
  查看服务状态:   docker-compose ps

${BLUE}下一步：${NC}

  1. 访问前端: http://localhost:3000
  2. 使用管理员账号登录
  3. 开始使用OpenWan！

${YELLOW}注意：${NC}
  - 这是开发环境配置，不适合生产使用
  - 生产部署请参考 docs/DEPLOYMENT.md
  - 数据库密码已保存在 .env 文件

${GREEN}祝您使用愉快！${NC}

EOF
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    local max_retries=30
    local retry=0
    
    while [ $retry -lt $max_retries ]; do
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            log_success "健康检查通过"
            return 0
        fi
        retry=$((retry + 1))
        echo -n "."
        sleep 2
    done
    
    log_error "健康检查失败"
    log_warn "服务可能需要更多时间启动，请稍后手动检查"
    return 1
}

# 主函数
main() {
    print_banner
    
    log_info "开始本地开发环境部署..."
    echo ""
    
    # 检查系统要求
    check_requirements
    
    # 停止已有容器
    stop_existing_containers
    
    # 可选：清理旧数据
    cleanup_old_data
    
    # 创建目录
    create_directories
    
    # 生成配置文件
    generate_env_file
    
    # 拉取镜像
    pull_docker_images
    
    # 构建应用
    build_application
    
    # 启动服务
    start_services
    
    # 初始化数据
    initialize_data
    
    # 健康检查
    health_check
    
    # 显示访问信息
    show_access_info
    
    log_success "本地开发环境部署完成！"
}

# 执行主函数
main "$@"
