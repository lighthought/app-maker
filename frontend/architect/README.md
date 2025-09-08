# Frontend Architecture Documentation

## 概述

本目录包含 AutoCodeWeb 前端架构的详细文档，帮助开发者快速了解前端技术栈、开发规范和约束条件。

## 文档结构

- **ARCHITECTURE.md** - 完整的前端架构文档
  - 系统架构概览
  - 详细的 UML 类图
  - 数据流图
  - 技术栈说明
  - 目录结构说明
  - 开发约束和规范
  - 性能优化策略
  - 安全考虑
  - 扩展性设计

## 快速开始

1. **技术栈**: Vue 3 + TypeScript + Naive UI + Pinia + Axios
2. **状态管理**: 通过 Pinia stores 管理应用状态
3. **数据获取**: stores 通过 httpService 封装的接口与后端交互
4. **组件开发**: 使用 Composition API 和 TypeScript
5. **样式管理**: SCSS + CSS 变量 + 响应式设计

## 关键约束

- **组件命名**: PascalCase
- **状态管理**: 每个功能模块对应一个 Store
- **HTTP 请求**: 统一通过 HttpService 发送
- **路由管理**: 使用 Vue Router + 路由守卫
- **样式隔离**: 使用 scoped 样式

## 目录清理

已清理以下空文件夹：
- `components/business/`
- `composables/`
- `config/`
- `directives/`
- `hooks/`
- `views/`

这些空文件夹在架构文档中未包含，保持代码库的整洁性。
