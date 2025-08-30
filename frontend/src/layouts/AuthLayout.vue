<template>
  <div class="auth-layout">
    <router-view />
  </div>
</template>

<script setup lang="ts">
// Auth layout component
</script>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--accent-color) 100%);
}
</style>