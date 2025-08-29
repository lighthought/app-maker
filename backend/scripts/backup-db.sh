#!/bin/bash

# AutoCodeWeb 数据库备份脚本

# 配置变量
DB_NAME="autocodeweb"
DB_USER="autocodeweb"
DB_HOST="postgres"
DB_PORT="5432"
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/autocodeweb_${DATE}.sql"

# 创建备份目录
mkdir -p ${BACKUP_DIR}

# 执行备份
echo "开始备份数据库 ${DB_NAME}..."
pg_dump -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} > ${BACKUP_FILE}

if [ $? -eq 0 ]; then
    echo "数据库备份成功: ${BACKUP_FILE}"
    
    # 压缩备份文件
    gzip ${BACKUP_FILE}
    echo "备份文件已压缩: ${BACKUP_FILE}.gz"
    
    # 删除7天前的备份文件
    find ${BACKUP_DIR} -name "*.gz" -mtime +7 -delete
    echo "已清理7天前的备份文件"
else
    echo "数据库备份失败!"
    exit 1
fi
