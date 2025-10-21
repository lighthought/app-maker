#!/bin/bash

# 简单的本地引用设置脚本

echo "=== 设置简单的本地共享模块引用 ==="

# 1. 初始化共享模块
echo "初始化共享模块..."
if [ ! -f "go.mod" ]; then
    go mod init shared-models
fi
go mod tidy

# 2. 设置 backend 项目
echo "设置 backend 项目..."
cd ../backend
if [ -f "go.mod" ]; then
    # 添加本地替换
    go mod edit -replace shared-models=../shared-models
    # 添加依赖
    go mod edit -require shared-models@v0.0.0
    go mod tidy
    echo "✅ backend 设置完成"
else
    echo "❌ backend/go.mod 不存在"
fi

# 3. 设置 agents 项目
echo "设置 agents 项目..."
cd ../agents
if [ -f "go.mod" ]; then
    # 添加本地替换
    go mod edit -replace shared-models=../shared-models
    # 添加依赖
    go mod edit -require shared-models@v0.0.0
    go mod tidy
    echo "✅ agents 设置完成"
else
    echo "❌ agents/go.mod 不存在"
fi

cd ../shared-models
echo ""
echo "🎉 设置完成！现在可以在项目中使用："
echo ""
echo "import ("
echo "    \"github.com/lighthought/app-maker/shared-models/agent\""
echo "    \"github.com/lighthought/app-maker/shared-models/common\""
echo "    \"github.com/lighthought/app-maker/shared-models/client\""
echo ")"
