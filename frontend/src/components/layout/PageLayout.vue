<template>
  <n-layout position="absolute" style="height: 100vh;">
    <!-- 顶部导航 -->
    <n-layout-header bordered position="absolute">
      <Header 
        :is-mobile="isMobile"
        :collapsed="collapsed"
        @toggle-sidebar="collapsed = !collapsed" 
        @open-settings="showSettingsModal = true" 
      />
    </n-layout-header>
    
    <!-- 下方内容区域 -->
    <n-layout has-sider position="absolute" style="top: 64px;">
      <!-- 侧边栏 (桌面端) -->
      <n-layout-sider
        v-if="!isMobile"
        bordered
        collapse-mode="width"
        :collapsed-width="64"
        :width="180"
        :collapsed="collapsed"
        show-trigger
        @collapse="collapsed = true"
        @expand="collapsed = false"
        class="desktop-sidebar"
      >
        <Sidebar :collapsed="collapsed" />
      </n-layout-sider>
      
      <!-- 页面内容 -->
      <n-layout-content>
        <div class="content-wrapper">
          <slot />
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>

  <!-- 用户设置弹窗 -->
  <UserSettingsModal v-model:show="showSettingsModal" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { NLayout, NLayoutSider, NLayoutHeader, NLayoutContent } from 'naive-ui'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'
import UserSettingsModal from '@/components/UserSettingsModal.vue'

// 布局相关状态
const collapsed = ref(false)
const showSettingsModal = ref(false)
const isMobile = ref(false)

// 检测移动端
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.content-wrapper {
  padding: var(--spacing-lg);
  min-height: calc(100vh - 64px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .content-wrapper {
    padding: var(--spacing-md);
  }
}

@media (max-width: 480px) {
  .content-wrapper {
    padding: var(--spacing-sm);
  }
}
</style>
