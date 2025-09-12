@echo off
echo Starting Agents Server...

REM 设置环境变量
set NODE_ENV=development
set PORT=3001
set REDIS_URL=redis://localhost:6379
set BACKEND_API_URL=http://localhost:8080
set PROJECT_DATA_PATH=F:/app-maker/app_data/projects
set LOG_LEVEL=info

REM 检查Node.js是否安装
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Node.js is not installed or not in PATH
    pause
    exit /b 1
)

REM 检查pnpm是否安装
pnpm --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: pnpm is not installed. Please run: npm install -g pnpm
    pause
    exit /b 1
)

REM 检查claude是否安装
claude --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: claude is not installed. Please run: npm install -g claude
    pause
    exit /b 1
)

REM 安装依赖（如果node_modules不存在）
if not exist "node_modules" (
    echo Installing dependencies...
    pnpm install
    if %errorlevel% neq 0 (
        echo Error: Failed to install dependencies
        pause
        exit /b 1
    )
)

REM 构建项目
echo Building project...
pnpm build
if %errorlevel% neq 0 (
    echo Error: Build failed
    pause
    exit /b 1
)

REM 启动服务
echo Starting Agents Server...
pnpm start

pause
