<template>
  <div class="header">
    <div class="header-left">
      <n-button
        quaternary
        circle
        @click="$emit('toggle-sidebar')"
      >
        <template #icon>
          <n-icon><Menu /></n-icon>
        </template>
      </n-button>
    </div>
    <div class="header-right">
      <n-button quaternary circle>
        <template #icon>
          <n-icon><Bell /></n-icon>
        </template>
      </n-button>
      <n-button quaternary circle>
        <template #icon>
          <n-icon><Settings /></n-icon>
        </template>
      </n-button>
      <n-avatar round size="medium" src="/avatar.jpg" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { NButton, NIcon, NAvatar } from 'naive-ui'

// 临时使用简单的图标，后续可以集成 @vicons/ionicons5
const Menu = () => '☰'
const Bell = () => '🔔'
const Settings = () => '⚙️'

defineEmits<{
  'toggle-sidebar': []
}>()
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
</style>