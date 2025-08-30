<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
// App root component
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
</style>