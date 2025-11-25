-- Add batch upload fields to transaction_data table
-- These fields support multiple file upload functionality without requiring upload sessions

ALTER TABLE transaction_data
ADD COLUMN session_code VARCHAR(50) NULL COMMENT 'Batch session code for multiple file uploads',
ADD COLUMN user_id INT NULL COMMENT 'User ID for batch uploads',
ADD COLUMN file_path VARCHAR(255) NULL COMMENT 'File path for batch uploads',
ADD COLUMN filename VARCHAR(255) NULL COMMENT 'Original filename for batch uploads';

-- Add indexes for performance
CREATE INDEX idx_transaction_data_session_code ON transaction_data(session_code);
CREATE INDEX idx_transaction_data_user_id ON transaction_data(user_id);