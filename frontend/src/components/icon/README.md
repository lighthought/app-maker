# 图标库使用说明

## 简介

这是一个统一管理的 SVG 图标库，所有自定义图标都集中在 `index.ts` 文件中。

## 使用方式

### 1. 在组件中导入图标

```typescript
import { AddIcon, CheckIcon, CloseIcon } from '@/components/icon'
```

### 2. 在模板中使用

#### 方式一：直接使用 (推荐)
```vue
<template>
  <n-icon>
    <AddIcon />
  </n-icon>
</template>
```

#### 方式二：使用 component :is
```vue
<template>
  <component :is="AddIcon" />
</template>
```

#### 方式三：在 script 中渲染
```typescript
import { h } from 'vue'
import { NIcon } from 'naive-ui'
import { AddIcon } from '@/components/icon'

const renderIcon = (icon: any) => {
  return () => h(NIcon, null, { default: icon })
}

// 然后在模板中使用
<component :is="renderIcon(AddIcon)" />
```

## 添加新图标

### 1. 编辑 `index.ts` 文件

在 `index.ts` 文件中添加新图标：

```typescript
export const YourNewIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',  // 或 'none' 用于 stroke 图标
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'YOUR_SVG_PATH_DATA' })
])
```

### 2. Stroke 图标示例 (线条图标)

```typescript
export const YourStrokeIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'none',
  stroke: 'currentColor',
  'stroke-width': '2',
  'stroke-linecap': 'round',
  'stroke-linejoin': 'round',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'YOUR_SVG_PATH_DATA' })
])
```

### 3. 多路径图标示例

```typescript
export const ComplexIcon = () => h('svg', { 
  viewBox: '0 0 24 24', 
  fill: 'currentColor',
  style: 'width: 1em; height: 1em;'
}, [
  h('path', { d: 'FIRST_PATH' }),
  h('path', { d: 'SECOND_PATH' }),
  h('circle', { cx: '12', cy: '12', r: '3' })
])
```

## 图标分类

当前图标库按功能分类：

- **通用操作**: AddIcon, CheckIcon, CloseIcon, CopyIcon, RefreshIcon, SearchIcon, EditIcon, SettingsIcon
- **文件和文件夹**: FolderIcon, FileIcon, EmptyIcon
- **代码和开发**: CodeIcon, PreviewIcon
- **设备**: DesktopIcon, TabletIcon, PhoneIcon
- **状态**: ClockIcon, ErrorIcon, WarningIcon, LoadingIcon, TrendingUpIcon, CheckCircleIcon, CloseCircleIcon
- **可见性**: EyeIcon, EyeOffIcon
- **分享和网络**: ShareIcon, GlobeIcon, ExternalLinkIcon
- **用户和角色**: UserIcon, AssistantIcon
- **导航**: HomeIcon, DashboardIcon, ChevronUpIcon, ChevronDownIcon, ChevronRightIcon, ArrowLeftIcon
- **输入和操作**: SendIcon, DownloadIcon, ReloadIcon
- **Agent 角色**: DevIcon, PmIcon, ArchIcon, UxIcon, QaIcon, OpsIcon, InfoIcon
- **Header**: MenuIcon, BellIcon, LogoutIcon

## 注意事项

1. 所有图标都使用 `1em` 作为宽高，这样可以继承父元素的字体大小
2. 图标颜色通过 `currentColor` 继承，可以通过 CSS 的 `color` 属性控制
3. 添加新图标时，请遵循现有的命名规范：`[Name]Icon`
4. 为新图标添加适当的分类注释，保持代码整洁
5. SVG viewBox 通常是 `0 0 24 24`，保持一致性

## 获取 SVG 图标

推荐的图标资源：

- [Material Design Icons](https://fonts.google.com/icons)
- [Heroicons](https://heroicons.com/)
- [Feather Icons](https://feathericons.com/)
- [Lucide Icons](https://lucide.dev/)
- [Ionicons](https://ionic.io/ionicons)

从这些资源获取 SVG 后，复制 `path` 标签的 `d` 属性值即可。

