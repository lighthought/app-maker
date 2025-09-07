# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 安装 swag 工具并生成 API 文档
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 安装 Node.js、npm、Go 和其他必要工具
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    zip \
    nodejs \
    npm \
    curl \
    git \
    bash \
    python3 \
    make \
    g++ \
    go 
    #\
   # && npm install -g npx  // npx 现在是 npm 的一部分

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制生成的文档
COPY --from=builder /app/docs ./docs

# 复制模板文件到临时位置
COPY --from=builder /app/data/template.zip ./template.zip

# 复制脚本文件
COPY --from=builder /app/scripts ./scripts

# 创建必要的目录
RUN mkdir -p /app/logs /app/configs /app/bmad-config /app/data /app/data/tmp /app/data/projects /app/data/projects/cache /app/scripts && \
    chown -R appuser:appgroup /app && \
    chmod +x /app/scripts/*.sh

# 创建启动脚本，在容器启动时复制 template.zip
RUN echo '#!/bin/bash' > /app/start.sh && \
    echo 'if [ ! -f /app/data/template.zip ] && [ -f /app/template.zip ]; then' >> /app/start.sh && \
    echo '    echo "Copying template.zip to data directory..."' >> /app/start.sh && \
    echo '    cp /app/template.zip /app/data/template.zip' >> /app/start.sh && \
    echo 'fi' >> /app/start.sh && \
    echo './main' >> /app/start.sh && \
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
