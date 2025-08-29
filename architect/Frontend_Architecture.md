# AutoCodeWeb 前端架构设计

## 1. 前端架构概述

### 1.1 架构理念
- **组件化开发**：基于Vue 3 Composition API的现代组件架构
- **状态管理**：Pinia驱动的响应式状态管理
- **路由管理**：Vue Router 4的单页应用路由
- **UI一致性**：Naive UI组件库的统一设计语言
- **TypeScript优先**：完整的类型安全和开发体验

### 1.2 技术栈选型
- **核心框架**：Vue.js 3.4+
- **构建工具**：Vite 5.0+
- **UI组件库**：Naive UI 2.38+
- **状态管理**：Pinia 2.1+
- **路由管理**：Vue Router 4.2+
- **HTTP客户端**：Axios 1.6+
- **开发语言**：TypeScript 5.2+

## 2. 项目结构设计

### 2.1 目录结构
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

### 2.2 组件分层架构
```
┌─────────────────────────────────────────────────────────────┐
│                        页面层 (Pages)                       │
│                    业务逻辑和页面组合                        │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      业务组件层 (Business)                   │
│                  特定业务功能的组件                          │
│              (Agent协作、项目管理、用户管理等)                │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      通用组件层 (Common)                     │
│                  可复用的通用组件                            │
│              (按钮、表单、表格、弹窗等)                      │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      基础组件层 (Base)                       │
│                  最基础的原子组件                            │
│              (图标、文本、容器等)                            │
└─────────────────────────────────────────────────────────────┘
```

## 3. 核心架构设计

### 3.1 状态管理架构 (Pinia)
```typescript
// stores/user.ts - 用户状态管理
export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string>('')
  const permissions = ref<string[]>([])
  
  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const hasPermission = computed(() => (permission: string) => 
    permissions.value.includes(permission)
  )
  
  // 动作
  const login = async (credentials: LoginCredentials) => { /* ... */ }
  const logout = () => { /* ... */ }
  const updateProfile = async (profile: UserProfile) => { /* ... */ }
  
  return {
    user, token, permissions, isLoggedIn, hasPermission,
    login, logout, updateProfile
  }
})

// stores/project.ts - 项目状态管理
export const useProjectStore = defineStore('project', () => {
  const projects = ref<Project[]>([])
  const currentProject = ref<Project | null>(null)
  const projectStatus = ref<ProjectStatus>('idle')
  
  const createProject = async (projectData: CreateProjectData) => { /* ... */ }
  const updateProject = async (projectId: string, updates: Partial<Project>) => { /* ... */ }
  const deleteProject = async (projectId: string) => { /* ... */ }
  
  return {
    projects, currentProject, projectStatus,
    createProject, updateProject, deleteProject
  }
})
```

### 3.2 路由架构 (Vue Router 4)
```typescript
// router/index.ts - 路由配置
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
      },
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/pages/Dashboard.vue'),
        meta: { title: '控制台', requiresAuth: true }
      },
      {
        path: '/create-project',
        name: 'CreateProject',
        component: () => import('@/pages/CreateProject.vue'),
        meta: { title: '创建项目', requiresAuth: true }
      },
      {
        path: '/project/:id',
        name: 'ProjectDetail',
        component: () => import('@/pages/ProjectDetail.vue'),
        meta: { title: '项目详情', requiresAuth: true }
      },
      {
        path: '/preview/:id',
        name: 'ProjectPreview',
        component: () => import('@/pages/ProjectPreview.vue'),
        meta: { title: '项目预览', requiresAuth: true }
      }
    ]
  },
  {
    path: '/auth',
    component: () => import('@/layouts/AuthLayout.vue'),
    children: [
      {
        path: 'login',
        name: 'Login',
        component: () => import('@/pages/auth/Login.vue')
      },
      {
        path: 'register',
        name: 'Register',
        component: () => import('@/pages/auth/Register.vue')
      }
    ]
  }
]
```

### 3.3 HTTP客户端架构 (Axios)
```typescript
// utils/http.ts - HTTP客户端配置
import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

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
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )
    
    // 响应拦截器
    this.instance.interceptors.response.use(
      (response: AxiosResponse) => response.data,
      (error) => {
        if (error.response?.status === 401) {
          // 处理未授权
          localStorage.removeItem('token')
          window.location.href = '/auth/login'
        }
        return Promise.reject(error)
      }
    )
  }
  
  public get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.get(url, config)
  }
  
  public post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.post(url, data, config)
  }
  
  public put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.put(url, data, config)
  }
  
  public delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.instance.delete(url, config)
  }
}

export const httpService = new HttpService()
```

## 4. 组件架构设计

### 4.1 布局组件架构
```typescript
// layouts/DefaultLayout.vue - 默认布局
<template>
  <n-layout has-sider>
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <Sidebar :collapsed="collapsed" />
    </n-layout-sider>
    
    <!-- 主内容区 -->
    <n-layout>
      <!-- 顶部导航 -->
      <n-layout-header bordered>
        <Header @toggle-sidebar="collapsed = !collapsed" />
      </n-layout-header>
      
      <!-- 页面内容 -->
      <n-layout-content>
        <div class="content-wrapper">
          <router-view />
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NLayout, NLayoutSider, NLayoutHeader, NLayoutContent } from 'naive-ui'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'

const collapsed = ref(false)
</script>
```

### 4.2 业务组件架构
```typescript
// components/business/AgentChat.vue - Agent对话组件
<template>
  <div class="agent-chat">
    <!-- 对话历史 -->
    <div class="chat-history" ref="chatHistoryRef">
      <div
        v-for="message in messages"
        :key="message.id"
        :class="['message', message.type]"
      >
        <div class="message-avatar">
          <n-avatar
            :src="message.avatar"
            :size="40"
            round
          />
        </div>
        <div class="message-content">
          <div class="message-header">
            <span class="agent-name">{{ message.agentName }}</span>
            <span class="message-time">{{ formatTime(message.timestamp) }}</span>
          </div>
          <div class="message-text" v-html="message.content" />
        </div>
      </div>
    </div>
    
    <!-- 用户输入区 -->
    <div class="chat-input">
      <n-input-group>
        <n-input
          v-model:value="inputMessage"
          type="textarea"
          :rows="3"
          placeholder="描述你的项目需求..."
          @keydown.enter.prevent="sendMessage"
        />
        <n-button
          type="primary"
          :disabled="!inputMessage.trim()"
          @click="sendMessage"
        >
          发送
        </n-button>
      </n-input-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { NAvatar, NInputGroup, NInput, NButton } from 'naive-ui'
import type { ChatMessage } from '@/types/chat'

const props = defineProps<{
  projectId: string
}>()

const emit = defineEmits<{
  messageSent: [message: ChatMessage]
}>()

const inputMessage = ref('')
const messages = ref<ChatMessage[]>([])
const chatHistoryRef = ref<HTMLElement>()

const sendMessage = async () => {
  if (!inputMessage.value.trim()) return
  
  const message: ChatMessage = {
    id: Date.now().toString(),
    type: 'user',
    content: inputMessage.value,
    timestamp: new Date(),
    agentName: '用户',
    avatar: '/avatars/user.png'
  }
  
  messages.value.push(message)
  emit('messageSent', message)
  
  inputMessage.value = ''
  
  await nextTick()
  scrollToBottom()
}

const scrollToBottom = () => {
  if (chatHistoryRef.value) {
    chatHistoryRef.value.scrollTop = chatHistoryRef.value.scrollHeight
  }
}
</script>
```

## 5. 样式架构设计

### 5.1 CSS架构
```scss
// styles/variables.scss - 设计变量
:root {
  // 颜色系统
  --primary-color: #3182CE;
  --primary-hover: #2C5AA0;
  --success-color: #38A169;
  --warning-color: #D69E2E;
  --error-color: #E53E3E;
  
  // 中性色
  --text-primary: #2D3748;
  --text-secondary: #4A5568;
  --text-disabled: #A0AEC0;
  --border-color: #E2E8F0;
  --background-color: #F7FAFC;
  
  // 间距系统
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-xxl: 48px;
  
  // 圆角系统
  --border-radius-sm: 4px;
  --border-radius-md: 8px;
  --border-radius-lg: 12px;
  --border-radius-xl: 16px;
  
  // 阴影系统
  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 6px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.1);
  --shadow-xl: 0 20px 25px rgba(0, 0, 0, 0.1);
}

// styles/mixins.scss - 混入函数
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

@mixin responsive($breakpoint) {
  @if $breakpoint == mobile {
    @media (max-width: 767px) { @content; }
  } @else if $breakpoint == tablet {
    @media (min-width: 768px) and (max-width: 1023px) { @content; }
  } @else if $breakpoint == desktop {
    @media (min-width: 1024px) { @content; }
  }
}
```

### 5.2 组件样式规范
```scss
// 组件样式示例
.agent-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--background-color);
  
  .chat-history {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-lg);
    
    .message {
      display: flex;
      margin-bottom: var(--spacing-lg);
      
      &.user {
        flex-direction: row-reverse;
        
        .message-content {
          background: var(--primary-color);
          color: white;
          border-radius: var(--border-radius-lg) var(--border-radius-sm);
        }
      }
      
      &.agent {
        .message-content {
          background: white;
          border: 1px solid var(--border-color);
          border-radius: var(--border-radius-sm) var(--border-radius-lg);
        }
      }
      
      .message-avatar {
        margin: 0 var(--spacing-md);
      }
      
      .message-content {
        max-width: 70%;
        padding: var(--spacing-md) var(--spacing-lg);
        box-shadow: var(--shadow-sm);
      }
    }
  }
  
  .chat-input {
    padding: var(--spacing-lg);
    border-top: 1px solid var(--border-color);
    background: white;
  }
}
```

## 6. 性能优化架构

### 6.1 代码分割策略
```typescript
// 路由级别的代码分割
const routes = [
  {
    path: '/dashboard',
    component: () => import('@/pages/Dashboard.vue') // 懒加载
  }
]

// 组件级别的代码分割
const LazyComponent = defineAsyncComponent(() => import('@/components/HeavyComponent.vue'))
```

### 6.2 缓存策略
```typescript
// 组件缓存
<template>
  <router-view v-slot="{ Component, route }">
    <keep-alive :include="cachedComponents">
      <component :is="Component" :key="route.path" />
    </keep-alive>
  </router-view>
</template>

// 数据缓存
const useProjectCache = () => {
  const cache = new Map<string, { data: any; timestamp: number }>()
  const CACHE_DURATION = 5 * 60 * 1000 // 5分钟
  
  const get = (key: string) => {
    const item = cache.get(key)
    if (item && Date.now() - item.timestamp < CACHE_DURATION) {
      return item.data
    }
    return null
  }
  
  const set = (key: string, data: any) => {
    cache.set(key, { data, timestamp: Date.now() })
  }
  
  return { get, set }
}
```

## 7. 开发工具配置

### 7.1 Vite配置
```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  },
  build: {
    target: 'es2015',
    outDir: 'dist',
    assetsDir: 'assets',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['naive-ui'],
          utils: ['axios', 'lodash-es']
        }
      }
    }
  }
})
```

### 7.2 TypeScript配置
```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

---

*本文档为 AutoCodeWeb 项目的前端架构设计，由架构师 Winston 创建*
