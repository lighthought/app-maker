#!/bin/bash

# SSH 权限修复脚本
echo "=== SSH 权限修复 ==="

# 检查当前用户
echo "当前用户: $(whoami)"
echo "用户ID: $(id)"

# 修复 SSH 目录权限
SSH_DIR="/home/appuser/.ssh"
echo "修复 SSH 目录权限: $SSH_DIR"

# 确保目录存在
mkdir -p "$SSH_DIR"

# 设置正确的权限
chmod 700 "$SSH_DIR"
chown -R appuser:appgroup "$SSH_DIR" 2>/dev/null || echo "无法设置所有者，可能权限不足"

echo "SSH 目录权限修复完成:"
ls -ld "$SSH_DIR"

# 如果存在密钥文件，修复其权限
SSH_KEY_PATH="/home/appuser/.ssh/id_rsa"
if [ -f "$SSH_KEY_PATH" ]; then
    echo "修复私钥权限: $SSH_KEY_PATH"
    chmod 600 "$SSH_KEY_PATH"
    chown appuser:appgroup "$SSH_KEY_PATH" 2>/dev/null || echo "无法设置私钥所有者"
    ls -l "$SSH_KEY_PATH"
fi

if [ -f "$SSH_KEY_PATH.pub" ]; then
    echo "修复公钥权限: $SSH_KEY_PATH.pub"
    chmod 644 "$SSH_KEY_PATH.pub"
    chown appuser:appgroup "$SSH_KEY_PATH.pub" 2>/dev/null || echo "无法设置公钥所有者"
    ls -l "$SSH_KEY_PATH.pub"
fi

# 尝试生成新的 SSH 密钥
echo "尝试生成新的 SSH 密钥..."
ssh-keygen -t rsa -b 4096 -f "$SSH_KEY_PATH" -N "" -C "app-maker-$(date +%s)"

if [ $? -eq 0 ]; then
    echo "✓ SSH 密钥生成成功"
    echo "私钥权限: $(ls -l $SSH_KEY_PATH)"
    echo "公钥权限: $(ls -l $SSH_KEY_PATH.pub)"
    echo ""
    echo "公钥内容:"
    cat "$SSH_KEY_PATH.pub"
else
    echo "✗ SSH 密钥生成失败"
fi

echo "=== 权限修复完成 ==="

