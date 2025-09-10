# AutoCodeWeb 部署指南

## 概述

AutoCodeWeb 是一个全栈应用，包含前端（Vue.js + Vite）和后端（Go + Gin）。本文档介绍如何使用 Docker Compose 进行容器化部署。

## 架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (Nginx/Vite)  │◄──►│   (Go/Gin)      │◄──►│   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Cache         │
                    │   (Redis)       │
                    └─────────────────┘
```

## 快速开始

### 1. 环境准备

确保系统已安装：
- Docker
- Docker Compose
- Make (可选，用于简化命令)

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
vim .env
```

### 3. 开发环境部署

```bash
# 构建并启动开发环境
make build-dev
make run-dev

# 或者直接使用 docker-compose
docker-compose up -d
```

**访问地址：**
- 前端: http://localhost:3000
- 后端API: http://localhost:8098
- Swagger文档: http://localhost:8098/swagger/index.html

### 4. 生产环境部署

```bash
# 构建并启动生产环境
make build-prod
make run-prod

# 或者直接使用 docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

**访问地址：**
- 前端: http://localhost
- 后端API: http://localhost:8080
- Swagger文档: http://localhost:8080/swagger/index.html

## 常用命令

### 开发环境

```bash
# 启动服务
make run-dev

# 停止服务
make stop-dev

# 查看日志
make logs-dev
make logs-frontend-dev
make logs-backend-dev

# 重启服务
make restart-dev

# 进入容器
make shell-frontend-dev
make shell-backend-dev
```

### 生产环境

```bash
# 启动服务
make run-prod

# 停止服务
make stop-prod

# 查看日志
make logs-prod
make logs-frontend-prod
make logs-backend-prod

# 重启服务
make restart-prod

# 进入容器
make shell-frontend-prod
make shell-backend-prod
```

### 通用命令

```bash
# 健康检查
make health-check

# 清理资源
make clean-safe

# 运行测试
make test

# 代码检查
make lint

# 格式化代码
make fmt
```

## 环境配置

### 开发环境

- **前端**: Vite 开发服务器，支持热重载
- **后端**: Go 开发模式，支持热重载
- **数据库**: PostgreSQL 15，端口 5434
- **缓存**: Redis 7，端口 6379

### 生产环境

- **前端**: Nginx 静态文件服务
- **后端**: Go 生产模式
- **数据库**: PostgreSQL 15（内部网络）
- **缓存**: Redis 7（内部网络）

## 网络配置

### 开发环境端口映射

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 | 3000 | Vite 开发服务器 |
| 后端 | 8098 | API 服务 |
| 数据库 | 5434 | PostgreSQL |
| 缓存 | 6379 | Redis |

### 生产环境端口映射

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 | 80 | Nginx 静态服务 |
| 后端 | 8080 | API 服务 |

## API 代理配置

### 开发环境

前端通过 Vite 代理将 `/api` 请求转发到后端：

```javascript
// vite.config.ts
proxy: {
  '/api': {
    target: 'http://backend:8080',
    changeOrigin: true,
    rewrite: (path) => path.replace(/^\/api/, '')
  }
}
```

### 生产环境

前端通过 Nginx 代理将 `/api` 请求转发到后端：

```nginx
# nginx.conf
location /api/ {
    proxy_pass http://backend/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 数据库管理

```bash
# 数据库迁移
make db-migrate

# 种子数据
make db-seed

# 进入数据库
docker-compose exec autocodeweb-postgres psql -U autocodeweb -d autocodeweb
```

## 缓存管理

```bash
# 清理缓存
make cache-clear

# 查看缓存信息
make cache-info

# 进入 Redis
docker-compose exec autocodeweb-redis redis-cli
```

## 监控和日志

### 健康检查

```bash
# 检查服务状态
make health-check

# 手动检查
curl http://localhost:8098/api/v1/health
curl http://localhost:8098/api/v1/cache/health
```

### 日志查看

```bash
# 查看所有服务日志
make logs-dev

# 查看特定服务日志
make logs-frontend-dev
make logs-backend-dev

# 实时日志
docker-compose logs -f
```

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -ano | findstr :3000
   netstat -ano | findstr :8098
   
   # 停止冲突服务
   docker-compose down
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose ps autocodeweb-postgres
   
   # 查看数据库日志
   docker-compose logs autocodeweb-postgres
   ```

3. **前端无法访问后端**
   ```bash
   # 检查网络连接
   docker-compose exec frontend ping backend
   
   # 检查 API 代理
   curl http://localhost:3000/api/v1/health
   ```

### 重置环境

```bash
# 完全重置
make clean
make build-dev
make run-dev

# 安全重置（保留其他项目）
make clean-safe
make build-dev
make run-dev
```

## 性能优化

### 前端优化

- 启用 Gzip 压缩
- 静态资源缓存
- 代码分割

### 后端优化

- 连接池配置
- 缓存策略
- 日志级别

## 安全考虑

1. **生产环境**
   - 修改默认密码
   - 启用 HTTPS
   - 配置防火墙

2. **环境变量**
   - 使用强密钥
   - 定期轮换
   - 环境隔离

## 扩展部署

### 多实例部署

```bash
# 扩展后端服务
docker-compose up -d --scale backend=3

# 负载均衡配置
# 需要额外的 Nginx 配置
```

### 集群部署

- 使用 Docker Swarm
- 使用 Kubernetes
- 使用云服务

## 备份和恢复

### 数据库备份

```bash
# 备份数据库
docker-compose exec autocodeweb-postgres pg_dump -U autocodeweb autocodeweb > backup.sql

# 恢复数据库
docker-compose exec -T autocodeweb-postgres psql -U autocodeweb autocodeweb < backup.sql
```

### 配置文件备份

```bash
# 备份配置
tar -czf config-backup.tar.gz backend/configs/ .env

# 恢复配置
tar -xzf config-backup.tar.gz
```

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建
make build-dev
make run-dev

# 或者生产环境
make build-prod
make run-prod
```

## 联系支持

如有问题，请查看：
- 项目文档
- 日志文件
- 健康检查端点