#!/bin/bash

# 数据库迁移执行脚本
# 使用方法: ./run-migration.sh [migration-script-name]
# 例如: ./run-migration.sh migration-add-agent-question-fields.sql

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置（可通过环境变量覆盖）
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-app-maker-postgres-1}"
POSTGRES_USER="${POSTGRES_USER:-autocodeweb}"
POSTGRES_DB="${POSTGRES_DB:-autocodeweb}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 显示帮助信息
show_help() {
    echo -e "${BLUE}数据库迁移执行脚本${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 [migration-script-name]"
    echo ""
    echo "参数:"
    echo "  migration-script-name  要执行的迁移脚本文件名（可选）"
    echo ""
    echo "示例:"
    echo "  $0 migration-add-agent-question-fields.sql"
    echo "  $0  # 列出所有可用的迁移脚本"
    echo ""
    echo "环境变量:"
    echo "  POSTGRES_CONTAINER  PostgreSQL 容器名称 (默认: app-maker-postgres-1)"
    echo "  POSTGRES_USER       PostgreSQL 用户名 (默认: autocodeweb)"
    echo "  POSTGRES_DB         PostgreSQL 数据库名 (默认: autocodeweb)"
    echo ""
}

# 列出所有迁移脚本
list_migrations() {
    echo -e "${BLUE}可用的迁移脚本:${NC}"
    echo ""
    local count=0
    for script in "$SCRIPT_DIR"/migration-*.sql; do
        if [ -f "$script" ]; then
            count=$((count + 1))
            local basename=$(basename "$script")
            echo -e "${GREEN}[$count]${NC} $basename"
            # 提取脚本说明（第一行注释）
            local description=$(head -n 3 "$script" | grep "^--" | sed 's/^-- //' | tail -n 1)
            if [ -n "$description" ]; then
                echo -e "    ${YELLOW}$description${NC}"
            fi
            echo ""
        fi
    done
    
    if [ $count -eq 0 ]; then
        echo -e "${YELLOW}没有找到迁移脚本${NC}"
        return 1
    fi
}

# 检查 Docker 容器是否运行
check_container() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${POSTGRES_CONTAINER}$"; then
        echo -e "${RED}错误: PostgreSQL 容器 '${POSTGRES_CONTAINER}' 未运行${NC}"
        echo -e "${YELLOW}提示: 使用 'docker ps' 查看运行中的容器${NC}"
        echo -e "${YELLOW}或设置环境变量: export POSTGRES_CONTAINER=your-container-name${NC}"
        return 1
    fi
    echo -e "${GREEN}✓ PostgreSQL 容器已运行${NC}"
}

# 备份数据库
backup_database() {
    local backup_file="backup_$(date +%Y%m%d_%H%M%S).sql"
    echo -e "${YELLOW}正在备份数据库...${NC}"
    
    if docker exec "$POSTGRES_CONTAINER" pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > "$backup_file"; then
        echo -e "${GREEN}✓ 数据库备份成功: $backup_file${NC}"
        return 0
    else
        echo -e "${RED}✗ 数据库备份失败${NC}"
        return 1
    fi
}

# 执行迁移脚本
execute_migration() {
    local script_name=$1
    local script_path="$SCRIPT_DIR/$script_name"
    
    if [ ! -f "$script_path" ]; then
        echo -e "${RED}错误: 迁移脚本不存在: $script_path${NC}"
        return 1
    fi
    
    echo -e "${BLUE}准备执行迁移脚本: $script_name${NC}"
    echo ""
    
    # 显示脚本说明
    echo -e "${YELLOW}脚本说明:${NC}"
    head -n 5 "$script_path" | grep "^--" | sed 's/^-- //'
    echo ""
    
    # 确认执行
    read -p "$(echo -e ${YELLOW}是否继续执行迁移? [y/N]: ${NC})" -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}迁移已取消${NC}"
        return 1
    fi
    
    # 备份数据库
    if ! backup_database; then
        echo -e "${RED}备份失败，迁移已取消${NC}"
        return 1
    fi
    
    # 复制脚本到容器
    echo -e "${YELLOW}正在复制迁移脚本到容器...${NC}"
    docker cp "$script_path" "$POSTGRES_CONTAINER:/tmp/$script_name"
    
    # 执行迁移
    echo -e "${YELLOW}正在执行迁移脚本...${NC}"
    echo ""
    
    if docker exec -it "$POSTGRES_CONTAINER" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/tmp/$script_name"; then
        echo ""
        echo -e "${GREEN}✓ 迁移执行成功！${NC}"
        
        # 清理临时文件
        docker exec "$POSTGRES_CONTAINER" rm -f "/tmp/$script_name"
        
        return 0
    else
        echo ""
        echo -e "${RED}✗ 迁移执行失败${NC}"
        echo -e "${YELLOW}数据库备份文件: $backup_file${NC}"
        echo -e "${YELLOW}您可以使用备份文件恢复数据库${NC}"
        return 1
    fi
}

# 主函数
main() {
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}  数据库迁移工具${NC}"
    echo -e "${BLUE}======================================${NC}"
    echo ""
    
    # 检查参数
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_help
        exit 0
    fi
    
    # 检查容器
    if ! check_container; then
        exit 1
    fi
    
    echo ""
    
    # 如果没有提供脚本名称，列出所有迁移脚本
    if [ -z "$1" ]; then
        list_migrations
        echo ""
        echo -e "${YELLOW}请指定要执行的迁移脚本，例如:${NC}"
        echo -e "${YELLOW}  $0 migration-add-agent-question-fields.sql${NC}"
        exit 0
    fi
    
    # 执行迁移
    execute_migration "$1"
}

# 运行主函数
main "$@"

