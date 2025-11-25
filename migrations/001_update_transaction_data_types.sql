-- Migration to update transaction_data table column types
-- Change VARCHAR columns to DECIMAL for proper numeric calculations

ALTER TABLE `transaction_data`
MODIFY COLUMN `um_pajak_db` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Uang Muka Pajak Debet',
MODIFY COLUMN `pm_db` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Pajak Masukan Debet',
MODIFY COLUMN `wth_21_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Withholding Tax 21 Credit',
MODIFY COLUMN `wth_23_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Withholding Tax 23 Credit',
MODIFY COLUMN `wth_26_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Withholding Tax 26 Credit',
MODIFY COLUMN `wth_4_2_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Withholding Tax 4.2 Credit',
MODIFY COLUMN `wth_15_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Withholding Tax 15 Credit',
MODIFY COLUMN `pk_cr` DECIMAL(15,2) NULL DEFAULT 0.00 COMMENT 'Pajak Kredit Credit';

-- Update analisa_tambahan to proper TEXT type for longer descriptions
ALTER TABLE `transaction_data`
MODIFY COLUMN `analisa_tambahan` TEXT NULL COMMENT 'Additional Analysis';

-- Convert existing empty strings to NULL for proper DECIMAL handling
UPDATE `transaction_data`
SET
    `um_pajak_db` = NULL WHERE `um_pajak_db` = '',
    `pm_db` = NULL WHERE `pm_db` = '',
    `wth_21_cr` = NULL WHERE `wth_21_cr` = '',
    `wth_23_cr` = NULL WHERE `wth_23_cr` = '',
    `wth_26_cr` = NULL WHERE `wth_26_cr` = '',
    `wth_4_2_cr` = NULL WHERE `wth_4_2_cr` = '',
    `wth_15_cr` = NULL WHERE `wth_15_cr` = '',
    `pk_cr` = NULL WHERE `pk_cr` = '';

-- Update remaining NULL values to 0.00 for DECIMAL columns
UPDATE `transaction_data`
SET
    `um_pajak_db` = 0.00 WHERE `um_pajak_db` IS NULL,
    `pm_db` = 0.00 WHERE `pm_db` IS NULL,
    `wth_21_cr` = 0.00 WHERE `wth_21_cr` IS NULL,
    `wth_23_cr` = 0.00 WHERE `wth_23_cr` IS NULL,
    `wth_26_cr` = 0.00 WHERE `wth_26_cr` IS NULL,
    `wth_4_2_cr` = 0.00 WHERE `wth_4_2_cr` IS NULL,
    `wth_15_cr` = 0.00 WHERE `wth_15_cr` IS NULL,
    `pk_cr` = 0.00 WHERE `pk_cr` IS NULL;