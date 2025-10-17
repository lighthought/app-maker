-- 添加 auto_go_next 和用户确认相关字段的迁移脚本
-- 执行时间: 2024-12-19

-- 1. 为用户表添加 auto_go_next 字段
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS auto_go_next BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN users.auto_go_next IS '用户全局自动进入下一阶段配置，true表示跳过所有确认步骤';

-- 2. 为项目表添加用户确认相关字段
ALTER TABLE projects 
ADD COLUMN IF NOT EXISTS waiting_for_user_confirm BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS confirm_stage VARCHAR(50) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS auto_go_next BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN projects.waiting_for_user_confirm IS '是否等待用户确认，true表示当前阶段需要用户确认';
COMMENT ON COLUMN projects.confirm_stage IS '等待用户确认的阶段名称';
COMMENT ON COLUMN projects.auto_go_next IS '项目级自动进入下一阶段配置，覆盖用户全局设置';

-- 3. 为 Epic 表添加排序字段
ALTER TABLE project_epics 
ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN project_epics.display_order IS 'Epic 显示顺序，用于前端拖拽排序';

-- 4. 为 Story 表添加排序字段
ALTER TABLE epic_stories 
ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN epic_stories.display_order IS 'Story 显示顺序，用于前端拖拽排序';

-- 5. 为现有数据设置默认排序值
-- 为现有的 Epics 设置排序值
UPDATE project_epics 
SET display_order = epic_number 
WHERE display_order = 0;

-- 为现有的 Stories 设置排序值（基于 story_number 的数值部分）
UPDATE epic_stories 
SET display_order = CAST(
    CASE 
        WHEN story_number ~ '^[0-9]+$' THEN story_number
        WHEN story_number ~ '^[0-9]+\.[0-9]+$' THEN SPLIT_PART(story_number, '.', 1)
        ELSE '0'
    END AS INTEGER
) 
WHERE display_order = 0;

-- 6. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_projects_waiting_confirm ON projects(waiting_for_user_confirm);
CREATE INDEX IF NOT EXISTS idx_projects_confirm_stage ON projects(confirm_stage);
CREATE INDEX IF NOT EXISTS idx_project_epics_display_order ON project_epics(project_guid, display_order);
CREATE INDEX IF NOT EXISTS idx_epic_stories_display_order ON epic_stories(epic_id, display_order);

-- 7. 添加触发器以自动更新 updated_at 字段
CREATE TRIGGER update_users_auto_go_next_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_confirm_fields_updated_at 
    BEFORE UPDATE ON projects 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_project_epics_display_order_updated_at 
    BEFORE UPDATE ON project_epics 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_epic_stories_display_order_updated_at 
    BEFORE UPDATE ON epic_stories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 迁移完成
SELECT 'Migration completed: Added auto_go_next and user confirmation fields' AS result;
