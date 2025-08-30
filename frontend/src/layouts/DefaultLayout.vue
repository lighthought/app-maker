<template>
  <n-layout has-sider v-if="!isHomePage">
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
  
  <!-- 主页特殊布局 -->
  <div v-else>
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { NLayout, NLayoutSider, NLayoutHeader, NLayoutContent } from 'naive-ui'
import Sidebar from '@/components/layout/Sidebar.vue'
import Header from '@/components/layout/Header.vue'

const route = useRoute()
const collapsed = ref(false)

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