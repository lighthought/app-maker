-- Migration Script: Add CLI Configuration and Preview Token Support
-- Date: 2025-01-12
-- Description: Adds CLI tool and model configuration fields to users and projects tables,
--              and creates preview_tokens table for shareable preview links

\c autocodeweb;

-- ============================================================================
-- 1. Add CLI/Model configuration fields to users table
-- ============================================================================

DO $$
BEGIN
    -- Add default_cli_tool field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'default_cli_tool'
    ) THEN
        ALTER TABLE users ADD COLUMN default_cli_tool VARCHAR(50) DEFAULT 'claude-code';
        RAISE NOTICE 'Added default_cli_tool column to users table';
    END IF;

    -- Add default_ai_model field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'default_ai_model'
    ) THEN
        ALTER TABLE users ADD COLUMN default_ai_model VARCHAR(100) DEFAULT 'glm-4.6';
        RAISE NOTICE 'Added default_ai_model column to users table';
    END IF;

    -- Add default_model_provider field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'default_model_provider'
    ) THEN
        ALTER TABLE users ADD COLUMN default_model_provider VARCHAR(50) DEFAULT 'zhipu';
        RAISE NOTICE 'Added default_model_provider column to users table';
    END IF;

    -- Add default_model_api_url field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'default_model_api_url'
    ) THEN
        ALTER TABLE users ADD COLUMN default_model_api_url VARCHAR(500);
        RAISE NOTICE 'Added default_model_api_url column to users table';
    END IF;
END $$;

-- ============================================================================
-- 2. Add CLI/Model configuration fields to projects table
-- ============================================================================

DO $$
BEGIN
    -- Add cli_tool field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' AND column_name = 'cli_tool'
    ) THEN
        ALTER TABLE projects ADD COLUMN cli_tool VARCHAR(50);
        RAISE NOTICE 'Added cli_tool column to projects table';
    END IF;

    -- Add ai_model field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' AND column_name = 'ai_model'
    ) THEN
        ALTER TABLE projects ADD COLUMN ai_model VARCHAR(100);
        RAISE NOTICE 'Added ai_model column to projects table';
    END IF;

    -- Add model_provider field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' AND column_name = 'model_provider'
    ) THEN
        ALTER TABLE projects ADD COLUMN model_provider VARCHAR(50);
        RAISE NOTICE 'Added model_provider column to projects table';
    END IF;

    -- Add model_api_url field
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' AND column_name = 'model_api_url'
    ) THEN
        ALTER TABLE projects ADD COLUMN model_api_url VARCHAR(500);
        RAISE NOTICE 'Added model_api_url column to projects table';
    END IF;
END $$;

-- ============================================================================
-- 3. Create preview_tokens table and sequence
-- ============================================================================

-- Create preview tokens ID sequence
CREATE SEQUENCE IF NOT EXISTS public.preview_tokens_id_num_seq
    INCREMENT BY 1
    START 1
    MINVALUE 1
    MAXVALUE 99999999999
    CACHE 1;

-- Create preview_tokens table
CREATE TABLE IF NOT EXISTS preview_tokens (
    id VARCHAR(50) PRIMARY KEY DEFAULT public.generate_table_id('PREV', 'public.preview_tokens_id_num_seq'),
    token VARCHAR(255) UNIQUE NOT NULL,
    project_id VARCHAR(50) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ============================================================================
-- 4. Create indexes for new fields
-- ============================================================================

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_default_cli_tool ON users(default_cli_tool);
CREATE INDEX IF NOT EXISTS idx_users_default_model_provider ON users(default_model_provider);

-- Indexes for projects table
CREATE INDEX IF NOT EXISTS idx_projects_cli_tool ON projects(cli_tool);
CREATE INDEX IF NOT EXISTS idx_projects_model_provider ON projects(model_provider);

-- Indexes for preview_tokens table
CREATE INDEX IF NOT EXISTS idx_preview_tokens_token ON preview_tokens(token);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_project_id ON preview_tokens(project_id);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_expires_at ON preview_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_preview_tokens_created_at ON preview_tokens(created_at);

-- ============================================================================
-- 5. Add trigger for preview_tokens table
-- ============================================================================

CREATE TRIGGER update_preview_tokens_updated_at 
    BEFORE UPDATE ON preview_tokens 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 6. Update existing users with default values
-- ============================================================================

-- Set default values for existing users (if any have NULL values)
UPDATE users 
SET 
    default_cli_tool = 'claude-code',
    default_ai_model = 'glm-4.6',
    default_model_provider = 'zhipu',
    default_model_api_url = 'https://open.bigmodel.cn/api/anthropic'
WHERE 
    default_cli_tool IS NULL 
    OR default_ai_model IS NULL 
    OR default_model_provider IS NULL;

-- ============================================================================
-- 7. Verification - Display updated schema
-- ============================================================================

-- Show users table columns
\echo ''
\echo '=========================================='
\echo 'Users Table Columns:'
\echo '=========================================='
SELECT 
    column_name, 
    data_type, 
    character_maximum_length,
    column_default
FROM information_schema.columns 
WHERE table_name = 'users' 
    AND column_name IN ('default_cli_tool', 'default_ai_model', 'default_model_provider', 'default_model_api_url')
ORDER BY ordinal_position;

-- Show projects table columns
\echo ''
\echo '=========================================='
\echo 'Projects Table Columns:'
\echo '=========================================='
SELECT 
    column_name, 
    data_type, 
    character_maximum_length,
    column_default
FROM information_schema.columns 
WHERE table_name = 'projects' 
    AND column_name IN ('cli_tool', 'ai_model', 'model_provider', 'model_api_url')
ORDER BY ordinal_position;

-- Show preview_tokens table structure
\echo ''
\echo '=========================================='
\echo 'Preview Tokens Table:'
\echo '=========================================='
\d preview_tokens

-- Show all tables
\echo ''
\echo '=========================================='
\echo 'All Tables:'
\echo '=========================================='
\dt

-- Show migration completion message
\echo ''
\echo '=========================================='
\echo 'Migration completed successfully!'
\echo '=========================================='
\echo 'Added fields:'
\echo '  - users: default_cli_tool, default_ai_model, default_model_provider, default_model_api_url'
\echo '  - projects: cli_tool, ai_model, model_provider, model_api_url'
\echo '  - preview_tokens: New table created'
\echo '=========================================='

