import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import i18n from '@/locales'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/pages/Home.vue'),
    meta: { titleKey: 'nav.home', requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/pages/Dashboard.vue'),
    meta: { titleKey: 'nav.dashboard', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/create-project',
    name: 'CreateProject',
    component: () => import('@/pages/CreateProject.vue'),
    meta: { titleKey: 'nav.createProject', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/project/:guid',
    name: 'ProjectEdit',
    component: () => import('@/pages/ProjectEdit.vue'),
    meta: { titleKey: 'project.editProject', requiresAuth: true, layout: 'default' }
  },
  {
    path: '/auth',
    name: 'Auth',
    component: () => import('@/pages/Auth.vue'),
    meta: { titleKey: 'auth.login', requiresAuth: false }
  },
  {
    path: '/debug/websocket',
    name: 'WebSocketDebug',
    component: () => import('@/pages/WebSocketDebug.vue'),
    meta: { titleKey: 'common.websocketDebug', requiresAuth: true, layout: 'default' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  // 设置页面标题
  const titleKey = to.meta.titleKey as string
  if (titleKey) {
    document.title = i18n.global.t(titleKey)
  } else if (to.meta.title) {
    document.title = to.meta.title as string
  } else {
    document.title = 'AppMaker'
  }
  
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
  if (to.meta.requiresAuth) {
    if (!userStore.isAuthenticated || !userStore.token) {
      console.log('需要认证但未认证，跳转到登录页')
      next('/auth')
      return
    }
    
    // 验证 token 有效性
    try {
      const isValid = await userStore.validateTokenOnStartup()
      if (isValid) {
        console.log('token 验证成功，允许访问')
        next()
      } else {
        console.log('token 验证失败，清除认证状态并跳转到登录页')
        userStore.clearAuth()
        next('/auth')
      }
    } catch (error: any) { 
      console.error('token 验证异常:', error)
      console.log('token 验证异常，清除认证状态并跳转到登录页')
      userStore.clearAuth()
      next('/auth')
    }
  } else {
    console.log('不需要认证，正常导航')
    next()
  }
})

export default router