-- Add indexes for session_code optimization
-- This will improve query performance when using session_code-based relationships

-- Add index on transaction_data.session_code for faster lookups
CREATE INDEX IF NOT EXISTS idx_transaction_data_session_code ON transaction_data(session_code);

-- Add composite index on transaction_data(session_code, is_processed) for common queries
CREATE INDEX IF NOT EXISTS idx_transaction_data_session_code_processed ON transaction_data(session_code, is_processed);

-- Add index on upload_sessions.session_code if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_upload_sessions_session_code ON upload_sessions(session_code);

-- Add composite index on upload_sessions(user_id, session_code) for user-specific queries
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_session_code ON upload_sessions(user_id, session_code);

-- Add composite index on upload_sessions(user_id, created_at) for pagination
CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_created_at ON upload_sessions(user_id, created_at);