<template>
  <div class="header">
    <div class="header-left">
      <n-button
        quaternary
        @click="goToHome"
        class="mobile-logo-button"
      >
        <div class="desktop-logo">
          <img src="@/assets/logo.svg" alt="AppMaker" class="header-logo" />
          <span class="header-title">AppMaker</span>
        </div>
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { NButton, NIcon, NAvatar, NPopover, NDivider } from 'naive-ui'
// 导入图标
import {
  MenuIcon,
  BellIcon,
  SettingsIcon,
  UserIcon,
  LogoutIcon,
  CloseIcon
} from '@/components/icon'

interface Props {
  isMobile?: boolean
  collapsed?: boolean
}

interface Emits {
  'toggle-sidebar': []
  'open-settings': []
}

const props = withDefaults(defineProps<Props>(), {
  isMobile: false,
  collapsed: false
})
const emit = defineEmits<Emits>()

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()

// 状态管理
const showNotifications = ref(false)
const showUserMenu = ref(false)

// 事件处理
const goToHome = () => {
  router.push('/')
}

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

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* 桌面端产品 Logo */
.desktop-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: var(--spacing-sm);
  cursor: default;
}

/* 移动端产品图标按钮 */
.mobile-logo-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  height: auto;
}

.header-logo {
  width: 28px;
  height: 28px;
}

.header-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--primary-color);
  white-space: nowrap;
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

/* 响应式设计 */
@media (max-width: 768px) {
  .header {
    padding: 0 var(--spacing-md);
    height: 56px;
  }
  
  .header-logo {
    width: 24px;
    height: 24px;
  }
  
  .header-title {
    font-size: 1rem;
  }
  
  .notification-panel,
  .user-menu-panel {
    width: 280px;
  }
}

@media (max-width: 480px) {
  .header {
    padding: 0 var(--spacing-sm);
    height: 52px;
  }
  
  .mobile-logo-button {
    padding: 6px 8px;
    gap: 6px;
  }
  
  .header-logo {
    width: 20px;
    height: 20px;
  }
  
  .header-title {
    font-size: 0.9rem;
  }
  
  .header-right {
    gap: 4px;
  }
  
  .notification-panel,
  .user-menu-panel {
    width: calc(100vw - 32px);
    max-width: 320px;
  }
  
  .notification-header,
  .user-info {
    padding: var(--spacing-sm) var(--spacing-md);
  }
  
  .notification-content {
    padding: var(--spacing-md);
  }
  
  .menu-items {
    padding: var(--spacing-xs);
  }
}
</style>