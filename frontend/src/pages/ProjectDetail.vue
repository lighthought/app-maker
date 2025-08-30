<template>
  <div class="project-detail-page">
    <h1>项目详情</h1>
    <p>项目ID: {{ $route.params.id }}</p>
  </div>
</template>

<script setup lang="ts">
// ProjectDetail page component
</script>

<style scoped>
.project-detail-page {
  padding: var(--spacing-lg);
}
</style>