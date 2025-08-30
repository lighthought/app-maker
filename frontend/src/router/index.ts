import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

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
    name: 'Auth',
    component: () => import('@/pages/auth/Auth.vue'),
    meta: { title: '登录/注册', requiresAuth: false }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 煲应用` : '煲应用'
  
  // 检查是否需要认证
  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    next('/auth')
  } else if (to.path === '/auth' && userStore.isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router