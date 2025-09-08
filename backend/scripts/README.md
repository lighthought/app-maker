# 数据库脚本说明

本目录包含 AutoCodeWeb 项目的数据库相关脚本。

## 文件说明

### 1. init-db.sql
PostgreSQL 数据库初始化脚本，包含：
- 创建数据库和表结构
- 设置索引和约束
- 创建触发器
- 插入默认数据

**使用方法：**
```bash
# 连接到 PostgreSQL
psql -U postgres -h localhost

# 执行初始化脚本
\i scripts/init-db.sql
```

### 2. backup-db.sh (Linux/Mac)
Linux/Mac 系统的数据库备份脚本，包含：
- 自动备份数据库
- 压缩备份文件
- 清理过期备份（7天前）

**使用方法：**
```bash
# 添加执行权限
chmod +x scripts/backup-db.sh

# 执行备份
./scripts/backup-db.sh
```

### 3. backup-db.bat (Windows)
Windows 系统的数据库备份脚本，包含：
- 自动备份数据库
- 支持压缩（需要安装 7-Zip）

**使用方法：**
```cmd
# 双击执行或在命令行中运行
scripts\backup-db.bat
```

### 4. test-db-connection.go
数据库连接测试脚本，用于：
- 测试数据库连接
- 检查数据库版本
- 验证必要的扩展

**使用方法：**
```bash
# 运行测试
go run scripts/test-db-connection.go
```

## 数据库配置

### 默认配置
- **主机**: localhost
- **端口**: 5432
- **用户名**: postgres
- **密码**: password
- **数据库名**: autocodeweb

### 环境变量覆盖
可以通过环境变量覆盖默认配置：
```bash
export DATABASE_HOST=your-db-host
export DATABASE_PORT=5432
export DATABASE_USER=your-username
export DATABASE_PASSWORD=your-password
export DATABASE_NAME=your-db-name
```

## 数据库结构

### 主要表
1. **users** - 用户表
2. **projects** - 项目表（包含开发状态跟踪）


### 项目表字段说明
- **dev_status**: 开发子状态（pending, environment_processing, prd_generating 等）
- **dev_progress**: 开发进度（0-100）
- **current_task_id**: 当前执行的任务ID
- **requirements**: 项目需求（NOT NULL）
- **project_path**: 项目路径（NOT NULL）

### 任务表字段说明
- **type**: 任务类型（project_development, build, deploy 等）
- **status**: 任务状态（pending, in_progress, completed, failed, cancelled）
- **priority**: 任务优先级（0-9，0为最高优先级）
- **description**: 任务描述
- **started_at**: 开始时间
- **completed_at**: 完成时间

### 扩展
- **uuid-ossp**: 用于生成 UUID
- **pgcrypto**: 用于密码加密

### 索引
- 用户邮箱、角色、状态索引
- 项目用户ID、状态、创建时间索引
- 任务项目ID、状态、类型、优先级索引
- 标签名称索引

## 注意事项

1. **权限**: 确保 PostgreSQL 用户有创建数据库和扩展的权限
2. **扩展**: 初始化脚本会自动创建必要的扩展
3. **备份**: 建议定期执行备份脚本
4. **安全**: 生产环境中请修改默认密码和配置

## 故障排除

### 常见问题
1. **连接失败**: 检查 PostgreSQL 服务是否运行
2. **权限不足**: 确保用户有足够权限
3. **扩展不存在**: 某些扩展可能需要管理员权限

### 日志查看
```bash
# 查看 PostgreSQL 日志
tail -f /var/log/postgresql/postgresql-*.log
```
