<template>
  <div class="sidebar">
    <div class="sidebar-header" @click="goToHome">
      <img src="@/assets/logo.svg" alt="App-Maker" class="sidebar-logo" />
      <h2 v-if="!collapsed">App-Maker</h2>
      <h2 v-else>AC</h2>
    </div>
    <div class="sidebar-content">
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="24"
        :options="menuOptions"
        :value="activeKey"
        @update:value="handleMenuUpdate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NMenu, NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
})

const router = useRouter()
const route = useRoute()

const activeKey = computed(() => route.name as string)

// SVG 图标组件
const HomeIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z' })
])

const DashboardIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z' })
])

const AddIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z' })
])

const menuOptions: MenuOption[] = [
  {
    label: '控制台',
    key: 'Dashboard',
    icon: renderIcon(DashboardIcon)
  },
  {
    label: '创建项目',
    key: 'CreateProject',
    icon: renderIcon(AddIcon)
  }
]

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: icon })
}

const handleMenuUpdate = (key: string) => {
  router.push({ name: key })
}

const goToHome = () => {
  router.push('/')
}
</script>

<style scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  background: white;
  cursor: pointer;
  transition: all 0.3s ease;
  gap: var(--spacing-sm);
}

.sidebar-logo {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
}

.sidebar-header:hover {
  background: var(--background-color);
  transform: scale(1.02);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: bold;
  color: var(--primary-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-content {
  flex: 1;
  padding: var(--spacing-sm);
}

/* 修复折叠时的图标间距 */
:deep(.n-menu--collapsed .n-menu-item) {
  padding: 0 !important;
  margin: 4px 0;
}

:deep(.n-menu--collapsed .n-menu-item-content) {
  justify-content: center !important;
  padding: 12px !important;
}

:deep(.n-menu--collapsed .n-menu-item-content__icon) {
  margin: 0 !important;
}
</style>