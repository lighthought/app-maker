<template>
  <PageLayout>
    <div class="project-detail-page">
      <h1>项目详情</h1>
      <p>项目ID: {{ $route.params.id }}</p>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import PageLayout from '@/components/layout/PageLayout.vue'
// ProjectDetail page component
</script>

<style scoped>
.project-detail-page {
  padding: var(--spacing-lg);
}
</style>