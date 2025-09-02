-- AutoCodeWeb 数据库初始化脚本
-- 创建数据库和用户

-- 创建数据库（如果不存在）
SELECT 'CREATE DATABASE autocodeweb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'autocodeweb')\gexec

-- 连接到新创建的数据库
\c autocodeweb;

-- 创建用户（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'autocodeweb') THEN
        CREATE ROLE autocodeweb WITH LOGIN PASSWORD 'AutoCodeWeb2024!@#';
    END IF;
END
$$;

-- 给用户授权
GRANT ALL PRIVILEGES ON DATABASE autocodeweb TO autocodeweb;
GRANT ALL PRIVILEGES ON SCHEMA public TO autocodeweb;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO autocodeweb;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO autocodeweb;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO autocodeweb;

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建项目表
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'in_progress', 'completed', 'failed')),
    requirements TEXT,
    project_path VARCHAR(500) UNIQUE,
    backend_port INTEGER DEFAULT 8080 CHECK (backend_port >= 1024 AND backend_port <= 65535),
    frontend_port INTEGER DEFAULT 3000 CHECK (frontend_port >= 1024 AND frontend_port <= 65535),
    api_base_url VARCHAR(100) DEFAULT '/api/v1',
    app_secret_key VARCHAR(255),
    database_password VARCHAR(255),
    redis_password VARCHAR(255),
    jwt_secret_key VARCHAR(255),
    subnetwork VARCHAR(50) DEFAULT '172.20.0.0/16',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建任务表
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    priority INTEGER DEFAULT 2,
    dependencies TEXT[],
    max_retries INTEGER DEFAULT 3,
    retry_count INTEGER DEFAULT 0,
    retry_delay INTEGER DEFAULT 60,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    deadline TIMESTAMP,
    result TEXT,
    error_message TEXT,
    metadata TEXT,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建标签表
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    color VARCHAR(7) DEFAULT '#666666',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建项目标签关联表
CREATE TABLE IF NOT EXISTS project_tags (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, tag_id)
);

-- 创建任务日志表
CREATE TABLE IF NOT EXISTS task_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建任务依赖关系表
CREATE TABLE IF NOT EXISTS task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    dependency_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, dependency_id)
);

-- 插入默认管理员用户
-- 密码: Admin123!@# (使用 pgcrypto 加密)
INSERT INTO users (email, username, password, role, status) VALUES 
('admin@autocodeweb.com', 'admin', crypt('Admin123!@#', gen_salt('bf')), 'admin', 'active')
ON CONFLICT (email) DO NOTHING;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);

CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline);

CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_task_logs_level ON task_logs(level);
CREATE INDEX IF NOT EXISTS idx_task_logs_created_at ON task_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_dependencies_dependency_id ON task_dependencies(dependency_id);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- 插入默认标签
INSERT INTO tags (name, color) VALUES 
    ('Web应用', '#3B82F6'),
    ('移动应用', '#10B981'),
    ('桌面应用', '#F59E0B'),
    ('API服务', '#8B5CF6'),
    ('数据库', '#EF4444'),
    ('机器学习', '#06B6D4'),
    ('区块链', '#84CC16'),
    ('游戏开发', '#F97316')
ON CONFLICT (name) DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表添加更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 显示创建的表
\dt

-- 显示创建的索引
\di

-- 显示创建的触发器
\dy
