-- Check session codes for recent uploads
SELECT
    session_code,
    filename,
    COUNT(*) as row_count,
    MIN(id) as first_id,
    MAX(id) as last_id
FROM transaction_data
WHERE session_code LIKE 'BATCH-%'
GROUP BY session_code, filename
ORDER BY session_code, filename;

-- Check unique session codes in recent data
SELECT DISTINCT session_code, COUNT(*) as total_rows
FROM transaction_data
WHERE session_code LIKE 'BATCH-%'
GROUP BY session_code;