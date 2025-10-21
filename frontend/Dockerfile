# 专业级优化的前端 Dockerfile - 开发环境
# 第一阶段：依赖安装阶段（最大化缓存利用）
FROM node:18-alpine AS deps

WORKDIR /app

# 复制 package.json 和 lock 文件（这一层会被缓存）
COPY package.json pnpm-lock.yaml ./

# 安装 pnpm 和项目依赖
RUN npm install -g pnpm@latest && \
    pnpm install --frozen-lockfile

# 第二阶段：构建阶段
FROM node:18-alpine AS builder

WORKDIR /app

# 复制依赖
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/package.json ./package.json
COPY --from=deps /app/pnpm-lock.yaml ./pnpm-lock.yaml

# 复制源代码
COPY . .

# 构建应用
RUN pnpm exec vite build

# 第三阶段：生产环境
FROM nginx:alpine AS production

# 安装必要的工具（最小化）
RUN apk add --no-cache wget

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 创建非 root 用户
RUN addgroup -g 1001 -S nginx && \
    adduser -u 1001 -S nginx -G nginx

# 设置权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d

# 创建必要的目录
RUN mkdir -p /var/cache/nginx/client_temp && \
    mkdir -p /var/cache/nginx/proxy_temp && \
    mkdir -p /var/cache/nginx/fastcgi_temp && \
    mkdir -p /var/cache/nginx/uwsgi_temp && \
    mkdir -p /var/cache/nginx/scgi_temp

# 切换到非 root 用户
USER nginx

# 暴露端口
EXPOSE 3000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]
