<template>
  <PageLayout>
    <div class="project-preview-page">
      <h1>项目预览</h1>
      <p>项目ID: {{ $route.params.id }}</p>
    </div>
  </PageLayout>
</template>

<script setup lang="ts">
import PageLayout from '@/components/layout/PageLayout.vue'
// ProjectPreview page component
</script>

<style scoped>
.project-preview-page {
  padding: var(--spacing-lg);
}
</style>