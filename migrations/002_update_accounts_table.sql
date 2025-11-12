-- Update accounts table structure

-- Add new columns if they don't exist
ALTER TABLE accounts
ADD COLUMN IF NOT EXISTS koreksi_obyek VARCHAR(50) NULL DEFAULT NULL AFTER nature;

ALTER TABLE accounts
ADD COLUMN IF NOT EXISTS analisa_tambahan VARCHAR(50) NULL DEFAULT NULL AFTER koreksi_obyek;

-- Rename account_number to account_code if needed (if table was created with old structure)
-- ALTER TABLE accounts CHANGE COLUMN account_number account_code VARCHAR(50) NOT NULL;

-- Ensure all required columns exist with correct types
ALTER TABLE accounts MODIFY COLUMN account_code VARCHAR(50) NOT NULL;
ALTER TABLE accounts MODIFY COLUMN account_name VARCHAR(255) NOT NULL;
ALTER TABLE accounts MODIFY COLUMN account_type VARCHAR(100) NULL DEFAULT NULL;
ALTER TABLE accounts MODIFY COLUMN nature VARCHAR(50) NULL DEFAULT NULL COMMENT 'Asset, Liability, Equity, Revenue, Expense';
ALTER TABLE accounts MODIFY COLUMN koreksi_obyek VARCHAR(50) NULL DEFAULT NULL;
ALTER TABLE accounts MODIFY COLUMN analisa_tambahan VARCHAR(50) NULL DEFAULT NULL;
ALTER TABLE accounts MODIFY COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1;

-- Add unique index on account_code if not exists
ALTER TABLE accounts ADD UNIQUE INDEX IF NOT EXISTS idx_account_code (account_code);

-- Add index on account_name for faster searches
ALTER TABLE accounts ADD INDEX IF NOT EXISTS idx_account_name (account_name);
