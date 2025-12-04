-- Test query untuk koreksi rules
SELECT id, keyword, value, not_value, is_active, created_at, updated_at
FROM pajak_db.koreksi_rules
LIMIT 5;