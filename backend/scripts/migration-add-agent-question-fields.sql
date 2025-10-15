-- 迁移脚本：添加 Agent 与用户交互功能的数据库支持
-- 创建时间: 2025-10-15
-- 说明: 
--   1. 为 project_msgs 表添加 has_question 和 waiting_user_response 字段
--   2. 为 projects 和 dev_stages 表的 status 字段添加 'paused' 状态支持

-- 连接到数据库
\c autocodeweb;

-- ========================================
-- 第一部分：修改 project_msgs 表
-- ========================================

-- 添加 has_question 字段（是否包含问题）
ALTER TABLE project_msgs 
ADD COLUMN IF NOT EXISTS has_question BOOLEAN DEFAULT FALSE;

-- 添加 waiting_user_response 字段（是否等待用户回复）
ALTER TABLE project_msgs 
ADD COLUMN IF NOT EXISTS waiting_user_response BOOLEAN DEFAULT FALSE;

-- 添加索引以优化查询性能
CREATE INDEX IF NOT EXISTS idx_project_msgs_has_question 
ON project_msgs(has_question) 
WHERE has_question = TRUE;

CREATE INDEX IF NOT EXISTS idx_project_msgs_waiting_user_response 
ON project_msgs(waiting_user_response) 
WHERE waiting_user_response = TRUE;

-- 添加注释
COMMENT ON COLUMN project_msgs.has_question IS '消息是否包含问题（需要用户回答）';
COMMENT ON COLUMN project_msgs.waiting_user_response IS '是否正在等待用户回复';

-- ========================================
-- 第二部分：修改 projects 表的 status 约束
-- ========================================

-- 删除旧的 status 约束
ALTER TABLE projects 
DROP CONSTRAINT IF EXISTS projects_status_check;

-- 添加新的 status 约束，包含 'paused' 状态
ALTER TABLE projects 
ADD CONSTRAINT projects_status_check 
CHECK (status IN ('pending', 'in_progress', 'done', 'failed', 'paused'));

-- ========================================
-- 第三部分：修改 dev_stages 表的 status 约束
-- ========================================

-- 删除旧的 status 约束
ALTER TABLE dev_stages 
DROP CONSTRAINT IF EXISTS dev_stages_status_check;

-- 添加新的 status 约束，包含 'paused' 状态
ALTER TABLE dev_stages 
ADD CONSTRAINT dev_stages_status_check 
CHECK (status IN ('pending', 'in_progress', 'done', 'failed', 'paused'));

-- ========================================
-- 显示修改后的表结构
-- ========================================

\d project_msgs
\d projects
\d dev_stages

-- 完成提示
SELECT 'Migration completed successfully!' AS status;
SELECT '  - Added has_question and waiting_user_response fields to project_msgs table' AS detail;
SELECT '  - Added paused status support to projects and dev_stages tables' AS detail;

