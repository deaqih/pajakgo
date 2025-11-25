-- Add not_value column to koreksi_rules table
ALTER TABLE koreksi_rules
ADD COLUMN not_value VARCHAR(255) DEFAULT NULL AFTER value;

-- Update existing records with NULL not_value
-- No action needed as DEFAULT NULL is specified