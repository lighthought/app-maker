# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 shared-models 目录
COPY shared-models/ ../shared-models/

# 复制 backend 的 go mod 文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制 backend 源代码
COPY backend/ .

# 安装 swag 工具并生成 API 文档
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 安装 ca-certificates、tzdata、SSH 客户端和密钥生成工具
RUN apk --no-cache add ca-certificates tzdata zip bash git curl openssh-client

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/logs /app/configs /app/data /app/data/tmp /app/data/projects /app/data/projects/cache /app/scripts /home/appuser/.ssh

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制生成的文档
COPY --from=builder /app/docs ./docs

# 复制模板文件到临时位置
COPY --from=builder /app/data/template.zip ./template.zip

# 复制脚本文件
COPY --from=builder /app/scripts ./scripts

# 设置权限和所有者
RUN chown -R appuser:appgroup /app && \
    chown -R appuser:appgroup /home/appuser/.ssh && \
    chmod 700 /home/appuser/.ssh

# 设置脚本执行权限（如果脚本存在）
RUN if [ -d /app/scripts ] && [ "$(ls -A /app/scripts/*.sh 2>/dev/null)" ]; then \
        chmod +x /app/scripts/*.sh; \
    fi

# 创建 known_hosts 文件并设置正确权限
RUN touch /home/appuser/.ssh/known_hosts && \
    chown appuser:appgroup /home/appuser/.ssh/known_hosts && \
    chmod 644 /home/appuser/.ssh/known_hosts
    
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
