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
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- 全局函数，用来按指定前缀生成表的 ID 字符串，例如 public.user 表，ID 为 "USER_000" + 'id'
CREATE OR REPLACE FUNCTION public.generate_table_id(IN prefix VARCHAR(32) DEFAULT 'DEFAULTID_', IN seq_name VARCHAR(50) DEFAULT 'default_id_num_seq')
    RETURNS VARCHAR(32)
    LANGUAGE 'plpgsql'
    VOLATILE 
AS $BODY$
DECLARE
    next_val BIGINT;
BEGIN
    next_val := nextval(seq_name);
    RETURN prefix || LPAD(next_val::TEXT, 11, '0');
END;
$BODY$;

ALTER FUNCTION public.generate_table_id(VARCHAR(32), VARCHAR(50))
    OWNER TO autocodeweb;

COMMENT ON FUNCTION public.generate_table_id(VARCHAR(32), VARCHAR(50))
    IS '获取ID的全局方法';

-- 创建用户ID序列
CREATE SEQUENCE IF NOT EXISTS public.users_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('USER', 'public.users_id_num_seq'),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建项目ID序列
CREATE SEQUENCE IF NOT EXISTS public.projects_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建项目表
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('PROJ', 'public.projects_id_num_seq'),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    requirements TEXT NOT NULL,
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'in_progress', 'completed', 'failed')),
    dev_status VARCHAR(50) DEFAULT 'pending',
    dev_progress INTEGER DEFAULT 0 CHECK (dev_progress >= 0 AND dev_progress <= 100),
    current_task_id VARCHAR(50),
    project_path VARCHAR(500) UNIQUE NOT NULL,
    backend_port INTEGER DEFAULT 8080 CHECK (backend_port >= 1024 AND backend_port <= 65535),
    frontend_port INTEGER DEFAULT 3000 CHECK (frontend_port >= 1024 AND frontend_port <= 65535),
    api_base_url VARCHAR(200) DEFAULT '/api/v1',
    app_secret_key VARCHAR(255),
    database_password VARCHAR(255),
    redis_password VARCHAR(255),
    jwt_secret_key VARCHAR(255),
    subnetwork VARCHAR(50) DEFAULT '172.20.0.0/16',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建任务ID序列
CREATE SEQUENCE IF NOT EXISTS public.tasks_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建任务表
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('TASK', 'public.tasks_id_num_seq'),
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    description TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建标签ID序列
CREATE SEQUENCE IF NOT EXISTS public.tags_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;
-- 创建标签表
CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('TAGS', 'public.tags_id_num_seq'),
    name VARCHAR(255) UNIQUE NOT NULL,
    color VARCHAR(7) DEFAULT '#666666',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建项目标签关联表
CREATE TABLE IF NOT EXISTS project_tags (
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id VARCHAR(50) NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, tag_id)
);

-- 创建任务日志ID序列
CREATE SEQUENCE IF NOT EXISTS public.task_logs_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;
    
-- 创建任务日志表
CREATE TABLE IF NOT EXISTS task_logs (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('LOGS', 'public.task_logs_id_num_seq'),
    task_id VARCHAR(50) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);

CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_task_logs_level ON task_logs(level);
CREATE INDEX IF NOT EXISTS idx_task_logs_created_at ON task_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- 插入默认标签
INSERT INTO tags (name, color) VALUES 
    ('Web应用', '#3B82F6'),
    ('移动应用', '#10B981'),
    ('桌面应用', '#F59E0B'),
    ('API服务', '#8B5CF6')
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
