-- Add name column to users table if it doesn't exist

-- Check if column exists before adding it
-- For MySQL 5.7+
ALTER TABLE users
ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '' AFTER id;

-- If the column was just added, update existing users with their username as default name
UPDATE users SET name = username WHERE name = '';
