# Database Migration Instructions

## Migration 001: Update Transaction Data Types

This migration updates the `transaction_data` table to use proper data types for tax calculation fields.

### Changes:
- **VARCHAR → DECIMAL(15,2)**: For all tax calculation columns
- **VARCHAR(255) → TEXT**: For `analisa_tambahan` field
- **Data cleanup**: Converts empty strings to proper NULL/0 values

### Migration Files:
- `001_update_transaction_data_types.sql` - Main migration
- `001_update_transaction_data_types_rollback.sql` - Rollback script

### Columns Updated:
- `um_pajak_db` (Uang Muka Pajak Debet)
- `pm_db` (Pajak Masukan Debet)
- `wth_21_cr` (Withholding Tax 21 Credit)
- `wth_23_cr` (Withholding Tax 23 Credit)
- `wth_26_cr` (Withholding Tax 26 Credit)
- `wth_4_2_cr` (Withholding Tax 4.2 Credit)
- `wth_15_cr` (Withholding Tax 15 Credit)
- `pk_cr` (Pajak Kredit Credit)
- `analisa_tambahan` (Additional Analysis)

### How to Run Migration:

1. **Backup your database first:**
   ```sql
   mysqldump -u username -p accounting_web > backup_before_migration.sql
   ```

2. **Run the main migration:**
   ```sql
   mysql -u username -p accounting_web < migrations/001_update_transaction_data_types.sql
   ```

3. **Verify the changes:**
   ```sql
   DESCRIBE transaction_data;
   SELECT COUNT(*) FROM transaction_data WHERE um_pajak_db IS NULL;
   ```

### Rollback (if needed):

If you need to rollback the changes:

```sql
mysql -u username -p accounting_web < migrations/001_update_transaction_data_types_rollback.sql
```

### Benefits:

- ✅ **Proper Numeric Calculations**: DECIMAL type allows accurate tax calculations
- ✅ **Better Performance**: Numeric operations are faster with proper data types
- ✅ **Data Integrity**: Prevents invalid text values in numeric fields
- ✅ **Consistent Formatting**: Standardized decimal places for all tax fields

### Important Notes:

- Migration will preserve existing data and convert string values to numeric
- Empty strings ('') will be converted to 0.00
- Invalid numeric strings will be converted to NULL then to 0.00
- All existing application code using these fields will continue to work
- The Go models already use `*float64` which is compatible with DECIMAL