# AI 项目名称生成功能

## 功能概述

本项目已集成本地 Ollama AI 服务，用于智能生成项目名称和描述。当用户创建新项目时，系统会自动调用 AI 服务分析用户需求，生成符合要求的中文描述和英文标题。

## 配置说明

### 1. Traefik 配置

已在 `traefik-external.yml` 中配置了 Ollama 服务的路由：

```yaml
http:
  routers:
    ollama-chat:
      rule: "Host(`chat.app-maker.localhost`)"
      service: ollama-service
      entryPoints:
        - web

  services:
    ollama-service:
      loadBalancer:
        servers:
          - url: "http://host.docker.internal:11434"
        healthCheck:
          path: "/api/tags"
          interval: "30s"
          timeout: "5s"
```

### 2. 后端配置

在 `configs/config.yaml` 中添加了 AI 配置：

```yaml
ai:
  ollama:
    base_url: "http://chat.app-maker.localhost"
    model: "qwen2.5:7b"
    timeout: 30
```

### 3. 环境要求

- 本地需要运行 Ollama 服务
- 需要下载 `qwen2.5:7b` 模型：`ollama pull qwen2.5:7b`
- 确保 Ollama 服务在 `localhost:11434` 端口运行

## 实现细节

### 1. AI 工具类 (`internal/utils/ai_utils.go`)

- `OllamaClient`: Ollama API 客户端
- `GenerateProjectSummary()`: 生成项目总结的核心方法
- `TestConnection()`: 测试 Ollama 连接

### 2. 项目名称生成器 (`internal/services/project_name_generator.go`)

- 优先使用 AI 生成项目名称和描述
- 如果 AI 服务不可用，自动回退到简单的规则生成
- 支持错误处理和日志记录

### 3. 提示词设计

系统使用以下提示词与 AI 交互：

```
你是一个需求总结专家。请将以下用户需求文本总结为50字左右，使其易于阅读和理解。总结应简洁明了，并抓住应用或网站需求的主要内容。避免使用复杂的句子结构或技术术语。整个对话和指令都应以中文呈现。另外，给出符合要点的一到两个单词英文的标题，类似 GirlDress。输出json格式的结果，例如:
{"title": "GirlDress", "content": "女生装扮应用，分享效果，导入购物链接"}

用户需求：{用户输入的需求}
```

## 使用示例

### 测试 AI 功能

运行测试脚本：

```bash
cd backend
go run test_ai.go
```

### API 调用示例

创建项目时，系统会自动调用 AI 生成项目名称：

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "requirements": "我想创建一个女生装扮应用，用户可以上传自己的照片，然后选择不同的服装、发型、妆容等，生成装扮效果图"
  }'
```

预期响应：

```json
{
  "code": 0,
  "message": "项目创建成功",
  "data": {
    "id": "PROJ_00000000001",
    "name": "GirlDress",
    "description": "女生装扮应用，支持照片上传、服装搭配、效果生成和社交分享",
    "status": "draft",
    "requirements": "我想创建一个女生装扮应用...",
    "created_at": "2025-01-27T10:00:00Z"
  }
}
```

## 错误处理

### 1. AI 服务不可用

如果 Ollama 服务不可用，系统会自动回退到简单的规则生成：

```go
// 回退逻辑
func (g *projectNameGenerator) fallbackToDefaultConfig(requirements string, projectConfig *models.Project) bool {
    // 使用简单规则生成项目名
    projectName := g.generateSimpleProjectName(requirements)
    // ...
}
```

### 2. 连接超时

配置了 30 秒超时，避免长时间等待：

```go
HTTPClient: &http.Client{
    Timeout: 30 * time.Second,
}
```

### 3. 日志记录

所有 AI 调用都有详细的日志记录：

```go
logger.Info("AI 项目配置生成成功",
    logger.String("projectName", projectConfig.Name),
    logger.String("projectDescription", projectConfig.Description),
)
```

## 部署注意事项

1. **Docker 环境**: 确保 Ollama 服务在 Docker 网络中可访问
2. **模型下载**: 部署前需要下载所需的 AI 模型
3. **网络配置**: 确保 Traefik 路由配置正确
4. **监控**: 建议监控 AI 服务的可用性和响应时间

## 扩展功能

未来可以考虑添加：

1. **多模型支持**: 支持不同的 AI 模型
2. **缓存机制**: 缓存相似需求的结果
3. **用户偏好**: 根据用户历史偏好调整生成策略
4. **批量生成**: 支持批量生成多个项目名称选项
