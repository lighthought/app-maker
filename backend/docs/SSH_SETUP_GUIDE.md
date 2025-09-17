# SSH 配置使用说明

## 概述

系统已从使用 GitLab Access Token 改为使用 SSH 密钥进行 Git 操作。这样可以提供更好的安全性和便利性。

## 配置变更

### 1. Dockerfile 更新
- 添加了 `openssh-client` 包，提供 SSH 客户端和 `ssh-keygen` 工具
- 创建了 `/home/appuser/.ssh` 目录并设置正确的权限

### 2. GitService 更新
- 移除了 `gitlabToken` 字段
- 添加了 `sshKeyPath` 和 `sshKnownHosts` 字段
- 新增了 SSH 配置相关方法：
  - `SetupSSH()`: 配置 SSH 密钥和 known_hosts
  - `generateSSHKey()`: 生成 SSH 密钥对
  - `setupKnownHosts()`: 配置 SSH known_hosts
  - `GetPublicKey()`: 获取 SSH 公钥内容

### 3. Docker Compose 更新
- 移除了 `GITLAB_TOKEN` 环境变量
- 添加了 SSH 相关环境变量：
  - `SSH_KEY_PATH`: SSH 私钥路径（默认：`/home/appuser/.ssh/id_rsa`）
  - `SSH_KNOWN_HOSTS`: SSH known_hosts 文件路径（默认：`/home/appuser/.ssh/known_hosts`）
- 添加了 `ssh_keys` 卷挂载，持久化 SSH 密钥

## 环境变量配置

在 `.env` 文件中配置以下变量：

```bash
# GitLab 配置（SSH 格式）
GITLAB_URL=git@gitlab.app-maker.localhost
GITLAB_USERNAME=your-username
GITLAB_EMAIL=your-email@example.com

# SSH 配置（可选，使用默认值）
SSH_KEY_PATH=/home/appuser/.ssh/id_rsa
SSH_KNOWN_HOSTS=/home/appuser/.ssh/known_hosts
```

## 使用流程

### 1. 首次启动
当容器首次启动时，系统会：
1. 检查 SSH 密钥是否存在，如果不存在则自动生成
2. 配置 SSH known_hosts 文件
3. 在 Git 操作时使用 SSH 密钥进行认证

### 2. 获取公钥
如果需要将公钥添加到 GitLab，可以通过以下方式获取：

```bash
# 进入容器
docker exec -it app-maker-backend-dev bash

# 查看公钥内容
cat /home/appuser/.ssh/id_rsa.pub
```

### 3. 手动配置 SSH（可选）
如果需要使用现有的 SSH 密钥：

```bash
# 将私钥复制到容器
docker cp ~/.ssh/id_rsa app-maker-backend-dev:/home/appuser/.ssh/id_rsa
docker cp ~/.ssh/id_rsa.pub app-maker-backend-dev:/home/appuser/.ssh/id_rsa.pub

# 设置正确的权限
docker exec app-maker-backend-dev chmod 600 /home/appuser/.ssh/id_rsa
docker exec app-maker-backend-dev chmod 644 /home/appuser/.ssh/id_rsa.pub
```

## 测试 SSH 配置

运行测试脚本验证 SSH 配置：

```bash
# 进入容器
docker exec -it app-maker-backend-dev bash

# 运行测试脚本
/app/scripts/test-ssh-setup.sh
```

## GitLab 配置

1. 登录 GitLab
2. 进入 Settings > SSH Keys
3. 将公钥内容粘贴到 Key 字段
4. 添加描述并保存

## 故障排除

### 1. SSH 连接失败
- 检查 GitLab 服务器是否可访问
- 确认 SSH 密钥已添加到 GitLab
- 检查 known_hosts 文件是否正确配置

### 2. 权限问题（常见问题）
如果遇到 "Permission denied" 错误，按以下步骤解决：

#### 方法一：使用修复脚本
```bash
# 进入容器
docker exec -it app-maker-backend-dev bash

# 运行权限修复脚本
/app/scripts/fix-ssh-permissions.sh
```

#### 方法二：手动修复
```bash
# 进入容器
docker exec -it app-maker-backend-dev bash

# 修复目录权限
sudo chown -R appuser:appgroup /home/appuser/.ssh
sudo chmod 700 /home/appuser/.ssh

# 生成密钥
ssh-keygen -t rsa -b 4096 -f /home/appuser/.ssh/id_rsa -N ""

# 设置文件权限
chmod 600 /home/appuser/.ssh/id_rsa
chmod 644 /home/appuser/.ssh/id_rsa.pub
```

#### 权限要求
- SSH 目录权限：700 (drwx------)
- SSH 私钥权限：600 (-rw-------)
- SSH 公钥权限：644 (-rw-r--r--)

### 3. 密钥生成失败
- 检查容器是否有足够的权限
- 确认 openssh-client 包已正确安装
- 运行测试脚本检查权限：`/app/scripts/test-ssh-setup.sh`

## 安全注意事项

1. SSH 私钥应该妥善保管，不要泄露
2. 定期轮换 SSH 密钥
3. 使用强密码保护私钥（如果需要）
4. 限制 SSH 密钥的访问权限
