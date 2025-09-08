#!/bin/bash

# Jenkins 触发脚本
# 用于触发 Jenkins 对指定项目执行构建和部署

set -e

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
    echo "Jenkins 触发脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --user-id USER_ID        用户ID [必需]"
    echo "  --project-id PROJECT_ID 项目ID [必需]"
    echo "  --project-path PATH     项目路径 [必需]"
    echo "  --build-type TYPE       构建类型 (dev|prod) [可选，默认: dev]"
    echo "  --jenkins-url URL       Jenkins URL [可选，默认: http://10.0.0.6:5016]"
    echo "  --job-name NAME         Jenkins 任务名称 [可选，默认: app-maker-flow]"
    echo "  --username USER         Jenkins 用户名 [可选，默认: admin]"
    echo "  --api-token TOKEN       Jenkins API Token [可选]"
    echo "  -h, --help              显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 --user-id USER123 --project-id PROJ456 --project-path /app/data/projects/USER123/PROJ456"
}

# 参数解析
USER_ID=""
PROJECT_ID=""
PROJECT_PATH=""
BUILD_TYPE="dev"
JENKINS_URL="http://10.0.0.6:5016"
JOB_NAME="app-maker-flow"
USERNAME="admin"
API_TOKEN=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --user-id)
            USER_ID="$2"
            shift 2
            ;;
        --project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        --project-path)
            PROJECT_PATH="$2"
            shift 2
            ;;
        --build-type)
            BUILD_TYPE="$2"
            shift 2
            ;;
        --jenkins-url)
            JENKINS_URL="$2"
            shift 2
            ;;
        --job-name)
            JOB_NAME="$2"
            shift 2
            ;;
        --username)
            USERNAME="$2"
            shift 2
            ;;
        --api-token)
            API_TOKEN="$2"
            shift 2
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
if [[ -z "$USER_ID" ]]; then
    log_error "必须指定用户ID (--user-id)"
    show_help
    exit 1
fi

if [[ -z "$PROJECT_ID" ]]; then
    log_error "必须指定项目ID (--project-id)"
    show_help
    exit 1
fi

if [[ -z "$PROJECT_PATH" ]]; then
    log_error "必须指定项目路径 (--project-path)"
    show_help
    exit 1
fi

if [[ "$BUILD_TYPE" != "dev" && "$BUILD_TYPE" != "prod" ]]; then
    log_error "构建类型必须是 'dev' 或 'prod'"
    exit 1
fi

log_info "开始触发 Jenkins 构建"
log_info "用户ID: $USER_ID"
log_info "项目ID: $PROJECT_ID"
log_info "项目路径: $PROJECT_PATH"
log_info "构建类型: $BUILD_TYPE"
log_info "Jenkins URL: $JENKINS_URL"
log_info "任务名称: $JOB_NAME"

# 检查项目路径是否存在
if [[ ! -d "$PROJECT_PATH" ]]; then
    log_error "项目路径不存在: $PROJECT_PATH"
    log_info "请检查以下路径:"
    log_info "  - 容器内路径: /app/data/projects/$USER_ID/$PROJECT_ID"
    log_info "  - 主机路径: $PROJECT_PATH"
    exit 1
fi

# 检查项目目录中是否有 Makefile
if [[ ! -f "$PROJECT_PATH/Makefile" ]]; then
    log_warning "项目目录中没有找到 Makefile: $PROJECT_PATH/Makefile"
    log_info "项目目录内容:"
    ls -la "$PROJECT_PATH" || true
fi

# 构建 Jenkins API URL
JENKINS_API_URL="$JENKINS_URL/job/$JOB_NAME/buildWithParameters"

# 构建参数
PARAMS="USER_ID=$USER_ID&PROJECT_ID=$PROJECT_ID&PROJECT_PATH=$PROJECT_PATH&BUILD_TYPE=$BUILD_TYPE"

log_info "发送 Jenkins 构建请求..."
log_info "API URL: $JENKINS_API_URL"
log_info "参数: $PARAMS"

# 发送请求
if [[ -n "$API_TOKEN" ]]; then
    # 使用 API Token 认证
    RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/jenkins_response.txt \
        -X POST \
        -u "$USERNAME:$API_TOKEN" \
        "$JENKINS_API_URL?$PARAMS")
else
    # 不使用认证（如果 Jenkins 允许匿名访问）
    RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/jenkins_response.txt \
        -X POST \
        "$JENKINS_API_URL?$PARAMS")
fi

HTTP_CODE="${RESPONSE: -3}"

if [[ "$HTTP_CODE" -ge 200 && "$HTTP_CODE" -lt 300 ]]; then
    log_success "Jenkins 构建触发成功 (HTTP $HTTP_CODE)"
    
    # 显示响应内容（如果有）
    if [[ -f /tmp/jenkins_response.txt ]]; then
        RESPONSE_CONTENT=$(cat /tmp/jenkins_response.txt)
        if [[ -n "$RESPONSE_CONTENT" ]]; then
            log_info "响应内容: $RESPONSE_CONTENT"
        fi
    fi
    
    log_info "项目 $PROJECT_ID 的构建任务已提交到 Jenkins"
    log_info "可以在 Jenkins 控制台中查看构建进度: $JENKINS_URL/job/$JOB_NAME/"
    
else
    log_error "Jenkins 构建触发失败 (HTTP $HTTP_CODE)"
    
    # 显示错误响应
    if [[ -f /tmp/jenkins_response.txt ]]; then
        ERROR_CONTENT=$(cat /tmp/jenkins_response.txt)
        if [[ -n "$ERROR_CONTENT" ]]; then
            log_error "错误响应: $ERROR_CONTENT"
        fi
    fi
    
    exit 1
fi

# 清理临时文件
rm -f /tmp/jenkins_response.txt

log_success "Jenkins 触发脚本执行完成"
