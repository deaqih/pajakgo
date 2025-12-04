-- Script to test cancel functionality
-- Update session 3 to "canceled" status

UPDATE upload_sessions
SET status = 'canceled',
    error_message = 'User canceled processing',
    updated_at = NOW()
WHERE id = 3;

-- Check the updated session
SELECT * FROM upload_sessions WHERE id = 3;