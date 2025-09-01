<template>
  <n-layout has-sider>
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="180"
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
        <Header @toggle-sidebar="collapsed = !collapsed" @open-settings="showSettingsModal = true" />
      </n-layout-header>
      
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
import { ref } from 'vue'
import { NLayout, NLayoutSider, NLayoutHeader, NLayoutContent } from 'naive-ui'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'
import UserSettingsModal from '@/components/UserSettingsModal.vue'

// 布局相关状态
const collapsed = ref(false)
const showSettingsModal = ref(false)
</script>

<style scoped>
.content-wrapper {
  padding: var(--spacing-lg);
  min-height: calc(100vh - 64px);
}
</style>
