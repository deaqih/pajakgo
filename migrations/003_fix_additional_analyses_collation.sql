-- Fix collation issues and missing columns in additional_analyses table
-- Migration script to update additional_analyses table structure

-- Step 1: Add missing columns if they don't exist
ALTER TABLE additional_analyses
ADD COLUMN IF NOT EXISTS `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' COLLATE utf8mb4_unicode_ci,
ADD COLUMN IF NOT EXISTS `status` VARCHAR(20) NOT NULL DEFAULT 'active' COLLATE utf8mb4_unicode_ci,
ADD COLUMN IF NOT EXISTS `notes` TEXT NULL DEFAULT NULL COLLATE utf8mb4_unicode_ci,
ADD COLUMN IF NOT EXISTS `created_by` INT NULL DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Step 2: Fix collation for account_code column
ALTER TABLE additional_analyses
MODIFY COLUMN `account_code` VARCHAR(50) NOT NULL DEFAULT '' COLLATE utf8mb4_unicode_ci;

-- Step 3: Fix collation for other string columns
ALTER TABLE additional_analyses
MODIFY COLUMN `analysis_type` VARCHAR(100) NOT NULL COLLATE utf8mb4_unicode_ci,
MODIFY COLUMN `analysis_title` VARCHAR(255) NOT NULL COLLATE utf8mb4_unicode_ci,
MODIFY COLUMN `analysis_content` TEXT NOT NULL COLLATE utf8mb4_unicode_ci,
MODIFY COLUMN `category` VARCHAR(50) NOT NULL DEFAULT 'manual' COLLATE utf8mb4_unicode_ci;

-- Step 4: Update any existing data to have proper defaults
UPDATE additional_analyses
SET priority = 'medium',
    status = 'active',
    category = 'manual'
WHERE priority IS NULL OR status IS NULL OR category IS NULL;

-- Step 5: Add indexes for better performance
ALTER TABLE additional_analyses
ADD INDEX IF NOT EXISTS `idx_account_code` (`account_code`),
ADD INDEX IF NOT EXISTS `idx_analysis_type` (`analysis_type`),
ADD INDEX IF NOT EXISTS `idx_status` (`status`),
ADD INDEX IF NOT EXISTS `idx_category` (`category`),
ADD INDEX IF NOT EXISTS `idx_priority` (`priority`),
ADD INDEX IF NOT EXISTS `idx_created_at` (`created_at`);

-- Step 6: Add foreign key constraint (optional, if accounts table exists and has account_code as unique)
-- ALTER TABLE additional_analyses
-- ADD CONSTRAINT `fk_additional_analyses_account_code`
-- FOREIGN KEY (`account_code`) REFERENCES `accounts`(`account_code`)
-- ON DELETE CASCADE ON UPDATE CASCADE;