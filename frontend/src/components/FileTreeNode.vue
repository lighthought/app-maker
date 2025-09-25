<template>
  <div>
    <!-- 当前节点 -->
    <div
      :class="[
        'tree-item',
        { 'tree-item--active': isSelected },
        { 'tree-item--folder': node.type === 'folder' }
      ]"
      :style="indentStyle"
      @click="handleClick"
    >
      <!-- 展开/收起图标 -->
      <span
        v-if="node.type === 'folder'"
        class="expand-icon"
        :style="{
          opacity: node.loaded && (!node.children || node.children.length === 0) ? 0 : 1,
          transition: 'opacity 0.2s ease'
        }"
      >
        <n-icon size="16">
          <ChevronDownIcon v-if="node.expanded" />
          <ChevronRightIcon v-else />
        </n-icon>
      </span>
      
      <!-- 文件/文件夹图标 -->
      <n-icon 
        size="20" 
        :color="getFileIconColor(node.type)"
        class="file-icon"
      >
        <FileIcon v-if="node.type === 'file'" />
        <FolderIcon v-else />
      </n-icon>
      
      <!-- 文件名 -->
      <span class="file-name">{{ node.name }}</span>
    </div>

    <!-- 子节点 -->
    <div v-if="node.type === 'folder' && node.expanded && node.children && node.children.length > 0" class="tree-children">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :selected-file="selectedFile"
        :project-guid="projectGuid"
        :level="level + 1"
        @select-file="$emit('selectFile', $event)"
        @expand-folder="$emit('expandFolder', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { NIcon } from 'naive-ui'
import type { FileTreeNode as FileTreeNodeType } from '@/stores/file'

interface Props {
  node: FileTreeNodeType
  selectedFile?: FileTreeNodeType | null
  projectGuid?: string
  level?: number
}

interface Emits {
  (e: 'selectFile', file: FileTreeNodeType): void
  (e: 'expandFolder', folder: FileTreeNodeType): void
}

const props = withDefaults(defineProps<Props>(), {
  selectedFile: null,
  projectGuid: '',
  level: 0
})

const emit = defineEmits<Emits>()

// 计算属性
const isSelected = computed(() => props.selectedFile?.path === props.node.path)

const indentStyle = computed(() => ({
  paddingLeft: `${props.level * 16 + 8}px`
}))

// 事件处理
const handleClick = () => {
  if (props.node.type === 'file') {
    emit('selectFile', props.node)
  } else {
    emit('expandFolder', props.node)
  }
}

// 获取文件图标颜色
const getFileIconColor = (type?: string) => {
  const colorMap = {
    file: '#666',
    folder: '#3182CE'
  }
  return colorMap[type as keyof typeof colorMap] || '#666'
}

// 图标组件
const FileIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z' })
])

const FolderIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M10,4H4C2.89,4 2,4.89 2,6V18A2,2 0 0,0 4,20H20A2,2 0 0,0 22,18V8C22,6.89 21.1,6 20,6H12L10,4Z' })
])

const ChevronRightIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M8.59 16.59L13.17 12 8.59 7.41 10 6l6 6-6 6-1.41-1.41z' })
])

const ChevronDownIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z' })
])
</script>

<style scoped>
.tree-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs); /* 减小间距 */
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  transition: background-color 0.2s ease;
  position: relative;
  font-size: 14px;
  min-height: 32px;
}

.tree-item:hover {
  background: var(--background-color);
}

.tree-item--active {
  background: #e6f3ff; /* 浅蓝色背景 */
  color: #1890ff; /* 蓝色文字 */
  border: 1px solid #91d5ff; /* 浅蓝色边框 */
}

.tree-item--active:hover {
  background: #bae7ff; /* 稍深的浅蓝色悬停背景 */
}

.tree-item--folder {
  font-weight: 500;
}

.expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  margin-right: 2px; /* 减小右边距 */
  flex-shrink: 0;
}

.file-icon {
  margin-right: 2px; /* 文件图标右边距 */
}

.file-name {
  font-size: 14px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
  display: flex;
  align-items: center;
}

.tree-children {
  position: relative;
}

/* 修复图标垂直对齐 */
.tree-item .n-icon {
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  font-size: 1rem;
}
</style>
