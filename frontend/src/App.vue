<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-loading-bar-provider>
        <n-dialog-provider>
          <n-notification-provider>
            <div id="app">
              <router-view v-slot="{ Component, route }">
                <transition
                  name="page"
                  mode="out-in"
                  @before-enter="beforeEnter"
                  @enter="enter"
                  @leave="leave"
                >
                  <component :is="Component" :key="route.path" />
                </transition>
              </router-view>
            </div>
          </n-notification-provider>
        </n-dialog-provider>
      </n-loading-bar-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { darkTheme } from 'naive-ui'

const route = useRoute()

// 使用默认主题（浅色主题）
const theme = null

// 页面过渡动画
const beforeEnter = (el: Element) => {
  (el as HTMLElement).style.opacity = '0'
  ;(el as HTMLElement).style.transform = 'translateY(20px)'
}

const enter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  element.style.transition = 'all 0.3s ease'
  element.style.opacity = '1'
  element.style.transform = 'translateY(0)'
  setTimeout(done, 300)
}

const leave = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  element.style.transition = 'all 0.3s ease'
  element.style.opacity = '0'
  element.style.transform = 'translateY(-20px)'
  setTimeout(done, 300)
}
</script>

<style>
#app {
  font-family: 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  height: 100vh;
  margin: 0;
  padding: 0;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  background-color: var(--background-color, #F7FAFC);
}

/* 页面过渡动画 */
.page-enter-active,
.page-leave-active {
  transition: all 0.3s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* 平滑滚动 */
html {
  scroll-behavior: smooth;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>