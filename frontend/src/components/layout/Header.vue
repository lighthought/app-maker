<template>
  <div class="header">
    <div class="header-left">
      <n-button
        quaternary
        circle
        @click="$emit('toggle-sidebar')"
      >
        <template #icon>
          <n-icon><MenuIcon /></n-icon>
        </template>
      </n-button>
    </div>
    <div class="header-right">
      <!-- 消息通知 -->
      <n-popover
        trigger="click"
        placement="bottom-end"
        :show="showNotifications"
        @clickoutside="showNotifications = false"
      >
        <template #trigger>
          <n-button quaternary circle @click="showNotifications = !showNotifications">
            <template #icon>
              <n-icon><BellIcon /></n-icon>
            </template>
          </n-button>
        </template>
        <div class="notification-panel">
          <div class="notification-header">
            <h3>{{ t('header.notifications') }}</h3>
            <n-button text size="small" @click="showNotifications = false">
              <template #icon>
                <n-icon><CloseIcon /></n-icon>
              </template>
            </n-button>
          </div>
          <div class="notification-content">
            <div class="empty-notification">
              <n-icon size="48" color="#CBD5E0">
                <BellIcon />
              </n-icon>
              <p>{{ t('header.noMessages') }}</p>
            </div>
          </div>
        </div>
      </n-popover>

      <!-- 用户菜单 -->
      <n-popover
        trigger="click"
        placement="bottom-end"
        :show="showUserMenu"
        @clickoutside="showUserMenu = false"
      >
        <template #trigger>
          <n-button quaternary circle @click="showUserMenu = !showUserMenu">
            <n-avatar round size="medium">
              <template #default>
                <n-icon><UserIcon /></n-icon>
              </template>
            </n-avatar>
          </n-button>
        </template>
        <div class="user-menu-panel">
          <div class="user-info">
            <n-avatar round size="medium">
              <template #default>
                <n-icon><UserIcon /></n-icon>
              </template>
            </n-avatar>
            <div class="user-details">
              <div class="username">{{ userStore.user?.username || userStore.user?.name || t('common.user') }}</div>
              <div class="user-email">{{ userStore.user?.email || '' }}</div>
            </div>
          </div>
          <n-divider style="margin: 0; height: 1px;" />
          <div class="menu-items">
            <n-button
              quaternary
              block
              @click="handleSettings"
            >
              <template #icon>
                <n-icon><SettingsIcon /></n-icon>
              </template>
              {{ t('header.userSettings') }}
            </n-button>
            <n-button
              quaternary
              block
              @click="handleLogout"
            >
              <template #icon>
                <n-icon><LogoutIcon /></n-icon>
              </template>
              {{ t('header.logout') }}
            </n-button>
          </div>
        </div>
      </n-popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { NButton, NIcon, NAvatar, NPopover, NDivider } from 'naive-ui'

interface Props {}

interface Emits {
  'toggle-sidebar': []
  'open-settings': []
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

// 状态管理
const showNotifications = ref(false)
const showUserMenu = ref(false)

// SVG 图标组件
const MenuIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z' })
])

const BellIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 22c1.1 0 2-.9 2-2h-4c0 1.1.89 2 2 2zm6-6v-5c0-3.07-1.64-5.64-4.5-6.32V4c0-.83-.67-1.5-1.5-1.5s-1.5.67-1.5 1.5v.68C7.63 5.36 6 7.92 6 11v5l-2 2v1h16v-1l-2-2z' })
])

const SettingsIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.07-0.94l2.03-1.58c0.18-0.14,0.23-0.41,0.12-0.61 l-1.92-3.32c-0.12-0.22-0.37-0.29-0.59-0.22l-2.39,0.96c-0.5-0.38-1.03-0.7-1.62-0.94L14.4,2.81c-0.04-0.24-0.24-0.41-0.48-0.41 h-3.84c-0.24,0-0.43,0.17-0.47,0.41L9.25,5.35C8.66,5.59,8.12,5.92,7.63,6.29L5.24,5.33c-0.22-0.08-0.47,0-0.59,0.22L2.74,8.87 C2.62,9.08,2.66,9.34,2.86,9.48l2.03,1.58C4.84,11.36,4.8,11.69,4.8,12s0.02,0.64,0.07,0.94l-2.03,1.58 c-0.18,0.14-0.23,0.41-0.12,0.61l1.92,3.32c0.12,0.22,0.37,0.29,0.59,0.22l2.39-0.96c0.5,0.38,1.03,0.7,1.62,0.94l0.36,2.54 c0.05,0.24,0.24,0.41,0.48,0.41h3.84c0.24,0,0.44-0.17,0.47-0.41l0.36-2.54c0.59-0.24,1.13-0.56,1.62-0.94l2.39,0.96 c0.22,0.08,0.47,0,0.59-0.22l1.92-3.32c0.12-0.22,0.07-0.47-0.12-0.61L19.14,12.94z M12,15.6c-1.98,0-3.6-1.62-3.6-3.6 s1.62-3.6,3.6-3.6s3.6,1.62,3.6,3.6S13.98,15.6,12,15.6z' })
])

const UserIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z' })
])

const LogoutIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z' })
])

const CloseIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z' })
])

// 事件处理
const handleSettings = () => {
  showUserMenu.value = false
  console.log('open-settings')
  emit('open-settings')
}

const handleLogout = async () => {
  try {
    await userStore.logout()
    showUserMenu.value = false
    // 跳转到首页并强制刷新，避免残留状态
    window.location.assign('/')
  } catch (error) {
    console.error('登出失败:', error)
  }
}
</script>

<style scoped>
.header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-lg);
  background: white;
  border-bottom: 1px solid var(--border-color);
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* 通知面板样式 */
.notification-panel {
  width: 320px;
  max-height: 400px;
  background: white;
  overflow: hidden;
}

.notification-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  background: var(--background-color);
  margin: 0;
}

.notification-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: bold;
  color: var(--primary-color);
}

.notification-content {
  padding: var(--spacing-lg);
}

.empty-notification {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  color: var(--text-secondary);
}

.empty-notification p {
  margin: var(--spacing-sm) 0 0 0;
  font-size: 0.9rem;
}

/* 用户菜单面板样式 */
.user-menu-panel {
  width: 280px;
  background: white;
  overflow: hidden;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: var(--background-color);
  margin: 0;
}

.user-details {
  flex: 1;
}

.username {
  font-weight: bold;
  color: var(--primary-color);
  font-size: 1rem;
  margin-bottom: var(--spacing-xs);
}

.user-email {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.menu-items {
  padding: var(--spacing-sm);
  margin-top: 0;
}

.menu-items .n-button {
  justify-content: flex-start;
  padding: var(--spacing-md) var(--spacing-lg);
  margin-bottom: var(--spacing-xs);
  border-radius: var(--border-radius-md);
}

.menu-items .n-button:hover {
  background: var(--background-color);
}

</style>