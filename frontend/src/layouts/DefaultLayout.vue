<template>
  <n-layout has-sider v-if="!isHomePage">
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
          <router-view />
        </div>
      </n-layout-content>
    </n-layout>
  </n-layout>
  
  <!-- 主页特殊布局 -->
  <div v-else>
    <router-view />
  </div>

  <!-- 用户设置弹窗 - 始终显示，不受布局条件影响 -->
  <UserSettingsModal v-model:show="showSettingsModal" />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { NLayout, NLayoutSider, NLayoutHeader, NLayoutContent } from 'naive-ui'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'
import UserSettingsModal from '@/components/UserSettingsModal.vue'

const route = useRoute()
const collapsed = ref(false)
const showSettingsModal = ref(false)

// 监听 showSettingsModal 变化
watch(showSettingsModal, (newVal) => {
  console.log('showSettingsModal changed:', newVal)
})

// 判断是否为主页
const isHomePage = computed(() => {
  return route.name === 'Home' || route.path === '/'
})
</script>

<style scoped>
.content-wrapper {
  padding: var(--spacing-lg);
  min-height: calc(100vh - 64px);
}
</style>