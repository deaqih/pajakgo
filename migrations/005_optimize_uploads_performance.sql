-- Add indexes for ultra-fast upload_sessions performance
-- These indexes ensure the /uploads page loads in milliseconds even with millions of records

-- Create index for user-specific queries (already exists but ensuring)
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_id ON upload_sessions(user_id);

-- Create composite index for user pagination queries (most common pattern)
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_created_at ON upload_sessions(user_id, created_at DESC);

-- Create index for status filtering (useful for admin dashboards)
CREATE INDEX IF NOT EXISTS idx_upload_sessions_status ON upload_sessions(status);

-- Create composite index for admin queries with status filter
CREATE INDEX IF NOT EXISTS idx_upload_sessions_status_created_at ON upload_sessions(status, created_at DESC);

-- Ensure the created_at index exists for ORDER BY performance
CREATE INDEX IF NOT EXISTS idx_upload_sessions_created_at ON upload_sessions(created_at DESC);