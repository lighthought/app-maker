<template>
  <div class="sidebar">
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
import { useI18n } from 'vue-i18n'
import { NMenu, NIcon } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
// 导入图标
import { HomeIcon, DashboardIcon, AddIcon } from '@/components/icon'

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false
})

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const activeKey = computed(() => route.name as string)

const menuOptions = computed((): MenuOption[] => [
  {
    label: t('nav.dashboard'),
    key: 'Dashboard',
    icon: renderIcon(DashboardIcon)
  },
  {
    label: t('nav.createProject'),
    key: 'CreateProject',
    icon: renderIcon(AddIcon)
  }
])

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: icon })
}

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