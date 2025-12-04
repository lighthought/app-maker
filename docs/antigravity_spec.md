# Antigravity Spec: App Maker

> [!IMPORTANT]
> This document is the **Source of Truth** for AI Agents working on the App Maker project. Read this first to understand the system context, architecture, and development workflows.

## 1. Project Overview

**App Maker** is a multi-Agent collaboration platform that automates the creation of apps and websites. It uses a microservices architecture with a Go backend, Vue 3 frontend, and a dedicated Agents service.

### Core Philosophy
- **"Idea to App"**: Users describe what they want; Agents build it.
- **Agent Collaboration**: Specialized Agents (PM, UX, Architect, Dev) work together via defined workflows.
- **Local First**: Designed to run locally with Docker and local LLMs (Ollama) or cloud APIs.

## 2. Technology Stack & Standards

### Backend (`/backend`)
- **Language**: Go 1.24+
- **Framework**: Gin (Web), GORM (ORM), Viper (Config), Zap (Log)
- **Database**: PostgreSQL 15+
- **Cache/Queue**: Redis 7+, Asynq (Task Queue)
- **Architecture**: Clean Architecture (API -> Service -> Repository -> Data)
- **Conventions**:
    - Use `internal/` for private packages.
    - Use `shared-models` for common types/utils.
    - Error handling: Wrap errors with context.

### Frontend (`/frontend`)
- **Language**: TypeScript 5.2+
- **Framework**: Vue 3 (Composition API), Vite
- **UI Library**: Naive UI
- **State**: Pinia
- **Conventions**:
    - Components: PascalCase.
    - Stores: `use[Name]Store` pattern.
    - API: Use `HttpService` wrapper.

### Agents Service (`/agents`)
- **Language**: Go 1.24+
- **Framework**: Gin, Asynq
- **Role**: Orchestrates AI tasks, executes CLI tools (BMad, Git), manages project workspace.
- **Communication**: Polling/Webhooks with Backend.

### Shared Models (`/shared-models`)
- **Language**: Go
- **Purpose**: Shared types (User, Project, Task), Auth helpers, Logger, HTTP Client.
- **Rule**: If a type is used by both Backend and Agents, put it here.

## 3. Directory Structure Map

```
/
├── backend/                # Main API Service
│   ├── cmd/server/         # Entry point
│   ├── internal/           # Private code (api, services, repos)
│   └── docs/               # Swagger docs
├── frontend/               # Web Application
│   ├── src/
│   │   ├── components/     # UI Components
│   │   ├── pages/          # Route Views
│   │   └── stores/         # Pinia Stores
├── agents/                 # AI Agents Execution Service
│   ├── internal/           # Agent logic & handlers
│   └── design/             # (Legacy) Design docs
├── shared-models/          # Shared Go code
├── docs/                   # Project Documentation (You are here)
│   ├── backend/            # Backend Architecture & API
│   ├── frontend/           # Frontend Architecture & UX
│   ├── agents/             # Agents Architecture
│   └── antigravity_spec.md # This file
└── docker-compose.yml      # Orchestration
```

## 4. Key Workflows

### 4.1 Development Workflow (User's Perspective)
1.  **Project Creation**: User inputs idea -> Backend creates Project record.
2.  **PM Analysis**: Backend triggers PM Agent -> Agents Service runs LLM analysis -> Generates PRD.
3.  **Architecture**: Architect Agent generates DB Schema & API Design.
4.  **Coding**: Dev Agent implements stories -> Commits to Git.
5.  **Feedback**: User reviews via Chat/Preview -> Cycle repeats.

### 4.2 Technical Workflow (Agent's Perspective)
- **Task Execution**:
    1.  Backend pushes task to Redis (Asynq).
    2.  Agents Service worker picks up task.
    3.  Worker executes logic (call LLM, run git, write file).
    4.  Worker updates Task status & Project status via Backend API.
- **Real-time Updates**:
    - Backend pushes updates to Frontend via WebSocket.

## 5. Coding Guidelines for Agents

### General
- **Readability**: Write clear, commented code.
- **Safety**: Validate inputs, handle errors gracefully.
- **Context**: When modifying a file, read the surrounding code to maintain consistency.

### Go (Backend/Agents)
- **Project Layout**: Follow Standard Go Project Layout.
- **Dependency Injection**: Use the `container` package for service injection.
- **Configuration**: Do not hardcode values; use `config` package.

### Vue (Frontend)
- **Composition API**: Use `<script setup lang="ts">`.
- **Typing**: Explicitly type Props, Emits, and Store state.
- **Styling**: Use Scoped SCSS or Utility classes.

## 6. Verification & Testing
- **Backend**: `go test ./...`
- **Frontend**: `pnpm type-check`, `pnpm lint`
- **E2E**: Manual verification via Browser (currently).

## 7. Common Pitfalls
- **Import Cycles**: Avoid circular dependencies between `internal` packages. Use interfaces or move shared code to `shared-models`.
- **Database Migrations**: When changing models, ensure GORM migrations are handled or create SQL migration scripts if auto-migrate is disabled.
- **Env Vars**: Check `.env.example` for required environment variables.

---
*End of Spec. Use this context to navigate and contribute to App Maker.*
