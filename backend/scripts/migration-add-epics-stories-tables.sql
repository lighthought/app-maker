-- Migration: Add project_epics and epic_stories tables
-- Date: 2025-10-16
-- Description: 添加 Epic 和 Story 管理表，支持 MVP Stories 开发功能

-- 创建 project_epics 表的 ID 序列
CREATE SEQUENCE IF NOT EXISTS project_epics_id_num_seq START 1;

-- Epic 表
CREATE TABLE IF NOT EXISTS project_epics (
    id VARCHAR(50) PRIMARY KEY DEFAULT generate_table_id('EPIC', 'project_epics_id_num_seq'),
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    project_guid VARCHAR(50) NOT NULL,
    epic_number INT NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    priority VARCHAR(20) NOT NULL,
    estimated_days INT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_path VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(project_id, epic_number)
);

-- 创建 epic_stories 表的 ID 序列
CREATE SEQUENCE IF NOT EXISTS epic_stories_id_num_seq START 1;

-- Story 表
CREATE TABLE IF NOT EXISTS epic_stories (
    id VARCHAR(50) PRIMARY KEY DEFAULT generate_table_id('STORY', 'epic_stories_id_num_seq'),
    epic_id VARCHAR(50) NOT NULL REFERENCES project_epics(id) ON DELETE CASCADE,
    story_number VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    priority VARCHAR(20) NOT NULL,
    estimated_days INT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_path VARCHAR(500),
    depends TEXT,
    techs TEXT,
    content TEXT,
    acceptance_criteria TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(epic_id, story_number)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_project_epics_project_id ON project_epics(project_id);
CREATE INDEX IF NOT EXISTS idx_project_epics_project_guid ON project_epics(project_guid);
CREATE INDEX IF NOT EXISTS idx_project_epics_status ON project_epics(status);
CREATE INDEX IF NOT EXISTS idx_epic_stories_epic_id ON epic_stories(epic_id);
CREATE INDEX IF NOT EXISTS idx_epic_stories_status ON epic_stories(status);

-- 添加注释
COMMENT ON TABLE project_epics IS '项目 Epic（史诗）表';
COMMENT ON TABLE epic_stories IS 'Epic Story（用户故事）表';
COMMENT ON COLUMN project_epics.epic_number IS 'Epic 编号，从 1 开始';
COMMENT ON COLUMN project_epics.priority IS '优先级: P0, P1, P2';
COMMENT ON COLUMN project_epics.status IS '状态: pending, in_progress, done, failed';
COMMENT ON COLUMN epic_stories.story_number IS 'Story 编号，如 US-001';
COMMENT ON COLUMN epic_stories.depends IS '依赖的其他 Story';
COMMENT ON COLUMN epic_stories.techs IS '技术要点';
COMMENT ON COLUMN epic_stories.content IS 'Story 的完整内容';

