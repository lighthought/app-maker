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
    default_cli_tool VARCHAR(50) DEFAULT 'claude-code',
    default_ai_model VARCHAR(100) DEFAULT 'glm-4.6',
    default_model_provider VARCHAR(50) DEFAULT 'zhipu',
    default_model_api_url VARCHAR(500),
    default_api_token VARCHAR(500),
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
    guid VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    requirements TEXT NOT NULL,
    user_id VARCHAR(50) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'done', 'failed')),
    dev_status VARCHAR(50) DEFAULT 'pending',
    dev_progress INTEGER DEFAULT 0 CHECK (dev_progress >= 0 AND dev_progress <= 100),
    current_task_id VARCHAR(50),
    project_path VARCHAR(500) UNIQUE NOT NULL,
    backend_port INTEGER DEFAULT 9501 CHECK (backend_port >= 9501 AND backend_port <= 11500),
    frontend_port INTEGER DEFAULT 3501 CHECK (frontend_port >= 3501 AND frontend_port <= 5500),
    api_base_url VARCHAR(200) DEFAULT '/api/v1',
    app_secret_key VARCHAR(255),
    postgres_port INTEGER DEFAULT 5501 CHECK (postgres_port >= 5501 AND postgres_port <= 7500),
    database_password VARCHAR(255),
    redis_password VARCHAR(255),
    redis_port INTEGER DEFAULT 7501 CHECK (redis_port >= 7501 AND redis_port <= 9500),
    jwt_secret_key VARCHAR(255),
    subnetwork VARCHAR(50) DEFAULT '172.20.0.0/16',
    preview_url VARCHAR(500),
    gitlab_repo_url VARCHAR(500),
    cli_tool VARCHAR(50),
    ai_model VARCHAR(100),
    model_provider VARCHAR(50),
    model_api_url VARCHAR(500),
    api_token VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建对话消息ID序列
CREATE SEQUENCE IF NOT EXISTS public.project_msgs_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建对话消息表
CREATE TABLE IF NOT EXISTS project_msgs (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('MSG', 'public.project_msgs_id_num_seq'),
    project_guid VARCHAR(50),
    type VARCHAR(20) NOT NULL CHECK (type IN ('user', 'agent', 'system')),
    agent_role VARCHAR(20) CHECK (agent_role IN ('user', 'analyst', 'dev', 'pm', 'po', 'architect', 'ux-expert', 'qa', 'sm', 'bmad-master')),
    agent_name VARCHAR(100),
    content TEXT,
    is_markdown BOOLEAN DEFAULT FALSE,
    markdown_content TEXT,
    is_expanded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建开发阶段ID序列
CREATE SEQUENCE IF NOT EXISTS public.dev_stages_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建开发阶段表
CREATE TABLE IF NOT EXISTS dev_stages (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('STAGE', 'public.dev_stages_id_num_seq'),
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    project_guid VARCHAR(50),
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'done', 'failed')),
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    description TEXT,
    failed_reason TEXT,
    task_id VARCHAR(50),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建预览令牌ID序列
CREATE SEQUENCE IF NOT EXISTS public.preview_tokens_id_num_seq
    INCREMENT BY 1            -- 步长
    START 1                   -- 起始值    
    MINVALUE 1
    MAXVALUE 99999999999      -- 11位数字容量
    CACHE 1;

-- 创建预览令牌表
CREATE TABLE IF NOT EXISTS preview_tokens (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('PREV', 'public.preview_tokens_id_num_seq'),
    token VARCHAR(255) UNIQUE NOT NULL,
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
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
CREATE INDEX IF NOT EXISTS idx_users_default_cli_tool ON users(default_cli_tool);
CREATE INDEX IF NOT EXISTS idx_users_default_model_provider ON users(default_model_provider);
CREATE INDEX IF NOT EXISTS idx_users_default_api_token ON users(default_api_token);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);
CREATE INDEX IF NOT EXISTS idx_projects_cli_tool ON projects(cli_tool);
CREATE INDEX IF NOT EXISTS idx_projects_model_provider ON projects(model_provider);
CREATE INDEX IF NOT EXISTS idx_projects_api_token ON projects(api_token);

CREATE INDEX IF NOT EXISTS idx_project_msgs_project_guid ON project_msgs(project_guid);
CREATE INDEX IF NOT EXISTS idx_project_msgs_type ON project_msgs(type);
CREATE INDEX IF NOT EXISTS idx_project_msgs_created_at ON project_msgs(created_at);

CREATE INDEX IF NOT EXISTS idx_dev_stages_project_id ON dev_stages(project_id);
CREATE INDEX IF NOT EXISTS idx_dev_stages_status ON dev_stages(status);
CREATE INDEX IF NOT EXISTS idx_dev_stages_created_at ON dev_stages(created_at);

CREATE INDEX IF NOT EXISTS idx_preview_tokens_token ON preview_tokens(token);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_project_id ON preview_tokens(project_id);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_expires_at ON preview_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_created_at ON preview_tokens(created_at);


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
CREATE TRIGGER update_project_msgs_updated_at BEFORE UPDATE ON project_msgs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dev_stages_updated_at BEFORE UPDATE ON dev_stages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- Note: preview_tokens 表没有 updated_at 字段，所以不需要触发器

-- 显示创建的表
\dt

-- 显示创建的索引
\di

-- 显示创建的触发器
\dy
