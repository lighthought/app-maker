# GitLab + Docker Compose 自动化部署架构

## 概述

本文档描述了 AutoCodeWeb 项目预览功能的 GitLab + Docker Compose 自动化部署架构方案。该方案通过 Git 版本控制和 CI/CD 流水线实现项目的自动化编译、测试和部署，适用于单主机环境。

## 整体架构图

```mermaid
graph TB
    subgraph "AutoCodeWeb Backend"
        A[项目生成服务] --> B[Git 操作服务]
        B --> C[GitLab API]
    end
    
    subgraph "GitLab 服务器"
        C --> D[Template Repository]
        C --> E[User Projects Repository]
        D --> F[Master Branch]
        E --> G[User Branches]
    end
    
    subgraph "GitLab CI/CD"
        H[Pipeline Job] --> I[构建阶段]
        I --> J[测试阶段]
        J --> K[部署阶段]
    end
    
    subgraph "部署环境"
        K --> L[Docker Registry]
        K --> M[Docker Compose]
        M --> N[项目容器]
    end
    
    subgraph "网络层"
        O[Nginx 反向代理] --> N
        P[主前端] --> Q[iframe 预览]
        Q --> O
    end
    
    F --> H
    G --> H
```

## 核心组件

### 1. GitLab 服务器配置

#### 1.1 Template Repository 结构

```
autocodeweb-template/
├── frontend/              # 前端模板
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── backend/               # 后端模板
│   ├── cmd/
│   ├── internal/
│   ├── go.mod
│   └── Dockerfile
├── docker-compose.yml     # 容器配置
├── .gitlab-ci.yml         # CI/CD 配置
├── .dockerignore
└── README.md
```

#### 1.2 User Projects Repository 结构

```
autocodeweb-projects/
├── {user_id}/
│   ├── {project_id}_001/
│   ├── {project_id}_002/
│   └── ...
```

#### 1.3 GitLab API Token 配置

```yaml
# backend/configs/config.yaml
gitlab:
  url: "http://your-gitlab-server"
  token: "glpat-xxxxxxxxxxxxxxxxxxxx"  # 系统级 Token
  template_repo: "autocodeweb-template"
  projects_repo: "autocodeweb-projects"
  webhook_url: "http://jenkins-server/webhook/gitlab"
```

### 2. 后端 Git 服务实现

#### 2.1 Git 服务接口

```go
type GitService interface {
    // 从模板仓库克隆项目
    CloneFromTemplate(ctx context.Context, project *Project) error
    
    // 替换项目中的占位符
    ReplacePlaceholders(ctx context.Context, project *Project) error
    
    // 提交到用户项目仓库
    PushToUserRepo(ctx context.Context, project *Project) error
    
    // 获取项目部署状态
    GetDeploymentStatus(ctx context.Context, project *Project) (*DeploymentStatus, error)
}
```

#### 2.2 Git 操作实现

```go
type gitService struct {
    gitlabClient *gitlab.Client
    templateRepo string
    projectsRepo string
    baseDir      string
}

func (s *gitService) CloneFromTemplate(ctx context.Context, project *Project) error {
    projectPath := fmt.Sprintf("%s/%s/%s", s.baseDir, project.UserID, project.ID)
    
    // 克隆模板仓库
    cmd := exec.CommandContext(ctx, "git", "clone", 
        fmt.Sprintf("http://gitlab-server/autocodeweb-template.git"),
        projectPath)
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to clone template: %w", err)
    }
    
    // 创建用户分支
    cmd = exec.CommandContext(ctx, "git", "checkout", "-b", 
        fmt.Sprintf("%s/%s", project.UserID, project.ID))
    cmd.Dir = projectPath
    
    return cmd.Run()
}

func (s *gitService) ReplacePlaceholders(ctx context.Context, project *Project) error {
    projectPath := fmt.Sprintf("%s/%s/%s", s.baseDir, project.UserID, project.ID)
    
    // 替换配置文件中的占位符
    replacements := map[string]string{
        "{{PROJECT_NAME}}": project.Name,
        "{{PROJECT_ID}}": project.ID,
        "{{BACKEND_PORT}}": strconv.Itoa(project.BackendPort),
        "{{FRONTEND_PORT}}": strconv.Itoa(project.FrontendPort),
        "{{API_BASE_URL}}": project.ApiBaseUrl,
        "{{SECRET_KEY}}": project.AppSecretKey,
        "{{DB_PASSWORD}}": project.DatabasePassword,
        "{{REDIS_PASSWORD}}": project.RedisPassword,
        "{{JWT_SECRET}}": project.JwtSecretKey,
    }
    
    return s.replaceInFiles(projectPath, replacements)
}

func (s *gitService) PushToUserRepo(ctx context.Context, project *Project) error {
    projectPath := fmt.Sprintf("%s/%s/%s", s.baseDir, project.UserID, project.ID)
    
    // 添加所有文件
    cmd := exec.CommandContext(ctx, "git", "add", ".")
    cmd.Dir = projectPath
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to add files: %w", err)
    }
    
    // 提交更改
    cmd = exec.CommandContext(ctx, "git", "commit", "-m", 
        fmt.Sprintf("Initialize project %s", project.Name))
    cmd.Dir = projectPath
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to commit: %w", err)
    }
    
    // 推送到用户仓库
    remoteURL := fmt.Sprintf("http://gitlab-server/autocodeweb-projects.git")
    cmd = exec.CommandContext(ctx, "git", "push", "origin", 
        fmt.Sprintf("%s/%s", project.UserID, project.ID))
    cmd.Dir = projectPath
    
    return cmd.Run()
}
```

### 3. GitLab CI/CD 配置

#### 3.1 Template Repository 的 .gitlab-ci.yml

```yaml
# .gitlab-ci.yml
stages:
  - build
  - test
  - deploy

variables:
  DOCKER_REGISTRY: "your-registry.com"
  DEPLOY_HOST: "your-deploy-host"
  DEPLOY_USER: "deploy-user"
  PROJECTS_BASE_DIR: "/opt/autocodeweb-projects"

# 只对用户分支触发
workflow:
  rules:
    - if: $CI_COMMIT_BRANCH =~ /^[^\/]+\/[^\/]+$/

# 使用 Docker-in-Docker 执行器
image: docker:latest
services:
  - docker:dind

before_script:
  - apk add --no-cache make git openssh-client
  - echo "开始构建项目..."

build-project:
  stage: build
  script:
    - |
      # 解析分支名获取用户ID和项目ID
      BRANCH_PARTS=(${CI_COMMIT_BRANCH//\// })
      USER_ID=${BRANCH_PARTS[0]}
      PROJECT_ID=${BRANCH_PARTS[1]}
      
      echo "构建用户项目: ${USER_ID}/${PROJECT_ID}"
      
      # 使用 make build-prod 构建生产环境镜像
      make build-prod
      
      # 重命名镜像并推送到 Registry
      docker tag autocodeweb-project:latest ${DOCKER_REGISTRY}/autocodeweb-project:${CI_COMMIT_SHA}
      docker push ${DOCKER_REGISTRY}/autocodeweb-project:${CI_COMMIT_SHA}
      
      echo "构建完成，镜像: ${DOCKER_REGISTRY}/autocodeweb-project:${CI_COMMIT_SHA}"
  artifacts:
    paths:
      - "*.log"
    expire_in: 1 hour
  only:
    - branches

test-project:
  stage: test
  image: node:18
  script:
    - cd frontend && npm test
    - cd backend && go test ./...
  only:
    - branches

deploy-project:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache openssh-client docker-compose
  script:
    - |
      # 解析分支名获取用户ID和项目ID
      BRANCH_PARTS=(${CI_COMMIT_BRANCH//\// })
      USER_ID=${BRANCH_PARTS[0]}
      PROJECT_ID=${BRANCH_PARTS[1]}
      
      echo "部署项目: ${USER_ID}/${PROJECT_ID}"
      
      # 生成 Docker Compose 配置
      cat > docker-compose-${USER_ID}-${PROJECT_ID}.yml << EOF
      version: '3.8'
      
      services:
        project-${USER_ID}-${PROJECT_ID}:
          image: ${DOCKER_REGISTRY}/autocodeweb-project:${CI_COMMIT_SHA}
          container_name: project-${USER_ID}-${PROJECT_ID}
          restart: unless-stopped
          ports:
            - "${PROJECT_ID}:8080"
          environment:
            - APP_ENVIRONMENT=production
            - DATABASE_HOST=postgres
            - REDIS_HOST=redis
          networks:
            - autocodeweb-network
          labels:
            - "traefik.enable=true"
            - "traefik.http.routers.project-${USER_ID}-${PROJECT_ID}.rule=Host(\`${USER_ID}-${PROJECT_ID}.your-domain.com\`)"
            - "traefik.http.routers.project-${USER_ID}-${PROJECT_ID}.entrypoints=web"
            - "traefik.http.services.project-${USER_ID}-${PROJECT_ID}.loadbalancer.server.port=8080"
      
      networks:
        autocodeweb-network:
          external: true
      EOF
      
      # 复制配置文件到部署主机
      scp -o StrictHostKeyChecking=no docker-compose-${USER_ID}-${PROJECT_ID}.yml ${DEPLOY_USER}@${DEPLOY_HOST}:${PROJECTS_BASE_DIR}/
      
      # 在部署主机上启动服务
      ssh -o StrictHostKeyChecking=no ${DEPLOY_USER}@${DEPLOY_HOST} << 'EOF'
        cd ${PROJECTS_BASE_DIR}
        
        # 停止旧容器（如果存在）
        docker-compose -f docker-compose-${USER_ID}-${PROJECT_ID}.yml down || true
        
        # 拉取最新镜像
        docker pull ${DOCKER_REGISTRY}/autocodeweb-project:${CI_COMMIT_SHA}
        
        # 启动新容器
        docker-compose -f docker-compose-${USER_ID}-${PROJECT_ID}.yml up -d
        
        # 等待服务启动
        sleep 10
        
        # 检查服务状态
        if docker-compose -f docker-compose-${USER_ID}-${PROJECT_ID}.yml ps | grep -q "Up"; then
          echo "部署成功"
        else
          echo "部署失败"
          exit 1
        fi
      EOF
      
      # 通知 AutoCodeWeb 后端部署成功
      DEPLOY_URL="http://${USER_ID}-${PROJECT_ID}.your-domain.com"
      curl -X POST "http://autocodeweb-backend/api/v1/projects/${PROJECT_ID}/deployment-status" \
        -H "Content-Type: application/json" \
        -d "{\"status\":\"deployed\",\"url\":\"${DEPLOY_URL}\",\"jenkins_build\":\"${CI_PIPELINE_ID}\"}"
      
      echo "部署完成，访问地址: ${DEPLOY_URL}"
  only:
    - branches
```

### 4. Docker Compose 网络配置

#### 4.1 创建外部网络

```bash
# 在部署主机上创建网络
docker network create autocodeweb-network
```

#### 4.2 Traefik 反向代理配置

**为什么需要 Traefik 反向代理？**

在项目部署方案中，每个用户项目都会分配不同的端口（如 8083、3004、5001 等）。直接访问这些端口会带来以下问题：

**问题场景**：
```bash
# 不使用反向代理的访问方式
http://your-domain.com:8083  # 用户A的项目
http://your-domain.com:3004  # 用户B的项目  
http://your-domain.com:5001  # 用户C的项目
```

**问题**：
- 用户需要记住端口号
- URL 不够友好
- 端口冲突管理复杂
- 难以实现 SSL 证书管理

**Traefik 解决方案**：
```mermaid
graph TB
    subgraph "Traefik 解决方案"
        A[Traefik 反向代理] --> B[用户A项目<br/>端口8083]
        A --> C[用户B项目<br/>端口3004]
        A --> D[用户C项目<br/>端口5001]
    end
    
    subgraph "友好访问"
        E[userA-project.your-domain.com] --> A
        F[userB-project.your-domain.com] --> A
        G[userC-project.your-domain.com] --> A
    end
```
```bash
# 使用 Traefik 的友好访问方式
https://userA-proj001.your-domain.com  # 自动路由到端口8083
https://userB-proj002.your-domain.com  # 自动路由到端口3004
https://userC-proj003.your-domain.com  # 自动路由到端口5001
```

**Traefik 的核心作用**：
1. **域名路由**：将友好的域名映射到内部端口
2. **SSL 管理**：自动为每个项目申请 HTTPS 证书
3. **服务发现**：自动发现和配置新部署的项目
4. **统一入口**：提供统一的 80/443 端口访问
5. **运维简化**：减少手动配置和证书管理

```yaml
# docker-compose-traefik.yml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"  # Traefik Dashboard
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./acme.json:/acme.json
    networks:
      - autocodeweb-network
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik.rule=Host(`traefik.your-domain.com`)"
      - "traefik.http.routers.traefik.service=api@internal"
      - "traefik.http.routers.traefik.middlewares=auth"
      - "traefik.http.middlewares.auth.basicauth.users=admin:$$2y$$10$$..."

networks:
  autocodeweb-network:
    external: true
```

#### 4.3 Traefik 配置文件

```yaml
# traefik.yml
api:
  dashboard: true
  insecure: false

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entrypoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false  # 只暴露有标签的容器
    network: autocodeweb-network

certificatesResolvers:
  letsencrypt:
    acme:
      email: your-email@example.com
      storage: acme.json
      httpChallenge:
        entryPoint: web
```

**Traefik 自动行为**：
1. **监听 Docker 事件**：自动发现新启动的容器
2. **服务发现**：识别带有 `traefik.enable=true` 标签的容器
3. **路由配置**：根据 `Host` 规则创建路由规则
4. **SSL 证书**：自动为每个域名申请 Let's Encrypt 证书
5. **请求转发**：将请求转发到对应的项目容器

**项目容器标签说明**：
```yaml
labels:
  - "traefik.enable=true"                                    # 启用 Traefik 路由
  - "traefik.http.routers.project-${USER_ID}-${PROJECT_ID}.rule=Host(`${USER_ID}-${PROJECT_ID}.your-domain.com`)"  # 域名规则
  - "traefik.http.routers.project-${USER_ID}-${PROJECT_ID}.entrypoints=web"  # 使用 web 入口点
  - "traefik.http.services.project-${USER_ID}-${PROJECT_ID}.loadbalancer.server.port=8080"  # 转发到容器内端口
  - "traefik.http.routers.project-${USER_ID}-${PROJECT_ID}.tls.certresolver=letsencrypt"  # 使用 Let's Encrypt 证书
```

### 5. 后端集成

#### 5.1 扩展项目模型

```go
type Project struct {
    // ... 现有字段
    GitBranch      string     `json:"git_branch"`
    DeployedUrl    string     `json:"deployed_url"`
    DeploymentStatus string   `json:"deployment_status"` // pending, deployed, failed
    JenkinsBuild   string     `json:"jenkins_build"`
    DeployedAt     *time.Time `json:"deployed_at"`
    ContainerPort  int        `json:"container_port"`
}
```

#### 5.2 部署状态回调接口

```go
// POST /api/v1/projects/{projectID}/deployment-status
type DeploymentStatusRequest struct {
    Status        string `json:"status"`
    URL           string `json:"url"`
    JenkinsBuild  string `json:"jenkins_build"`
    ErrorMessage  string `json:"error_message,omitempty"`
}

func (h *ProjectHandler) UpdateDeploymentStatus(c *gin.Context) {
    projectID := c.Param("projectID")
    var req DeploymentStatusRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Code: 400, Message: "Invalid request"})
        return
    }
    
    project, err := h.projectService.GetProject(c, projectID, getUserID(c))
    if err != nil {
        c.JSON(404, ErrorResponse{Code: 404, Message: "Project not found"})
        return
    }
    
    project.DeploymentStatus = req.Status
    project.DeployedUrl = req.URL
    project.JenkinsBuild = req.JenkinsBuild
    
    if req.Status == "deployed" {
        now := time.Now()
        project.DeployedAt = &now
    }
    
    if err := h.projectService.UpdateProject(c, projectID, project, getUserID(c)); err != nil {
        c.JSON(500, ErrorResponse{Code: 500, Message: "Failed to update project"})
        return
    }
    
    c.JSON(200, Response{Code: 0, Message: "success"})
}
```

### 6. 前端预览集成

```javascript
// 项目预览组件
const ProjectPreview = ({ project }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [previewUrl, setPreviewUrl] = useState(project.deployed_url);
    
    const startPreview = async () => {
        setIsLoading(true);
        try {
            // 触发部署
            const response = await fetch(`/api/v1/projects/${project.id}/deploy`);
            const result = await response.json();
            
            // 轮询部署状态
            const pollStatus = async () => {
                const statusResponse = await fetch(`/api/v1/projects/${project.id}`);
                const projectData = await statusResponse.json();
                
                if (projectData.data.deployment_status === 'deployed') {
                    setPreviewUrl(projectData.data.deployed_url);
                    setIsLoading(false);
                } else if (projectData.data.deployment_status === 'failed') {
                    alert('部署失败');
                    setIsLoading(false);
                } else {
                    setTimeout(pollStatus, 5000); // 5秒后重试
                }
            };
            
            pollStatus();
        } catch (error) {
            console.error('部署失败:', error);
            setIsLoading(false);
        }
    };
    
    return (
        <div>
            {!previewUrl && (
                <button onClick={startPreview} disabled={isLoading}>
                    {isLoading ? '部署中...' : '启动预览'}
                </button>
            )}
            
            {previewUrl && (
                <iframe 
                    src={previewUrl}
                    width="100%"
                    height="600px"
                    title="项目预览"
                />
            )}
        </div>
    );
};
```

## 部署流程

### 1. 项目创建流程

```mermaid
sequenceDiagram
    participant User
    participant Backend
    participant GitLab
    participant DeployHost
    participant Traefik
    
    User->>Backend: 创建项目
    Backend->>GitLab: 克隆模板仓库
    Backend->>Backend: 替换占位符
    Backend->>GitLab: 推送到用户分支
    GitLab->>GitLab: 触发构建
    GitLab->>DeployHost: 部署 Docker Compose
    DeployHost->>Traefik: 注册服务
    GitLab->>Backend: 回调部署状态
    Backend->>User: 返回预览URL
```

### 2. 分支命名规范

- 格式：`{user_id}/{project_id}`
- 示例：`USER00000000002/PROJ00000000001`
- 用途：用于识别用户和项目，触发对应的 CI/CD 流水线

### 3. URL 生成规则

- 格式：`http://{user_id}-{project_id}.your-domain.com`
- 示例：`http://USER00000000002-PROJ00000000001.your-domain.com`
- 配置：通过 Traefik 自动生成

**URL 生成流程**：
1. **项目创建时**：分配内部端口（如 8083、3004）
2. **域名生成**：根据用户ID和项目ID生成友好域名
3. **Traefik 路由**：自动将域名映射到内部端口
4. **SSL 证书**：自动申请 HTTPS 证书

**示例**：
```bash
# 项目分配端口
用户A项目001 -> 端口8083
用户B项目002 -> 端口3004

# 生成的访问URL
https://USER00000000002-PROJ00000000001.your-domain.com  # 自动路由到端口8083
https://USER00000000003-PROJ00000000002.your-domain.com  # 自动路由到端口3004
```

## 单主机环境配置

### 1. 部署主机准备

```bash
# 安装 Docker 和 Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 创建项目目录
sudo mkdir -p /opt/autocodeweb-projects
sudo chown $USER:$USER /opt/autocodeweb-projects

# 创建网络
docker network create autocodeweb-network
```

### 2. SSH 密钥配置

```bash
# 在 GitLab Runner 上生成 SSH 密钥
ssh-keygen -t rsa -b 4096 -C "gitlab-runner@your-domain.com"

# 复制公钥到部署主机
ssh-copy-id deploy-user@your-deploy-host
```

### 3. 端口管理

```bash
# 端口分配脚本
#!/bin/bash
# allocate-port.sh

PORT_RANGE_START=8001
PORT_RANGE_END=9000

allocate_port() {
    local user_id=$1
    local project_id=$2
    
    # 生成基于用户ID和项目ID的端口
    local hash=$(echo "${user_id}${project_id}" | md5sum | cut -d' ' -f1)
    local port=$((PORT_RANGE_START + 0x${hash:0:4} % (PORT_RANGE_END - PORT_RANGE_START)))
    
    echo $port
}

# 使用示例
PORT=$(allocate_port "USER00000000002" "PROJ00000000001")
echo "分配端口: $PORT"
```

## 优势分析

### 1. 简化部署
- 使用 Docker Compose 替代 Kubernetes，降低复杂度
- 单主机部署，减少资源消耗
- 配置简单，易于维护

### 2. 资源效率
- 不需要 Kubernetes 集群
- 减少内存和 CPU 开销
- 适合中小规模部署

### 3. 易于调试
- 可以直接访问容器日志
- 简单的网络配置
- 快速的问题排查

### 4. 成本效益
- 单主机部署成本低
- 减少运维复杂度
- 适合开发和生产环境

### 5. 用户体验优化
- **友好的域名访问**：`https://userA-proj001.your-domain.com` 替代 `http://your-domain.com:8083`
- **自动 HTTPS**：每个项目自动获得 SSL 证书
- **统一入口**：所有项目通过 80/443 端口访问
- **无需记忆端口**：用户只需要记住项目域名

### 6. 运维管理简化
- **自动服务发现**：新项目部署后自动配置路由
- **SSL 证书管理**：自动申请和续期证书
- **端口冲突避免**：动态端口分配，避免冲突
- **统一监控**：通过 Traefik Dashboard 监控所有项目

## 注意事项

### 1. 安全考虑
- GitLab Token 需要适当的权限范围
- SSH 密钥需要安全存储
- 容器间需要网络隔离

### 2. 资源管理
- 需要合理分配端口范围
- 及时清理未使用的容器
- 监控磁盘和内存使用情况

### 3. 网络配置
- 确保 GitLab、部署主机之间网络连通
- 配置正确的 DNS 解析
- 设置适当的防火墙规则

### 4. 备份策略
- 定期备份项目数据
- 备份 Docker Compose 配置
- 备份 Traefik 配置

### 5. Traefik 相关注意事项
- **域名解析**：确保所有子域名都指向部署主机
- **证书申请**：Let's Encrypt 有频率限制，避免频繁重新部署
- **容器标签**：确保项目容器正确配置 Traefik 标签
- **网络连通**：确保 Traefik 容器与项目容器在同一网络
- **Dashboard 安全**：配置 Traefik Dashboard 的访问认证

## 总结

GitLab + Docker Compose 自动化部署架构为 AutoCodeWeb 项目预览功能提供了适合单主机环境的解决方案。该方案不仅解决了容器内运行项目的技术限制，还提供了简化的部署流程和较低的资源消耗，特别适合中小规模的部署场景。