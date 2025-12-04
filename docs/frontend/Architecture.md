# App Maker Frontend Architecture

## 1. Overview

**App Maker Frontend** is a modern Single Page Application (SPA) built with **Vue 3 + TypeScript + Naive UI**. It utilizes **Pinia** for state management and **Axios** for HTTP communication, providing a responsive and type-safe user experience. The system supports multi-Agent collaboration project management, real-time dialogue, and file management.

### 1.1 Core Concepts
- **Component-Based**: Modern component architecture based on Vue 3 Composition API.
- **State Management**: Reactive state management driven by Pinia.
- **Routing**: SPA routing via Vue Router 4.
- **UI Consistency**: Unified design language using Naive UI.
- **Type Safety**: Full TypeScript support.

### 1.2 Technology Stack
- **Framework**: Vue.js 3.4+
- **Build Tool**: Vite 5.0+
- **UI Library**: Naive UI 2.38+
- **State Management**: Pinia 2.1+
- **Router**: Vue Router 4.2+
- **HTTP Client**: Axios 1.6+
- **Language**: TypeScript 5.2+
- **CSS Preprocessor**: SCSS

## 2. Project Structure

### 2.1 Directory Layout

```
frontend/src/
├── App.vue                 # Root Component
├── main.ts                 # Application Entry
├── vite-env.d.ts          # Vite Env Types
├── assets/                 # Static Assets
├── components/             # Components
│   ├── common/            # Generic Components
│   ├── layout/            # Layout Components (Header, Sidebar)
│   ├── business/          # Business Components (Chat, ProjectPanel)
│   └── ...
├── pages/                  # Page Components (Views)
│   ├── Home.vue           # Homepage
│   ├── Dashboard.vue      # Dashboard
│   ├── ProjectEdit.vue    # Project Editor
│   └── ...
├── router/                 # Router Configuration
├── stores/                 # Pinia Stores
│   ├── user.ts            # User State
│   ├── project.ts         # Project State
│   └── file.ts            # File State
├── styles/                 # Global Styles & Variables
├── types/                  # TypeScript Type Definitions
└── utils/                  # Utility Functions (HTTP, Time, etc.)
```

### 2.2 Component Layering

```
┌─────────────────────────────────────────────────────────────┐
│                        Pages Layer                          │
│               Route Views, Page Composition                 │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Business Components                    │
│           Specific Business Logic (Chat, ProjectPanel)      │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                      Common Components                      │
│           Reusable UI Elements (SmartInput, Modals)         │
└─────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                       Base Components                       │
│           Atomic Elements (Naive UI Wrappers, Icons)        │
└─────────────────────────────────────────────────────────────┘
```

## 3. Core Architecture

### 3.1 State Management (Pinia)

- **UserStore**: Manages authentication, user profile, and permissions.
- **ProjectStore**: Manages project lists, current project details, and chat messages.
- **FileStore**: Manages project file tree and file content.

### 3.2 Routing (Vue Router)

- **Layouts**: `DefaultLayout` (Sidebar + Header) and `AuthLayout`.
- **Guards**: Authentication checks via global `beforeEach` guard.
- **Lazy Loading**: Route components are lazy-loaded for performance.

### 3.3 HTTP Client (Axios)

- **Interceptors**:
    - Request: Adds `Authorization: Bearer <token>` header.
    - Response: Handles global errors (e.g., 401 Unauthorized redirects to login).
- **Service Wrapper**: `HttpService` class encapsulates common methods (`get`, `post`, etc.).

## 4. Key Features Implementation

### 4.1 Real-time Dialogue
- **WebSocket**: Connects to backend for real-time message updates.
- **Components**: `ConversationContainer`, `ConversationMessage` (Markdown rendering).

### 4.2 Project Management
- **Creation**: Wizard-like flow for creating new projects.
- **Editing**: Split-screen view with Chat, File Explorer, and Preview.
- **Preview**: IFrame integration for live project preview.

### 4.3 File Management
- **File Tree**: Recursive component for displaying directory structure.
- **Editor**: Monaco Editor integration for code viewing/editing.

## 5. Deployment

- **Development**: Vite Dev Server with proxy to backend.
- **Production**:
    - Static build via Vite (`dist/`).
    - Nginx serving static files and proxying API requests.
    - Docker containerization.

---
*Documentation consolidated and updated for App Maker Frontend.*
