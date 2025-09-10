# AutoCodeWeb 前端项目

## 项目简介

AutoCodeWeb 是一个基于 Vue.js 3 + TypeScript + Naive UI 的现代化前端项目，支持多 Agent 协作的自动代码生成平台。项目采用组件化开发，响应式设计，为用户提供直观、高效的项目创建和管理体验。

## 实际功能特性

### 已实现的核心功能
- ✅ **用户认证系统** - 完整的登录、注册、登出流程，支持token自动刷新
- ✅ **项目管理** - 项目创建、列表展示、详情查看、删除功能
- ✅ **实时对话** - 与AI Agent进行实时交互，支持Markdown渲染
- ✅ **开发进度跟踪** - 可视化项目开发阶段和进度
- ✅ **文件管理** - 查看项目文件结构、内容展示、项目下载
- ✅ **项目预览** - 实时预览项目效果（iframe嵌入）
- ✅ **响应式设计** - 适配桌面、平板、手机各种屏幕尺寸
- ✅ **国际化支持** - 中英文切换功能
- ✅ **分屏布局** - 项目编辑页面的左右分屏设计

## 技术栈

### 核心框架
- **Vue.js 3.4+** - 渐进式 JavaScript 框架，使用 Composition API
- **TypeScript 5.2+** - 类型安全的 JavaScript 超集
- **Vite 5.0+** - 下一代前端构建工具

### UI 组件库
- **Naive UI 2.37+** - Vue 3 组件库，支持 TypeScript
- **@iconify/vue 4.1+** - 图标库组件

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
- **@vueuse/core 10.7+** - Vue 组合式 API 工具集
- **marked 16.2+** - Markdown 解析器

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── assets/            # 静态资源（图片、字体等）
│   │   └── logo.svg       # 应用Logo
│   ├── components/        # 通用组件
│   │   ├── common/        # 基础组件
│   │   │   ├── index.ts   # 组件导出
│   │   │   └── SmartInput.vue # 智能输入组件
│   │   ├── layout/        # 布局组件
│   │   │   ├── Header.vue     # 顶部导航
│   │   │   ├── PageLayout.vue # 页面布局
│   │   │   └── Sidebar.vue    # 侧边栏
│   │   ├── ConversationMessage.vue # 对话消息组件
│   │   ├── ConversationContainer.vue # 对话容器组件
│   │   ├── DevStages.vue      # 开发阶段组件
│   │   ├── ProjectPanel.vue   # 项目面板组件
│   │   └── UserSettingsModal.vue # 用户设置弹窗
│   ├── pages/             # 页面组件
│   │   ├── Auth.vue           # 认证页面
│   │   ├── CreateProject.vue  # 创建项目页面
│   │   ├── Dashboard.vue      # 仪表板页面
│   │   ├── Home.vue           # 首页
│   │   └── ProjectEdit.vue    # 项目编辑页面
│   ├── router/            # 路由配置
│   │   └── index.ts           # 路由定义
│   ├── stores/            # 状态管理
│   │   ├── file.ts            # 文件状态
│   │   ├── project.ts         # 项目状态
│   │   └── user.ts            # 用户状态
│   ├── styles/            # 样式文件
│   │   ├── main.scss          # 主样式文件
│   │   ├── mixins.scss        # SCSS混入
│   │   └── variables.scss     # CSS变量
│   ├── types/             # TypeScript类型定义
│   │   ├── project.ts         # 项目相关类型
│   │   └── user.ts            # 用户相关类型
│   ├── utils/             # 工具函数
│   │   ├── config.ts          # 配置管理
│   │   ├── http.ts            # HTTP服务
│   │   ├── log.ts             # 日志工具
│   │   └── time.ts            # 时间工具
│   ├── App.vue            # 根组件
│   ├── main.ts            # 应用入口
│   └── vite-env.d.ts     # Vite 环境类型声明
├── Dockerfile             # 开发环境Docker配置
├── Dockerfile.prod        # 生产环境Docker配置
├── nginx.conf             # 开发环境Nginx配置
├── nginx.prod.conf        # 生产环境Nginx配置
├── package.json           # 依赖配置
├── pnpm-lock.yaml         # 依赖锁定文件
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
    component: () => import('@/pages/Home.vue'),
    meta: { title: 'AutoCode', requiresAuth: false }
  },
  {
    path: '/dashboard',
    component: () => import('@/pages/Dashboard.vue'),
    meta: { title: '控制台', requiresAuth: true }
  }
]
```

#### 路由守卫
```typescript
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title}`
  }
  
  // 权限检查
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      next('/auth')
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
          window.location.href = '/auth'
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
  },
  build: {
    target: 'es2015',
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['naive-ui'],
          utils: ['axios', '@vueuse/core']
        }
      }
    }
  }
})
```

## 部署

### Docker 部署

#### 开发环境
```dockerfile
# Dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY . .
RUN pnpm exec vite build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]
```

#### 生产环境
```dockerfile
# Dockerfile.prod
FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY . .
RUN pnpm exec vite build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.prod.conf /etc/nginx/nginx.conf
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

## 功能特性

### 核心功能
- **用户认证系统** - 支持登录、注册、密码重置
- **项目管理** - 创建、编辑、删除项目
- **实时对话** - 与AI Agent进行实时交互
- **开发进度跟踪** - 可视化项目开发阶段
- **文件管理** - 查看项目文件结构和内容
- **项目预览** - 实时预览项目效果

### 页面功能详情
- **首页 (Home)** - 产品介绍、快速创建项目、用户项目展示、中英文切换
- **认证页 (Auth)** - 登录/注册切换、表单验证、社交登录按钮、协议弹窗
- **仪表板 (Dashboard)** - 项目统计卡片、搜索筛选、分页展示、系统状态监控
- **创建项目 (CreateProject)** - 智能输入框、项目需求输入、自动跳转
- **项目编辑 (ProjectEdit)** - 分屏布局、对话交互、文件查看、代码展示

### 技术实现亮点
- **TypeScript 严格模式** - 完整的类型定义和类型安全
- **Pinia 状态管理** - 现代化的响应式状态管理
- **Vue 3 Composition API** - 更好的逻辑复用和类型推导
- **Naive UI 组件库** - 丰富的UI组件和主题系统
- **SCSS 模块化样式** - CSS变量、混入、响应式设计
- **Axios 拦截器** - 统一的请求/响应处理和错误处理
- **路由守卫** - 认证检查和权限控制
- **Docker 容器化** - 开发和生产环境容器部署

## 常见问题

### Q: 开发服务器启动失败
A: 检查 Node.js 版本和依赖安装，尝试删除 node_modules 重新安装

### Q: TypeScript 类型错误
A: 运行 `pnpm type-check` 查看详细错误信息，确保类型定义正确

### Q: 样式不生效
A: 检查 SCSS 文件是否正确导入，确保 CSS 变量定义正确

### Q: API 请求失败
A: 检查后端服务是否启动，确认 API 地址配置正确

### Q: 路由跳转问题
A: 检查路由守卫逻辑，确认用户认证状态

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
- 邮箱: qqjack2012@gmail.com
- 项目地址: https://github.com/lighthought/app-maker

---

*本文档为 AutoCodeWeb 前端项目的开发指南，由 DEV Agent James 创建*