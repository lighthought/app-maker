#!/bin/bash

# SSH 配置测试脚本
echo "=== SSH 配置测试 ==="

# 检查当前用户
echo "当前用户: $(whoami)"
echo "用户ID: $(id)"

# 检查 SSH 目录权限
SSH_DIR="/home/appuser/.ssh"
if [ -d "$SSH_DIR" ]; then
    echo "✓ SSH 目录存在: $SSH_DIR"
    echo "目录权限: $(ls -ld $SSH_DIR)"
else
    echo "✗ SSH 目录不存在: $SSH_DIR"
    echo "尝试创建目录..."
    mkdir -p "$SSH_DIR"
    chmod 700 "$SSH_DIR"
    echo "目录创建完成，权限: $(ls -ld $SSH_DIR)"
fi

# 检查 SSH 密钥是否存在
SSH_KEY_PATH=${SSH_KEY_PATH:-/home/appuser/.ssh/id_rsa}
if [ -f "$SSH_KEY_PATH" ]; then
    echo "✓ SSH 私钥存在: $SSH_KEY_PATH"
    echo "私钥权限: $(ls -l $SSH_KEY_PATH)"
else
    echo "✗ SSH 私钥不存在: $SSH_KEY_PATH"
fi

# 检查 SSH 公钥是否存在
if [ -f "$SSH_KEY_PATH.pub" ]; then
    echo "✓ SSH 公钥存在: $SSH_KEY_PATH.pub"
    echo "公钥权限: $(ls -l $SSH_KEY_PATH.pub)"
    echo "公钥内容:"
    cat "$SSH_KEY_PATH.pub"
else
    echo "✗ SSH 公钥不存在: $SSH_KEY_PATH.pub"
fi

# 检查 known_hosts 是否存在
SSH_KNOWN_HOSTS=${SSH_KNOWN_HOSTS:-/home/appuser/.ssh/known_hosts}
if [ -f "$SSH_KNOWN_HOSTS" ]; then
    echo "✓ known_hosts 存在: $SSH_KNOWN_HOSTS"
    echo "known_hosts 权限: $(ls -l $SSH_KNOWN_HOSTS)"
    echo "known_hosts 内容:"
    cat "$SSH_KNOWN_HOSTS"
else
    echo "✗ known_hosts 不存在: $SSH_KNOWN_HOSTS"
fi

# 测试 SSH 连接
GITLAB_URL=${GITLAB_URL:-git@gitlab.app-maker.localhost}
HOSTNAME=$(echo $GITLAB_URL | sed 's/git@//' | sed 's/:22//' | cut -d: -f1)

echo "测试 SSH 连接到: $HOSTNAME"
if ssh -o ConnectTimeout=10 -o BatchMode=yes -o StrictHostKeyChecking=no $HOSTNAME exit 2>/dev/null; then
    echo "✓ SSH 连接成功"
else
    echo "✗ SSH 连接失败"
fi

echo "=== 测试完成 ==="
