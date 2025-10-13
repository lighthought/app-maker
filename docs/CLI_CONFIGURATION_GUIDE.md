# CLI 工具与 AI 模型配置指南

## 概述

App Maker 平台支持多种 CLI 工具和 AI 模型，让您可以根据项目需求选择最合适的开发工具和模型。本指南将帮助您配置和使用这些工具。

## 支持的 CLI 工具

### 1. Claude Code
- **描述**: Anthropic 官方的 CLI 工具，适合复杂的代码生成任务
- **安装**: `npx bmad-method install -f -i claude-code -d .`
- **最佳场景**: 大型项目、复杂业务逻辑、架构设计

### 2. Qwen Code
- **描述**: 阿里云通义千问的代码生成工具
- **安装**: `npx bmad-method install -f -i qwen-code -d .`
- **最佳场景**: 中文代码注释、中国本地化项目

### 3. iFlow CLI
- **描述**: 基于工作流的代码生成工具
- **安装**: `npx bmad-method install -f -i iflow-cli -d .`
- **最佳场景**: 标准化流程、模板驱动开发

### 4. Auggie CLI
- **描述**: 轻量级 AI 辅助编码工具
- **安装**: `npx bmad-method install -f -i auggie-cli -d .`
- **最佳场景**: 快速原型开发、小型项目

### 5. Gemini
- **描述**: Google 的多模态 AI 工具
- **安装**: `npx bmad-method install -f -i gemini -d .`
- **最佳场景**: 多媒体项目、创意应用

## 支持的模型提供商

### 1. Ollama (本地)

**推荐模型**:
- `qwen2.5-coder:14b` - 通义千问编码模型
- `deepseek-coder:14b` - DeepSeek 编码模型
- `codellama:13b` - Meta 的 CodeLlama

**安装步骤**:
```bash
# 1. 安装 Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# 2. 拉取模型
ollama pull qwen2.5-coder:14b
ollama pull deepseek-coder:14b

# 3. 验证安装
ollama list
```

**配置**:
- API URL: `http://localhost:11434`
- 默认模型: `qwen2.5-coder:14b`

**优势**:
- 完全本地运行，数据隐私有保障
- 无需API密钥
- 适合内网环境

### 2. Zhipu AI (智谱华章)

**推荐模型**:
- `glm-4.6` - 最新的GLM模型
- `codegeex-4` - 专业代码生成模型

**配置步骤**:
1. 注册账号: https://open.bigmodel.cn
2. 获取 API Key
3. 设置环境变量:
   ```bash
   export ZHIPU_API_KEY="your-api-key"
   ```

**配置**:
- API URL: `https://open.bigmodel.cn/api/anthropic`
- 默认模型: `glm-4.6`

**优势**:
- 中文理解能力强
- 性价比高
- 国内访问速度快

### 3. Anthropic (Claude)

**推荐模型**:
- `claude-sonnet-4` - 最新 Sonnet 模型
- `claude-opus-3` - 最强大的模型

**配置步骤**:
1. 注册账号: https://console.anthropic.com
2. 获取 API Key
3. 设置环境变量:
   ```bash
   export ANTHROPIC_API_KEY="your-api-key"
   ```

**配置**:
- API URL: `https://api.anthropic.com`
- 默认模型: `claude-sonnet-4`

**优势**:
- 代码质量最高
- 支持长上下文
- 推理能力强

### 4. OpenAI (GPT)

**推荐模型**:
- `gpt-4o` - 最新的 GPT-4 优化版本
- `gpt-4-turbo` - 性能优化版本

**配置步骤**:
1. 注册账号: https://platform.openai.com
2. 获取 API Key
3. 设置环境变量:
   ```bash
   export OPENAI_API_KEY="your-api-key"
   ```

**配置**:
- API URL: `https://api.openai.com/v1`
- 默认模型: `gpt-4o`

**优势**:
- 生态完善
- 社区支持好
- 多语言支持

### 5. vLLM (本地)

**推荐模型**:
- `deepseek-coder:14b`
- 自定义微调模型

**安装步骤**:
```bash
# 1. 安装 vLLM
pip install vllm

# 2. 启动服务
python -m vllm.entrypoints.openai.api_server \
  --model deepseek-ai/deepseek-coder-14b-instruct \
  --port 8000

# 3. 验证服务
curl http://localhost:8000/v1/models
```

**配置**:
- API URL: `http://localhost:8000`
- 默认模型: `deepseek-coder:14b`

**优势**:
- 高性能推理
- 支持自定义模型
- 适合企业部署

## 用户设置配置

### 在 Web 界面配置

1. 点击右上角用户头像
2. 选择"用户设置"
3. 滚动到"开发设置"部分
4. 配置以下选项:
   - **CLI 工具**: 选择您偏好的 CLI 工具
   - **模型提供商**: 选择模型提供商
   - **AI 模型**: 输入具体的模型名称
   - **API 地址**: 输入 API 端点 URL

### 配置示例

#### 使用本地 Ollama

```
CLI 工具: claude-code
模型提供商: ollama
AI 模型: qwen2.5-coder:14b
API 地址: http://localhost:11434
```

#### 使用智谱 AI

```
CLI 工具: claude-code
模型提供商: zhipu
AI 模型: glm-4.6
API 地址: https://open.bigmodel.cn/api/anthropic
```

#### 使用 Claude

```
CLI 工具: claude-code
模型提供商: anthropic
AI 模型: claude-sonnet-4
API 地址: https://api.anthropic.com
```

## 项目特定配置

如果您想为某个项目使用不同的配置:

1. 在项目创建时选择特定配置
2. 或在项目编辑页面修改配置
3. 项目配置会覆盖用户默认配置

## 本地开发快速入门

### 使用 Ollama 进行本地开发

这是最快的入门方式，无需 API 密钥:

```bash
# 1. 安装 Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# 2. 拉取推荐模型
ollama pull qwen2.5-coder:14b

# 3. 在 App Maker 中配置
# - CLI 工具: claude-code
# - 模型提供商: ollama
# - AI 模型: qwen2.5-coder:14b
# - API 地址: http://localhost:11434

# 4. 创建测试项目
# 建议先用小型项目测试
```

### 性能优化建议

1. **小项目测试**:
   - 先用简单项目（如 Todo App）测试配置
   - 验证 CLI 工具和模型正常工作

2. **模型选择**:
   - 快速原型: `qwen2.5-coder:14b` (Ollama)
   - 生产质量: `glm-4.6` (Zhipu) 或 `claude-sonnet-4` (Anthropic)
   - 复杂项目: `gpt-4o` (OpenAI) 或 `claude-opus-3` (Anthropic)

3. **本地 vs 云端**:
   - 本地模型适合快速迭代和隐私保护
   - 云端模型提供更高的代码质量和更强的推理能力

## 故障排查

### Ollama 连接失败

```bash
# 检查 Ollama 服务状态
systemctl status ollama

# 或手动启动
ollama serve

# 测试连接
curl http://localhost:11434/api/version
```

### API 密钥问题

确保环境变量正确设置:

```bash
# 检查环境变量
echo $ANTHROPIC_API_KEY
echo $OPENAI_API_KEY
echo $ZHIPU_API_KEY

# 或在项目中直接配置 API URL
```

### 模型下载缓慢

```bash
# Ollama 使用镜像源
export OLLAMA_HOST=https://mirrors.tuna.tsinghua.edu.cn/ollama

# 或使用代理
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
```

## 最佳实践

1. **开发阶段**:
   - 使用本地 Ollama 进行快速迭代
   - 选择较小的模型（14B 参数）

2. **测试阶段**:
   - 切换到云端模型进行质量检查
   - 使用 `glm-4.6` 或 `claude-sonnet-4`

3. **生产部署**:
   - 使用最强模型确保代码质量
   - 配置适当的 API 限流

4. **成本控制**:
   - 开发时使用本地模型
   - 关键功能使用云端模型
   - 合理设置 token 限制

## 模型性能对比

| 模型 | 速度 | 代码质量 | 中文支持 | 成本 |
|------|------|----------|----------|------|
| qwen2.5-coder:14b (Ollama) | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 免费 |
| deepseek-coder:14b (Ollama) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 免费 |
| glm-4.6 (Zhipu) | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| claude-sonnet-4 (Anthropic) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| gpt-4o (OpenAI) | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

## 常见问题

### Q: 如何选择合适的 CLI 工具？

A: 
- 如果使用 Anthropic Claude 模型，选择 `claude-code`
- 如果使用通义千问，选择 `qwen-code`
- 其他情况选择 `claude-code` 作为通用工具

### Q: 本地模型的硬件要求？

A:
- **14B 模型**: 至少 16GB RAM，推荐 32GB
- **GPU**: 推荐 NVIDIA GPU (8GB+ VRAM)
- **CPU**: 可以运行但速度较慢

### Q: 如何切换模型？

A: 
1. 在用户设置中修改默认配置
2. 新项目会自动使用新配置
3. 现有项目可以在项目设置中单独修改

### Q: 多个项目可以使用不同的模型吗？

A: 可以。每个项目都可以独立配置 CLI 工具和模型。

## 更多资源

- [Ollama 官方文档](https://ollama.ai/docs)
- [智谱 AI 文档](https://open.bigmodel.cn/dev/api)
- [Anthropic 文档](https://docs.anthropic.com)
- [OpenAI 文档](https://platform.openai.com/docs)
- [vLLM 文档](https://docs.vllm.ai)

## 社区支持

如有问题，欢迎通过以下方式联系:
- GitHub Issues: https://github.com/lighthought/app-maker/issues
- 邮箱: support@app-maker.com

