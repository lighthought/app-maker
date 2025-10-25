# AutoCodeWeb Backend 部署指南

## 概述

本文档描述了 AutoCodeWeb Backend 的自动化构建和部署流程，包括开发环境和生产环境的配置。

## 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [自动化构建](#自动化构建)
- [Jenkins 集成](#jenkins-集成)
- [健康检查](#健康检查)
- [部署脚本](#部署脚本)
- [故障排除](#故障排除)

## 环境要求

### 必需软件

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.23+
- Make (可选，用于简化命令)

### 可选软件

- Jenkins (用于CI/CD)
- Docker Registry (用于镜像存储)

## 快速开始

### 开发环境

```bash
# 构建并启动开发环境
make build-dev-docker
make run-dev-docker

# 访问服务
# API: http://localhost:8098
# Swagger: http://localhost:8098/swagger/index.html
```

### 生产环境

```bash
# 构建生产环境镜像
make build-prod

# 部署生产环境
make deploy ENV=prod TAG=latest

# 访问服务
# API: http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
```

## 自动化构建

### 使用 Makefile

```bash
# 构建开发环境
make build-dev-docker

# 构建生产环境
make build-prod

# 生成Swagger文档
make swagger
```

### 使用构建脚本

```bash
# 构建开发环境
./scripts/jenkins-build.sh -e dev

# 构建生产环境并推送镜像
./scripts/jenkins-build.sh -e prod -t v1.0.0 -p

# 构建前清理缓存
./scripts/jenkins-build.sh -e dev -c
```

### 构建脚本参数

| 参数 | 说明 | 必需 | 默认值 |
|------|------|------|--------|
| `-e, --environment` | 构建环境 (dev/prod) | 是 | - |
| `-t, --tag` | 镜像标签 | 否 | latest |
| `-p, --push` | 推送镜像到仓库 | 否 | false |
| `-r, --registry` | 镜像仓库地址 | 否 | localhost:5000 |
| `-c, --clean` | 构建前清理缓存 | 否 | false |

## Jenkins 集成

### Jenkinsfile 配置

项目根目录包含 `Jenkinsfile`，支持以下功能：

- 多环境构建 (dev/prod)
- 参数化构建
- 自动化测试
- 代码质量检查
- 健康检查
- 镜像推送

### Jenkins 构建参数

| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `BUILD_ENVIRONMENT` | Choice | 构建环境 | dev |
| `IMAGE_TAG` | String | 镜像标签 | 自动生成 |
| `PUSH_IMAGE` | Boolean | 推送镜像 | false |
| `RUN_TESTS` | Boolean | 运行测试 | true |
| `CLEAN_BUILD` | Boolean | 清理构建 | false |

### Jenkins 流水线阶段

1. **Checkout**: 检出代码
2. **Setup Environment**: 设置构建环境
3. **Install Dependencies**: 安装依赖
4. **Generate Documentation**: 生成API文档
5. **Run Tests**: 运行测试
6. **Code Quality Check**: 代码质量检查
7. **Build Docker Image**: 构建Docker镜像
8. **Health Check**: 健康检查
9. **Push Image**: 推送镜像
10. **Deploy**: 部署

## 健康检查

### 健康检查端点

- **应用健康检查**: `/api/v1/health`
- **缓存健康检查**: `/api/v1/cache/health`

### Docker 健康检查配置

```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 手动健康检查

```bash
# 检查开发环境
make health-check ENV=dev

# 检查生产环境
make health-check ENV=prod

# 直接调用健康检查API
curl http://localhost:8098/api/v1/health
curl http://localhost:8098/api/v1/cache/health
```

## 部署脚本

### 使用 Makefile

```bash
# 部署开发环境
make deploy ENV=dev

# 部署生产环境
make deploy ENV=prod TAG=v1.0.0

# 强制重新部署
make deploy ENV=prod FORCE=true
```

### 使用部署脚本

```bash
# 部署开发环境
./scripts/deploy.sh -e dev

# 部署生产环境
./scripts/deploy.sh -e prod -t v1.0.0

# 强制重新部署
./scripts/deploy.sh -e prod -f
```

### 部署脚本参数

| 参数 | 说明 | 必需 | 默认值 |
|------|------|------|--------|
| `-e, --environment` | 部署环境 (dev/prod) | 是 | - |
| `-t, --tag` | 镜像标签 | 否 | latest |
| `-r, --registry` | 镜像仓库地址 | 否 | localhost:5000 |
| `-c, --config` | 配置文件路径 | 否 | 自动选择 |
| `-f, --force` | 强制重新部署 | 否 | false |

## 环境配置

### 开发环境配置

- 端口: 8098
- 数据库: PostgreSQL (5434)
- 缓存: Redis (6379)
- 日志级别: debug

### 生产环境配置

- 端口: 8080
- 数据库: PostgreSQL (环境变量)
- 缓存: Redis (环境变量)
- 日志级别: warn

### 环境变量

生产环境需要以下环境变量：

```bash
# 应用配置
APP_ENVIRONMENT=production
APP_SECRET_KEY=your-secret-key

# 数据库配置
DATABASE_HOST=your-db-host
DATABASE_USER=your-db-user
DATABASE_PASSWORD=your-db-password
DATABASE_NAME=your-db-name

# Redis配置
REDIS_HOST=your-redis-host
REDIS_PASSWORD=your-redis-password

# JWT配置
JWT_SECRET_KEY=your-jwt-secret
```

## 故障排除

### 常见问题

#### 1. 构建失败

```bash
# 检查Docker是否运行
docker info

# 清理构建缓存
docker system prune -f

# 重新构建
make build-dev-docker
```

#### 2. 服务启动失败

```bash
# 检查容器状态
docker-compose ps

# 查看日志
docker-compose logs backend

# 检查端口占用
netstat -tulpn | grep 8098
```

#### 3. 健康检查失败

```bash
# 检查服务是否响应
curl http://localhost:8098/api/v1/health

# 检查数据库连接
docker-compose exec backend wget -qO- http://localhost:8080/api/v1/health

# 重启服务
docker-compose restart backend
```

#### 4. 镜像推送失败

```bash
# 检查仓库连接
docker push localhost:5000/test

# 检查网络连接
ping localhost:5000

# 使用不同的仓库地址
./scripts/jenkins-build.sh -e prod -r your-registry.com
```

### 日志查看

```bash
# 查看实时日志
make logs-dev

# 查看生产环境日志
make logs-prod

# 查看特定服务日志
docker-compose logs -f backend
```

### 性能监控

```bash
# 查看容器资源使用
docker stats

# 查看系统资源
htop

# 查看网络连接
netstat -tulpn
```

## 最佳实践

### 1. 镜像标签管理

- 使用语义化版本号 (v1.0.0)
- 开发环境使用构建号 (dev-123)
- 生产环境使用版本号 (prod-v1.0.0)

### 2. 安全配置

- 使用非root用户运行容器
- 定期更新基础镜像
- 使用环境变量管理敏感信息

### 3. 监控和告警

- 配置健康检查
- 监控服务状态
- 设置日志轮转

### 4. 备份策略

- 定期备份数据库
- 备份配置文件
- 测试恢复流程

## 联系支持

如有问题，请联系开发团队或查看项目文档。
