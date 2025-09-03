#!/bin/bash

# 项目开发环境设置脚本
# 用于在项目目录中安装和配置 bmad-method 和 cursor-cli

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

# 检查必需的工具
check_requirements() {
    log_info "检查必需的工具..."
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装"
        exit 1
    fi
    
    # 检查 npm
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装"
        exit 1
    fi
    
    # 检查 npx
    if ! command -v npx &> /dev/null; then
        log_error "npx 未安装"
        exit 1
    fi
    
    log_success "所有必需工具已安装"
}

# 检查 bmad-method 是否已安装
check_bmad_installed() {
    local project_dir="$1"
    
    if [ -d "$project_dir/.bmad-core" ] && \
       [ -d "$project_dir/.bmad-core/agents" ] && \
       [ -d "$project_dir/.bmad-core/templates" ] && \
       [ -f "$project_dir/.bmad-core/core-config.yaml" ]; then
        log_success "bmad-method 已安装"
        return 0
    else
        log_info "bmad-method 未安装或安装不完整"
        return 1
    fi
}

# 检查后端项目是否已安装
check_backend_installed() {
    local project_dir="$1"
    
    if [ -d "$project_dir"/backend/docs ] && \
       [ -f "$project_dir"/backend/docs/swagger.yaml" ] && \
       [ -f "$project_dir"/backend/docs/docs.go" ]; then
        log_success "backend 项目已安装"
        return 0
    else
        log_info "backend 项目未安装或安装不完整"
        return 1
    fi
}

# 安装 bmad-method
install_bmad_method() {
    local project_dir="$1"
    
    log_info "在项目目录中安装 bmad-method..."
    
    cd "$project_dir"
    
    # 检查是否已安装
    if check_bmad_installed "$project_dir"; then
        log_warning "bmad-method 已安装，跳过安装"
        return 0
    fi
    
    # 初始化 package.json（如果不存在）
    if [ ! -f "package.json" ]; then
        log_info "初始化 package.json..."
        npm init -y
    fi
    
    # 安装 qwen-code 依赖
    log_info "安装 qwen-code 依赖..."
    npm i @qwen-code/qwen-code
    
    # 安装 bmad-method
    log_info "安装 bmad-method..."
    npx bmad-method install -f -i qwen-code -d .
    
    # 验证安装
    if check_bmad_installed "$project_dir"; then
        log_success "bmad-method 安装完成"
    else
        log_error "bmad-method 安装失败"
        return 1
    fi
}

# 初始化前端项目
setup_frontend_project() {
    local project_dir="$1"
    
    log_info "在项目目录中初始化前端项目..."
    
    cd "$project_dir"/frontend

    if [ -d "$project_dir/frontend/node_modules" ]; then
        log_warning "frontend 项目已安装，跳过安装"
        return 0
    fi

    log_info "安装前端项目依赖..."
    npm install
    
    # 检查是否已安装
    if [ -d "$project_dir"/frontend/node_modules ]; then
        log_warning "frontend 项目已安装"
    else
        log_error "frontend 项目安装失败"
        return 1
    fi
    return 0
}

# 初始化后端项目
setup_backend_project() {
    local project_dir="$1"
    
    log_info "在项目目录中初始化后端项目..."
    
    cd "$project_dir"/backend

    if check_backend_installed "$project_dir"; then
        log_warning "backend 项目已安装"
        return 0
    fi
    
    log_info "安装后端项目依赖..."
    go mod download

    log_info "安装 swagger 工具..."
    go install github.com/swaggo/swag/cmd/swag@latest

    log_info "构建后端项目..."
    go build -o server ./cmd/server
    
    return 0
}

# 创建开发脚本
create_dev_scripts() {
    local project_dir="$1"
    
    log_info "创建开发脚本..."
    
    cd "$project_dir"
    
    # 创建启动脚本
    cat > start-dev.sh << 'EOF'
#!/bin/bash

# 项目开发启动脚本

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR"

echo "🚀 启动项目开发环境..."
echo "项目目录: $PROJECT_DIR"

# 检查 bmad-method
if [ ! -d ".bmad-core" ] || [ ! -d ".bmad-core/agents" ]; then
    echo "📦 安装 bmad-method..."
    npm i @qwen-code/qwen-code
    npx bmad-method install -f -i qwen-code -d .
fi

# 启动 cursor-cli 聊天
echo "💬 启动 Cursor CLI 聊天..."
echo "项目ID: $(basename "$PROJECT_DIR")"
echo "使用命令: cursor chat --project $PROJECT_DIR"

# 这里可以添加更多的开发环境启动逻辑
echo "✅ 开发环境准备就绪"
EOF
    
    chmod +x start-dev.sh
    
    log_success "开发脚本创建完成"
}

# 主函数
main() {
    local project_dir="$1"
    local project_id="$2"
    
    if [ -z "$project_dir" ] || [ -z "$project_id" ]; then
        log_error "用法: $0 <项目目录> <项目ID>"
        exit 1
    fi
    
    if [ ! -d "$project_dir" ]; then
        log_error "项目目录不存在: $project_dir"
        exit 1
    fi
    
    log_info "开始设置项目开发环境..."
    log_info "项目目录: $project_dir"
    log_info "项目ID: $project_id"
    
    # 检查必需工具
    check_requirements
    
    # 安装工具
    install_bmad_method "$project_dir"

    # 初始化前端项目
    setup_frontend_project "$project_dir"

    # 初始化后端项目
    setup_backend_project "$project_dir"
    
    # 创建配置
    create_dev_scripts "$project_dir"
    
    log_success "项目开发环境设置完成！"
    echo ""
    echo "下一步操作："
    echo "1. 进入项目目录: cd $project_dir"
    echo "2. 启动开发环境: ./start-dev.sh"
    echo "3. 使用 qwen-code CLI: qwen-code chat --project $project_dir"
}

# 执行主函数
main "$@"
