#!/bin/bash

# AutoCodeWeb Backend Jenkins 自动化构建脚本
# 支持开发环境和生产环境的自动化构建

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

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "AutoCodeWeb Backend Jenkins 自动化构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -e, --environment ENV    构建环境 (dev|prod) [必需]"
    echo "  -t, --tag TAG            镜像标签 [可选，默认: latest]"
    echo "  -p, --push               推送镜像到仓库 [可选]"
    echo "  -r, --registry REGISTRY  镜像仓库地址 [可选，默认: localhost:5000]"
    echo "  -c, --clean              构建前清理 [可选]"
    echo "  -h, --help               显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -e dev                构建开发环境镜像"
    echo "  $0 -e prod -t v1.0.0    构建生产环境镜像，标签为v1.0.0"
    echo "  $0 -e prod -t v1.0.0 -p 构建并推送生产环境镜像"
    echo "  $0 -e dev -c            清理后构建开发环境镜像"
}

# 参数解析
ENVIRONMENT=""
IMAGE_TAG="latest"
PUSH_IMAGE=false
REGISTRY="localhost:5000"
CLEAN_BEFORE_BUILD=false

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
        -p|--push)
            PUSH_IMAGE=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -c|--clean)
            CLEAN_BEFORE_BUILD=true
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
    log_error "必须指定构建环境 (-e 或 --environment)"
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

log_info "开始构建 AutoCodeWeb Backend"
log_info "环境: $ENVIRONMENT"
log_info "镜像标签: $IMAGE_TAG"
log_info "镜像名称: $FULL_IMAGE_NAME"
log_info "推送镜像: $PUSH_IMAGE"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    log_error "Docker 未运行或无法访问"
    exit 1
fi

# 清理构建缓存（如果需要）
if [[ "$CLEAN_BEFORE_BUILD" == true ]]; then
    log_info "清理构建缓存..."
    docker system prune -f
    docker builder prune -f
fi

# 生成Swagger文档
log_info "生成Swagger文档..."
if command -v swag > /dev/null 2>&1; then
    swag init -g cmd/server/main.go -o docs
    log_success "Swagger文档生成完成"
else
    log_warning "swag 工具未安装，跳过Swagger文档生成"
fi

# 根据环境选择Dockerfile
if [[ "$ENVIRONMENT" == "prod" ]]; then
    DOCKERFILE="Dockerfile.prod"
else
    DOCKERFILE="Dockerfile"
fi

# 构建镜像
log_info "开始构建Docker镜像..."
log_info "使用Dockerfile: $DOCKERFILE"

if docker build -f "$DOCKERFILE" -t "$FULL_IMAGE_NAME" .; then
    log_success "镜像构建成功: $FULL_IMAGE_NAME"
else
    log_error "镜像构建失败"
    exit 1
fi

# 推送镜像（如果需要）
if [[ "$PUSH_IMAGE" == true ]]; then
    log_info "推送镜像到仓库..."
    
    # 检查仓库是否可访问
    if ! docker push "$FULL_IMAGE_NAME" > /dev/null 2>&1; then
        log_error "无法推送镜像到仓库: $REGISTRY"
        log_error "请检查仓库地址和网络连接"
        exit 1
    fi
    
    log_success "镜像推送成功: $FULL_IMAGE_NAME"
fi

# 运行健康检查（可选）
if [[ "$ENVIRONMENT" == "dev" ]]; then
    log_info "启动开发环境进行健康检查..."
    
    # 停止现有容器
    docker-compose down > /dev/null 2>&1 || true
    
    # 启动服务
    if docker-compose up -d; then
        log_info "等待服务启动..."
        sleep 30
        
        # 检查服务健康状态
        if docker-compose ps | grep -q "healthy"; then
            log_success "服务健康检查通过"
        else
            log_warning "服务健康检查失败，请检查日志"
            docker-compose logs backend
        fi
        
        # 停止服务
        docker-compose down
    else
        log_error "无法启动开发环境"
        exit 1
    fi
fi

log_success "构建完成！"
log_info "镜像: $FULL_IMAGE_NAME"
log_info "构建时间: $(date)"

# 显示镜像信息
log_info "镜像信息:"
docker images "$FULL_IMAGE_NAME"
