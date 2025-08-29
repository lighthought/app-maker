# AutoCodeWeb Backend

AutoCodeWeb 后端服务，基于 Go + Gin + GORM + PostgreSQL + Redis 构建的多Agent协作系统。

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin 1.9+
- **ORM**: GORM 1.25+
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **配置管理**: Viper
- **日志**: Zap
- **认证**: JWT

## 项目结构

```
backend/
├── cmd/                    # 应用程序入口
│   └── server/            # 主服务入口
│       └── main.go
├── internal/               # 内部包
│   ├── api/               # API层
│   │   ├── handlers/      # HTTP处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes/        # 路由定义
│   ├── config/            # 配置管理
│   ├── database/          # 数据库相关
│   ├── models/            # 数据模型
│   ├── repositories/      # 数据访问层
│   ├── services/          # 业务逻辑层
│   └── worker/            # 后台工作进程
├── pkg/                    # 可导出的包
│   ├── auth/              # 认证相关
│   ├── bmad/              # BMad-Method集成
│   ├── cache/             # 缓存管理
│   └── logger/            # 日志管理
├── configs/                # 配置文件
├── scripts/                # 脚本文件
├── go.mod                 # Go模块文件
├── Dockerfile             # Docker 构建文件
├── .dockerignore          # Docker 忽略文件
└── README.md              # 项目说明
```

## 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose

### 安装依赖

```bash
go mod tidy
go mod download
```

### 配置

1. 启动开发环境数据库服务：
```bash
# 启动 PostgreSQL 和 Redis
docker-compose up -d

# 检查服务状态
docker-compose ps
```

2. 复制配置文件模板（可选）：
```bash
cp configs/config.yaml configs/config.local.yaml
```

3. 修改 `configs/config.local.yaml` 中的数据库和Redis配置（如果需要自定义）

### 运行

```bash
# 开发模式（推荐）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看后端日志
docker-compose logs -f backend

# 本地开发模式（需要先启动数据库服务）
go run cmd/server/main.go
```

### 构建

```bash
# 本地构建
go build -o bin/server cmd/server/main.go

# Docker 构建（开发环境）
docker-compose build backend

# Docker 构建（生产环境）
docker build -t autocodeweb-backend .
```

## API 文档

启动服务后，访问以下端点：

- 健康检查: `GET /api/v1/health`
- 用户注册: `POST /api/v1/auth/register`
- 用户登录: `POST /api/v1/auth/login`

## 开发

### 代码规范

- 使用 Go 官方代码规范
- 所有导出函数必须有注释
- 错误处理要明确
- 使用 context 进行超时控制

### 测试

```bash
go test ./...
```

### 清理

```bash
# 清理构建文件
rm -rf bin/

# 清理 Docker 镜像
docker rmi autocodeweb-backend

# 停止开发环境服务
docker-compose down

# 清理开发环境数据（谨慎使用）
docker-compose down -v
```

## 部署

项目支持 Docker 容器化部署，具体配置请参考 Dockerfile 文件。

### Docker 部署

```bash
# 构建镜像
docker build -t autocodeweb-backend .

# 运行容器
docker run -d \
  --name autocodeweb-backend \
  -p 8080:8080 \
  -e APP_ENVIRONMENT=production \
  -e DATABASE_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  autocodeweb-backend
```

### 生产环境配置

在生产环境中，请确保：
- 修改所有默认密钥和密码
- 配置正确的数据库和Redis连接信息
- 设置适当的日志级别
- 配置健康检查端点

## 许可证

MIT License
