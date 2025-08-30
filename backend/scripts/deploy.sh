#!/bin/bash

# AutoCodeWeb Backend 部署脚本
# 支持开发环境和生产环境的自动化部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "AutoCodeWeb Backend 部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -e, --environment ENV    部署环境 (dev|prod) [必需]"
    echo "  -t, --tag TAG            镜像标签 [可选，默认: latest]"
    echo "  -r, --registry REGISTRY  镜像仓库地址 [可选，默认: localhost:5000]"
    echo "  -c, --config CONFIG      配置文件路径 [可选]"
    echo "  -f, --force              强制重新部署 [可选]"
    echo "  -h, --help               显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -e dev                部署开发环境"
    echo "  $0 -e prod -t v1.0.0    部署生产环境，使用v1.0.0标签"
    echo "  $0 -e prod -f            强制重新部署生产环境"
}

# 参数解析
ENVIRONMENT=""
IMAGE_TAG="latest"
REGISTRY="localhost:5000"
CONFIG_FILE=""
FORCE_DEPLOY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -f|--force)
            FORCE_DEPLOY=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证必需参数
if [[ -z "$ENVIRONMENT" ]]; then
    log_error "必须指定部署环境 (-e 或 --environment)"
    show_help
    exit 1
fi

if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
    log_error "环境必须是 'dev' 或 'prod'"
    exit 1
fi

# 设置变量
PROJECT_NAME="autocodeweb-backend"
IMAGE_NAME="${PROJECT_NAME}-${ENVIRONMENT}"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

log_info "开始部署 AutoCodeWeb Backend"
log_info "环境: $ENVIRONMENT"
log_info "镜像: $FULL_IMAGE_NAME"
log_info "强制部署: $FORCE_DEPLOY"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    log_error "Docker 未运行或无法访问"
    exit 1
fi

# 检查镜像是否存在
if ! docker images "$FULL_IMAGE_NAME" | grep -q "$IMAGE_TAG"; then
    log_error "镜像不存在: $FULL_IMAGE_NAME"
    log_info "请先构建镜像或检查镜像标签"
    exit 1
fi

# 根据环境选择配置文件
if [[ -z "$CONFIG_FILE" ]]; then
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        COMPOSE_FILE="docker-compose.prod.yml"
        ENV_FILE=".env.prod"
    else
        COMPOSE_FILE="docker-compose.yml"
        ENV_FILE=".env"
    fi
else
    COMPOSE_FILE="$CONFIG_FILE"
fi

# 检查配置文件是否存在
if [[ ! -f "$COMPOSE_FILE" ]]; then
    log_error "配置文件不存在: $COMPOSE_FILE"
    exit 1
fi

# 停止现有服务
log_info "停止现有服务..."
if [[ "$ENVIRONMENT" == "prod" ]]; then
    docker-compose -f "$COMPOSE_FILE" down || true
else
    docker-compose down || true
fi

# 清理旧镜像（如果需要）
if [[ "$FORCE_DEPLOY" == true ]]; then
    log_info "清理旧镜像..."
    docker image prune -f
fi

# 拉取最新镜像（如果需要）
if [[ "$ENVIRONMENT" == "prod" ]]; then
    log_info "拉取最新镜像..."
    docker pull "$FULL_IMAGE_NAME" || log_warning "无法拉取镜像，使用本地镜像"
fi

# 启动服务
log_info "启动服务..."
if [[ "$ENVIRONMENT" == "prod" ]]; then
    if [[ -f "$ENV_FILE" ]]; then
        docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
    else
        docker-compose -f "$COMPOSE_FILE" up -d
    fi
else
    docker-compose up -d
fi

# 等待服务启动
log_info "等待服务启动..."
sleep 30

# 检查服务状态
log_info "检查服务状态..."
if [[ "$ENVIRONMENT" == "prod" ]]; then
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        log_success "服务启动成功"
    else
        log_error "服务启动失败"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi
else
    if docker-compose ps | grep -q "Up"; then
        log_success "服务启动成功"
    else
        log_error "服务启动失败"
        docker-compose logs
        exit 1
    fi
fi

# 健康检查
log_info "执行健康检查..."
sleep 10

# 检查健康状态
if [[ "$ENVIRONMENT" == "prod" ]]; then
    HEALTH_STATUS=$(docker-compose -f "$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Status}}" | grep backend)
else
    HEALTH_STATUS=$(docker-compose ps --format "table {{.Name}}\t{{.Status}}" | grep backend)
fi

if echo "$HEALTH_STATUS" | grep -q "healthy"; then
    log_success "健康检查通过"
elif echo "$HEALTH_STATUS" | grep -q "Up"; then
    log_warning "服务已启动但健康检查未完成"
else
    log_error "健康检查失败"
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        docker-compose -f "$COMPOSE_FILE" logs backend
    else
        docker-compose logs backend
    fi
    exit 1
fi

# 显示服务信息
log_info "部署完成！"
log_info "服务信息:"

if [[ "$ENVIRONMENT" == "prod" ]]; then
    docker-compose -f "$COMPOSE_FILE" ps
    log_info "生产环境访问地址: http://localhost:8080"
    log_info "Swagger文档: http://localhost:8080/swagger/index.html"
else
    docker-compose ps
    log_info "开发环境访问地址: http://localhost:8098"
    log_info "Swagger文档: http://localhost:8098/swagger/index.html"
fi

log_success "部署成功完成！"
