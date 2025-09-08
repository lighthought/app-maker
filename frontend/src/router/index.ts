import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/pages/Home.vue'),
    meta: { title: 'AutoCode', requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/pages/Dashboard.vue'),
    meta: { title: '控制台', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/create-project',
    name: 'CreateProject',
    component: () => import('@/pages/CreateProject.vue'),
    meta: { title: '创建项目', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/project/:id',
    name: 'ProjectEdit',
    component: () => import('@/pages/ProjectEdit.vue'),
    meta: { title: '项目编辑', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/auth',
    name: 'Auth',
    component: () => import('@/pages/Auth.vue'),
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
  document.title = to.meta.title ? `${to.meta.title}` : 'AutoCode'
  
  console.log('路由守卫检查:', {
    to: to.path,
    from: from.path,
    isAuthenticated: userStore.isAuthenticated,
    requiresAuth: to.meta.requiresAuth,
    hasToken: !!userStore.token,
    hasUser: !!userStore.user
  })
  
  // 如果目标是 auth 页面，直接允许访问（避免循环重定向）
  if (to.path === '/auth') {
    console.log('访问登录页，直接允许')
    next()
    return
  }
  
  // 检查是否需要认证
  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    console.log('需要认证但未认证，跳转到登录页')
    next('/auth')
  } else {
    console.log('正常导航')
    next()
  }
})

export default router