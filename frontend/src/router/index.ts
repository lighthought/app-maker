import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

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
        component: () => import('@/pages/auth/Login.vue'),
        meta: { title: '登录' }
      },
      {
        path: 'register',
        name: 'Register',
        component: () => import('@/pages/auth/Register.vue'),
        meta: { title: '注册' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
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

export default router