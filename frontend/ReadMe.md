# AutoCodeWeb 前端项目

## 项目简介

AutoCodeWeb 是一个基于 Vue.js 3 + TypeScript + Naive UI 的现代化前端项目，支持多 Agent 协作的自动代码生成平台。项目采用组件化开发，响应式设计，为用户提供直观、高效的项目创建和管理体验。

## 技术栈

### 核心框架
- **Vue.js 3.4+** - 渐进式 JavaScript 框架
- **TypeScript 5.2+** - 类型安全的 JavaScript 超集
- **Vite 5.0+** - 下一代前端构建工具

### UI 组件库
- **Naive UI 2.38+** - Vue 3 组件库，支持 TypeScript

### 状态管理
- **Pinia 2.1+** - Vue 3 官方推荐的状态管理库

### 路由管理
- **Vue Router 4.2+** - Vue.js 官方路由管理器

### HTTP 客户端
- **Axios 1.6+** - 基于 Promise 的 HTTP 客户端

### 样式系统
- **SCSS** - CSS 预处理器
- **CSS Variables** - 主题变量系统
- **Glassmorphism** - 玻璃拟态设计风格

### 开发工具
- **@vueuse/core** - Vue 组合式 API 工具集
- **@iconify/vue** - 图标库

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── assets/            # 静态资源（图片、字体等）
│   ├── components/        # 通用组件
│   │   ├── common/        # 基础组件
│   │   ├── layout/        # 布局组件
│   │   └── business/      # 业务组件
│   ├── composables/       # 组合式函数
│   ├── config/            # 配置文件
│   ├── directives/        # 自定义指令
│   ├── hooks/             # 自定义Hooks
│   ├── layouts/           # 页面布局
│   ├── pages/             # 页面组件
│   ├── router/            # 路由配置
│   ├── stores/            # 状态管理
│   ├── styles/            # 样式文件
│   ├── types/             # TypeScript类型定义
│   ├── utils/             # 工具函数
│   └── views/             # 视图组件
├── .env                   # 环境变量
├── .env.development      # 开发环境变量
├── .env.production       # 生产环境变量
├── package.json           # 依赖配置
├── tsconfig.json          # TypeScript配置
├── vite.config.ts         # Vite配置
└── index.html             # 入口HTML
```

## 快速开始

### 环境要求
- Node.js 18.0+
- pnpm 8.0+ (推荐) 或 npm 8.0+

### 安装依赖
```bash
# 使用 pnpm (推荐)
pnpm install

# 或使用 npm
npm install
```

### 开发环境
```bash
# 启动开发服务器
pnpm dev

# 或使用 npm
npm run dev
```

开发服务器将在 `http://localhost:3000` 启动

### 构建生产版本
```bash
# 构建生产版本
pnpm build

# 预览生产版本
pnpm preview
```

### 类型检查
```bash
# 运行 TypeScript 类型检查
pnpm type-check
```

## 开发指南

### 组件开发规范

#### 组件命名
- 使用 PascalCase 命名组件
- 文件名与组件名保持一致
- 组件目录按功能分类：common、layout、business

#### 组件结构
```vue
<template>
  <!-- 模板内容 -->
</template>

<script setup lang="ts">
// 组件逻辑
</script>

<style scoped>
/* 组件样式 */
</style>
```

#### TypeScript 类型定义
```typescript
// 组件 Props 类型定义
interface Props {
  title: string
  count?: number
}

const props = withDefaults(defineProps<Props>(), {
  count: 0
})

// 组件 Emits 类型定义
const emit = defineEmits<{
  update: [value: string]
  delete: [id: number]
}>()
```

### 状态管理

#### Pinia Store 结构
```typescript
// stores/user.ts
export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string>('')
  
  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  
  // 动作
  const login = async (credentials: LoginCredentials) => {
    // 登录逻辑
  }
  
  return {
    user, token, isLoggedIn, login
  }
})
```

### 路由配置

#### 路由结构
```typescript
// router/index.ts
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/pages/Home.vue'),
        meta: { title: '首页', requiresAuth: false }
      }
    ]
  }
]
```

#### 路由守卫
```typescript
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - AutoCodeWeb`
  }
  
  // 权限检查
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      next('/auth/login')
      return
    }
  }
  
  next()
})
```

### 样式系统

#### CSS 变量
```scss
:root {
  // 颜色系统
  --primary-color: #2D3748;
  --accent-color: #3182CE;
  --success-color: #38A169;
  --warning-color: #D69E2E;
  --error-color: #E53E3E;
  
  // 间距系统
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-xxl: 48px;
}
```

#### 混入函数
```scss
@mixin flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

@mixin glassmorphism {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: var(--shadow-lg);
}
```

### HTTP 客户端

#### Axios 配置
```typescript
// utils/http.ts
class HttpService {
  private instance: AxiosInstance
  
  constructor() {
    this.instance = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json'
      }
    })
    
    this.setupInterceptors()
  }
  
  private setupInterceptors() {
    // 请求拦截器 - 添加认证头
    this.instance.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      }
    )
    
    // 响应拦截器 - 处理错误
    this.instance.interceptors.response.use(
      (response) => response.data,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('token')
          window.location.href = '/auth/login'
        }
        return Promise.reject(error)
      }
    )
  }
}
```

## 环境配置

### 环境变量
```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_ENV=development

# .env.production
VITE_API_BASE_URL=/api
VITE_APP_ENV=production
```

### Vite 配置
```typescript
// vite.config.ts
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
```

## 部署

### Docker 部署
```dockerfile
# Dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 静态部署
```bash
# 构建生产版本
pnpm build

# 将 dist 目录部署到 Web 服务器
```

## 开发规范

### 代码规范
- 使用 ESLint + Prettier 进行代码格式化
- 遵循 Vue 3 Composition API 最佳实践
- 使用 TypeScript 严格模式
- 组件和函数添加 JSDoc 注释

### Git 提交规范
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建过程或辅助工具的变动
```

### 性能优化
- 使用 Vue 3 的 `<script setup>` 语法
- 合理使用 `computed` 和 `watch`
- 组件懒加载
- 图片懒加载
- 代码分割

## 常见问题

### Q: 开发服务器启动失败
A: 检查 Node.js 版本和依赖安装，尝试删除 node_modules 重新安装

### Q: TypeScript 类型错误
A: 运行 `pnpm type-check` 查看详细错误信息，确保类型定义正确

### Q: 样式不生效
A: 检查 SCSS 文件是否正确导入，确保 CSS 变量定义正确

### Q: API 请求失败
A: 检查后端服务是否启动，确认 API 地址配置正确

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系方式

- 项目维护者: James (DEV Agent)
- 邮箱: dev@autocodeweb.com
- 项目地址: https://github.com/autocodeweb/frontend

---

*本文档为 AutoCodeWeb 前端项目的开发指南，由 DEV Agent James 创建*
