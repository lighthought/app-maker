# AutoCodeWeb 数据模型设计

## 1. 数据模型概述

### 1.1 设计理念
- **规范化设计**：遵循数据库设计范式，减少数据冗余
- **扩展性考虑**：支持未来功能扩展和业务增长
- **性能优化**：合理的索引设计和查询优化
- **数据完整性**：外键约束和业务规则约束
- **审计追踪**：记录数据创建、修改和删除历史

### 1.2 技术选型
- **数据库**：PostgreSQL 15+
- **ORM框架**：GORM 1.25+
- **迁移工具**：GORM Auto Migration
- **索引策略**：B-tree、Hash、GIN等
- **分区策略**：按时间分区（可选）

## 2. 核心数据模型

### 2.1 用户管理模型
```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('user', 'admin', 'moderator')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户会话表
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户权限表
CREATE TABLE user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE
);
```

### 2.2 项目管理模型
```sql
-- 项目表
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN (
        'draft', 'planning', 'development', 'testing', 'completed', 'deployed', 'archived'
    )),
    project_type VARCHAR(50) NOT NULL CHECK (project_type IN ('web', 'mobile', 'desktop', 'api')),
    figma_url VARCHAR(500),
    requirements TEXT,
    tech_stack JSONB,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- 项目成员表
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

-- 项目标签表
CREATE TABLE project_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, tag)
);
```

### 2.3 Agent协作模型
```sql
-- Agent会话表
CREATE TABLE agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_type VARCHAR(50) NOT NULL CHECK (session_type IN ('pm', 'ux', 'architect', 'po', 'qa')),
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'failed')),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Agent对话消息表
CREATE TABLE agent_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    message_type VARCHAR(20) NOT NULL CHECK (message_type IN ('user', 'agent', 'system')),
    agent_role VARCHAR(50),
    content TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Agent工作成果表
CREATE TABLE agent_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    artifact_type VARCHAR(50) NOT NULL CHECK (artifact_type IN (
        'prd', 'ux_spec', 'architecture', 'epics', 'stories', 'test_plan'
    )),
    content JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'review', 'approved', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 2.4 文档管理模型
```sql
-- 文档表
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    document_type VARCHAR(50) NOT NULL CHECK (document_type IN (
        'prd', 'ux_spec', 'architecture', 'api_spec', 'database_schema', 'test_plan'
    )),
    format VARCHAR(20) DEFAULT 'markdown' CHECK (format IN ('markdown', 'json', 'yaml', 'html')),
    version INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'review', 'approved', 'published')),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 文档版本表
CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    change_log TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(document_id, version)
);

-- 文档评论表
CREATE TABLE document_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id UUID REFERENCES document_comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 2.5 任务执行模型
```sql
-- 任务表
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_type VARCHAR(50) NOT NULL CHECK (task_type IN (
        'framework_setup', 'code_generation', 'testing', 'deployment', 'documentation'
    )),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN (
        'pending', 'running', 'completed', 'failed', 'cancelled'
    )),
    priority INTEGER DEFAULT 1 CHECK (priority BETWEEN 1 AND 5),
    assigned_to UUID REFERENCES users(id),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    progress INTEGER DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- 任务日志表
CREATE TABLE task_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    log_level VARCHAR(20) NOT NULL CHECK (log_level IN ('debug', 'info', 'warning', 'error')),
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### 2.6 部署和预览模型
```sql
-- 部署环境表
CREATE TABLE deployment_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    environment_type VARCHAR(50) NOT NULL CHECK (environment_type IN ('development', 'staging', 'production')),
    url VARCHAR(500),
    status VARCHAR(50) DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'maintenance')),
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 部署记录表
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES deployment_environments(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN (
        'pending', 'building', 'deploying', 'success', 'failed', 'rolled_back'
    )),
    build_log TEXT,
    deployment_log TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL REFERENCES users(id)
);

-- 预览配置表
CREATE TABLE preview_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    config_type VARCHAR(50) NOT NULL CHECK (config_type IN ('web', 'mobile', 'desktop')),
    settings JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 3. 索引设计

### 3.1 性能优化索引
```sql
-- 用户表索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 项目表索引
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_created_at ON projects(created_at);
CREATE INDEX idx_projects_project_type ON projects(project_type);

-- Agent会话索引
CREATE INDEX idx_agent_sessions_project_id ON agent_sessions(project_id);
CREATE INDEX idx_agent_sessions_status ON agent_sessions(status);
CREATE INDEX idx_agent_sessions_started_at ON agent_sessions(started_at);

-- 任务表索引
CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- 文档表索引
CREATE INDEX idx_documents_project_id ON documents(project_id);
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_created_at ON documents(created_at);

-- 部署表索引
CREATE INDEX idx_deployments_project_id ON deployments(project_id);
CREATE INDEX idx_deployments_environment_id ON deployments(environment_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_started_at ON deployments(started_at);
```

### 3.2 全文搜索索引
```sql
-- 项目全文搜索索引
CREATE INDEX idx_projects_fulltext ON projects USING GIN (
    to_tsvector('english', name || ' ' || COALESCE(description, ''))
);

-- 文档全文搜索索引
CREATE INDEX idx_documents_fulltext ON documents USING GIN (
    to_tsvector('english', title || ' ' || COALESCE(content, ''))
);

-- 任务全文搜索索引
CREATE INDEX idx_tasks_fulltext ON tasks USING GIN (
    to_tsvector('english', title || ' ' || COALESCE(description, ''))
);
```

## 4. 约束和触发器

### 4.1 数据完整性约束
```sql
-- 项目状态变更约束
ALTER TABLE projects ADD CONSTRAINT chk_project_status_transition 
CHECK (
    (status = 'draft' AND updated_at = created_at) OR
    (status IN ('planning', 'development', 'testing', 'completed', 'deployed', 'archived'))
);

-- 任务进度约束
ALTER TABLE tasks ADD CONSTRAINT chk_task_progress 
CHECK (
    (status = 'pending' AND progress = 0) OR
    (status = 'running' AND progress BETWEEN 1 AND 99) OR
    (status = 'completed' AND progress = 100) OR
    (status IN ('failed', 'cancelled'))
);

-- 文档版本约束
ALTER TABLE documents ADD CONSTRAINT chk_document_version 
CHECK (version >= 1);

-- 用户权限约束
ALTER TABLE user_permissions ADD CONSTRAINT chk_permission_expiry 
CHECK (expires_at IS NULL OR expires_at > granted_at);
```

### 4.2 自动更新触发器
```sql
-- 更新时间自动更新触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表添加更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 项目完成时间自动更新触发器
CREATE OR REPLACE FUNCTION update_project_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        NEW.completed_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_projects_completed_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_project_completed_at();
```

## 5. 分区策略（可选）

### 5.1 按时间分区
```sql
-- 为大型表启用分区（可选，用于生产环境）
-- 任务日志表按月份分区
CREATE TABLE task_logs_partitioned (
    id UUID NOT NULL,
    task_id UUID NOT NULL,
    log_level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
) PARTITION BY RANGE (created_at);

-- 创建分区
CREATE TABLE task_logs_2024_01 PARTITION OF task_logs_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE task_logs_2024_02 PARTITION OF task_logs_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- 自动创建分区的函数
CREATE OR REPLACE FUNCTION create_monthly_partition(table_name text, start_date date)
RETURNS void AS $$
DECLARE
    partition_name text;
    end_date date;
BEGIN
    partition_name := table_name || '_' || to_char(start_date, 'YYYY_MM');
    end_date := start_date + interval '1 month';
    
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF %I
                    FOR VALUES FROM (%L) TO (%L)',
                    partition_name, table_name, start_date, end_date);
END;
$$ LANGUAGE plpgsql;
```

## 6. 数据迁移脚本

### 6.1 初始化脚本
```sql
-- 创建数据库扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建枚举类型
CREATE TYPE user_role AS ENUM ('user', 'admin', 'moderator');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned');
CREATE TYPE project_status AS ENUM ('draft', 'planning', 'development', 'testing', 'completed', 'deployed', 'archived');
CREATE TYPE project_type AS ENUM ('web', 'mobile', 'desktop', 'api');
CREATE TYPE task_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');
CREATE TYPE document_status AS ENUM ('draft', 'review', 'approved', 'published');

-- 插入初始数据
INSERT INTO users (id, email, password_hash, name, role, email_verified) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'admin@autocodeweb.com', 
     '$2a$10$hashedpassword', 'System Admin', 'admin', true);

-- 创建默认配置
INSERT INTO preview_configs (project_id, name, config_type, settings) VALUES
    (NULL, 'Default Web Config', 'web', '{"viewport": "1920x1080", "theme": "light"}'),
    (NULL, 'Default Mobile Config', 'mobile', '{"viewport": "375x667", "theme": "light"}');
```

### 6.2 升级脚本
```sql
-- 版本升级脚本示例
-- 从 v1.0 升级到 v1.1
DO $$ 
BEGIN
    -- 添加新列
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'projects' AND column_name = 'tech_stack') THEN
        ALTER TABLE projects ADD COLUMN tech_stack JSONB;
    END IF;
    
    -- 更新现有数据
    UPDATE projects SET tech_stack = '{}' WHERE tech_stack IS NULL;
    
    -- 添加新约束
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'chk_project_status_transition') THEN
        ALTER TABLE projects ADD CONSTRAINT chk_project_status_transition 
        CHECK (status IN ('draft', 'planning', 'development', 'testing', 'completed', 'deployed', 'archived'));
    END IF;
END $$;
```

## 7. 性能优化建议

### 7.1 查询优化
```sql
-- 常用查询优化
-- 1. 项目列表查询（带分页和过滤）
EXPLAIN ANALYZE
SELECT p.*, u.name as owner_name, 
       COUNT(t.id) as task_count,
       COUNT(CASE WHEN t.status = 'completed' THEN 1 END) as completed_tasks
FROM projects p
LEFT JOIN users u ON p.user_id = u.id
LEFT JOIN tasks t ON p.id = t.project_id
WHERE p.user_id = $1 OR p.id IN (
    SELECT project_id FROM project_members WHERE user_id = $1
)
GROUP BY p.id, u.name
ORDER BY p.updated_at DESC
LIMIT $2 OFFSET $3;

-- 2. Agent会话查询（带最新消息）
EXPLAIN ANALYZE
SELECT s.*, 
       (SELECT content FROM agent_messages 
        WHERE session_id = s.id 
        ORDER BY created_at DESC LIMIT 1) as last_message
FROM agent_sessions s
WHERE s.project_id = $1
ORDER BY s.started_at DESC;

-- 3. 任务统计查询
EXPLAIN ANALYZE
SELECT 
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN status = 'running' THEN 1 END) as running_tasks,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_tasks
FROM tasks 
WHERE project_id = $1;
```

### 7.2 缓存策略
```sql
-- Redis缓存键设计
-- 用户信息缓存
-- Key: user:{user_id}
-- TTL: 1小时

-- 项目信息缓存
-- Key: project:{project_id}
-- TTL: 30分钟

-- 项目任务统计缓存
-- Key: project_stats:{project_id}
-- TTL: 5分钟

-- Agent会话缓存
-- Key: agent_session:{session_id}
-- TTL: 15分钟

-- 热门项目缓存
-- Key: hot_projects
-- TTL: 10分钟
```

---

*本文档为 AutoCodeWeb 项目的数据模型设计，由架构师 Winston 创建*
