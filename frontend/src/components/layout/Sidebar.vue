<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <h2 v-if="!collapsed">AutoCodeWeb</h2>
      <h2 v-else>ACW</h2>
    </div>
    <div class="sidebar-content">
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="activeKey"
        @update:value="handleMenuUpdate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NMenu } from 'naive-ui'
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

const menuOptions: MenuOption[] = [
  {
    label: '首页',
    key: 'Home',
    icon: () => '🏠'
  },
  {
    label: '控制台',
    key: 'Dashboard',
    icon: () => '📊'
  },
  {
    label: '创建项目',
    key: 'CreateProject',
    icon: () => '➕'
  }
]

const handleMenuUpdate = (key: string) => {
  router.push({ name: key })
}
</script>

<style scoped>
.sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
}

.sidebar-content {
  flex: 1;
  padding: var(--spacing-md);
}
</style>