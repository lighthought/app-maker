# 开发环境 Dockerfile - 使用 Nginx 代理
FROM node:18-alpine AS builder

WORKDIR /app

# 复制 package.json 和 lock 文件
COPY package.json pnpm-lock.yaml ./

# 安装依赖
RUN npm install -g pnpm && pnpm install

# 复制源代码
COPY . .

# 构建应用
RUN pnpm exec vite build

# 生产环境阶段
FROM nginx:alpine

# 复制构建产物到 nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 暴露端口
EXPOSE 3000

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]