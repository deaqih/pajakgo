-- Rollback migration: Convert DECIMAL columns back to VARCHAR
-- Only use this if you need to rollback the migration

ALTER TABLE `transaction_data`
MODIFY COLUMN `um_pajak_db` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `pm_db` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `wth_21_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `wth_23_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `wth_26_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `wth_4_2_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `wth_15_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci',
MODIFY COLUMN `pk_cr` VARCHAR(100) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci';

-- Revert analisa_tambahan back to VARCHAR(255)
ALTER TABLE `transaction_data`
MODIFY COLUMN `analisa_tambahan` VARCHAR(255) NULL DEFAULT '' COLLATE 'utf8mb4_unicode_ci';

-- Convert numeric values back to strings
UPDATE `transaction_data`
SET
    `um_pajak_db` = CONCAT(`um_pajak_db`) WHERE `um_pajak_db` IS NOT NULL,
    `pm_db` = CONCAT(`pm_db`) WHERE `pm_db` IS NOT NULL,
    `wth_21_cr` = CONCAT(`wth_21_cr`) WHERE `wth_21_cr` IS NOT NULL,
    `wth_23_cr` = CONCAT(`wth_23_cr`) WHERE `wth_23_cr` IS NOT NULL,
    `wth_26_cr` = CONCAT(`wth_26_cr`) WHERE `wth_26_cr` IS NOT NULL,
    `wth_4_2_cr` = CONCAT(`wth_4_2_cr`) WHERE `wth_4_2_cr` IS NOT NULL,
    `wth_15_cr` = CONCAT(`wth_15_cr`) WHERE `wth_15_cr` IS NOT NULL,
    `pk_cr` = CONCAT(`pk_cr`) WHERE `pk_cr` IS NOT NULL;

-- Set empty strings for NULL values
UPDATE `transaction_data`
SET
    `um_pajak_db` = '' WHERE `um_pajak_db` IS NULL,
    `pm_db` = '' WHERE `pm_db` IS NULL,
    `wth_21_cr` = '' WHERE `wth_21_cr` IS NULL,
    `wth_23_cr` = '' WHERE `wth_23_cr` IS NULL,
    `wth_26_cr` = '' WHERE `wth_26_cr` IS NULL,
    `wth_4_2_cr` = '' WHERE `wth_4_2_cr` IS NULL,
    `wth_15_cr` = '' WHERE `wth_15_cr` IS NULL,
    `pk_cr` = '' WHERE `pk_cr` IS NULL,
    `analisa_tambahan` = '' WHERE `analisa_tambahan` IS NULL;