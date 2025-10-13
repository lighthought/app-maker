-- Migration Script: Add API Token Fields
-- Date: 2025-01-12
-- Description: Adds API token fields to users and projects tables

\c autocodeweb;

-- ============================================================================
-- Add API token field to users table
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'default_api_token'
    ) THEN
        ALTER TABLE users ADD COLUMN default_api_token VARCHAR(500);
        RAISE NOTICE 'Added default_api_token column to users table';
    END IF;
END $$;

-- ============================================================================
-- Add API token field to projects table
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'projects' AND column_name = 'api_token'
    ) THEN
        ALTER TABLE projects ADD COLUMN api_token VARCHAR(500);
        RAISE NOTICE 'Added api_token column to projects table';
    END IF;
END $$;

-- ============================================================================
-- Create indexes for new fields
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_users_default_api_token ON users(default_api_token);
CREATE INDEX IF NOT EXISTS idx_projects_api_token ON projects(api_token);

\echo ''
\echo '=========================================='
\echo 'Migration completed successfully!'
\echo '=========================================='
\echo 'Added fields:'
\echo '  - users: default_api_token'
\echo '  - projects: api_token'
\echo '=========================================='

