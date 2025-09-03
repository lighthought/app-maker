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

# 安装 bmad-method
install_bmad_method() {
    local project_dir="$1"
    
    log_info "在项目目录中安装 bmad-method..."
    
    cd "$project_dir"
    
    # 检查是否已安装
    if [ -d "node_modules" ] && [ -f "package.json" ]; then
        log_warning "项目目录中已存在 node_modules，跳过安装"
        return 0
    fi
    
    # 初始化 package.json
    npm init -y
    
    # 安装 bmad-method
    npm install bmad-method
    
    log_success "bmad-method 安装完成"
}

# 安装 cursor-cli
install_cursor_cli() {
    local project_dir="$1"
    
    log_info "安装 cursor-cli..."
    
    # 检查是否已安装
    if command -v cursor &> /dev/null; then
        log_warning "cursor-cli 已安装"
        return 0
    fi
    
    # 使用 npm 全局安装 cursor-cli（更可靠的方式）
    log_info "使用 npm 安装 cursor-cli..."
    npm install -g @cursor/cli
    
    # 验证安装
    if command -v cursor &> /dev/null; then
        log_success "cursor-cli 安装完成"
        cursor --version
    else
        log_error "cursor-cli 安装失败"
        return 1
    fi
}

# 创建项目配置文件
create_project_config() {
    local project_dir="$1"
    local project_id="$2"
    
    log_info "创建项目配置文件..."
    
    cd "$project_dir"
    
    # 创建 .bmad-core 目录
    mkdir -p .bmad-core
    
    # 创建项目配置文件
    cat > .bmad-core/project-config.json << EOF
{
  "projectId": "$project_id",
  "createdAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "container",
  "tools": {
    "node": "$(node --version)",
    "npm": "$(npm --version)",
    "npx": "$(npx --version)"
  },
  "bmadMethod": {
    "installed": true,
    "version": "latest"
  },
  "cursorCli": {
    "installed": true,
    "version": "latest"
  }
}
EOF
    
    log_success "项目配置文件创建完成"
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
if [ ! -d "node_modules" ]; then
    echo "📦 安装项目依赖..."
    npm install
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
    install_cursor_cli "$project_dir"
    
    # 创建配置
    create_project_config "$project_dir" "$project_id"
    create_dev_scripts "$project_dir"
    
    log_success "项目开发环境设置完成！"
    echo ""
    echo "下一步操作："
    echo "1. 进入项目目录: cd $project_dir"
    echo "2. 启动开发环境: ./start-dev.sh"
    echo "3. 使用 Cursor CLI: cursor chat --project $project_dir"
}

# 执行主函数
main "$@"
