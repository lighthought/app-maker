# 专业级优化的多阶段构建 Dockerfile
# 第一阶段：依赖下载阶段（最大化缓存利用）
FROM golang:1.24-alpine AS deps

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache ca-certificates git tzdata

# 复制 shared-models 目录
COPY shared-models/ ../shared-models/

# 复制 backend 的 go mod 文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖（这一层会被缓存）
RUN go mod download

# 第二阶段：工具安装阶段（固定版本，利用缓存）
FROM golang:1.24-alpine AS tools

# 安装必要的系统依赖
RUN apk add --no-cache ca-certificates git tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件用于工具安装
COPY backend/go.mod backend/go.sum ./

# 固定 Swagger 版本，避免每次都下载最新版本
# 使用具体版本号而不是 @latest，确保缓存有效
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

# 第三阶段：构建阶段
FROM tools AS builder

# 复制 shared-models 目录
COPY shared-models/ ../shared-models/

# 复制 go mod 文件（确保依赖正确）
COPY backend/go.mod backend/go.sum ./

# 复制源代码
COPY backend/ .

# 生成 API 文档（使用已安装的 swag 工具）
RUN swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 构建应用（静态链接，减小运行时依赖）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main cmd/server/main.go

# 第四阶段：最小化运行环境
FROM alpine:3.22.2 AS runtime

# 安装必要的运行时工具（包括 git）
RUN apk add --no-cache bash ca-certificates curl git openssh-client tzdata wget

# 创建必要的目录结构
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs /app/docs
COPY --from=builder /app/data/template.zip /app/template.zip
COPY --from=builder /app/scripts /app/scripts

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/logs /app/configs /app/data /app/data/tmp /app/data/projects /app/data/projects/cache /app/scripts /home/appuser/.ssh

# 设置权限和所有者
RUN chown -R appuser:appgroup /app
RUN chown -R appuser:appgroup /home/appuser/.ssh
RUN chmod 700 /home/appuser/.ssh

# 创建 known_hosts 文件并设置正确权限
RUN touch /home/appuser/.ssh/known_hosts && \
    chown appuser:appgroup /home/appuser/.ssh/known_hosts && \
    chmod 644 /home/appuser/.ssh/known_hosts

# 创建启动脚本
RUN echo '#!/bin/bash' > /app/start.sh && \
    echo 'if [ ! -f /app/data/template.zip ] && [ -f /app/template.zip ]; then' >> /app/start.sh && \
    echo '    echo "Copying template.zip to data directory..."' >> /app/start.sh && \
    echo '    cp /app/template.zip /app/data/template.zip' >> /app/start.sh && \
    echo 'fi' >> /app/start.sh && \
    echo 'exec ./main' >> /app/start.sh && \
    chmod +x /app/start.sh && \
    chown appuser:appgroup /app/start.sh

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 启动应用
CMD ["./start.sh"]
