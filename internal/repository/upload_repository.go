package repository

import (
	"accounting-web/internal/models"
	"accounting-web/internal/utils"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type UploadRepository struct {
	db *sqlx.DB
}

func NewUploadRepository(db *sqlx.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// Upload Sessions
func (r *UploadRepository) CreateSession(session *models.UploadSession) error {
	query := `INSERT INTO upload_sessions (session_code, user_id, filename, file_path,
	          total_rows, status) VALUES (:session_code, :user_id, :filename, :file_path,
	          :total_rows, :status)`
	result, err := r.db.NamedExec(query, session)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	session.ID = int(id)
	return nil
}

func (r *UploadRepository) GetSessionByID(id int) (*models.UploadSession, error) {
	var session models.UploadSession
	query := "SELECT * FROM upload_sessions WHERE id = ? LIMIT 1"
	err := r.db.Get(&session, query, id)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UploadRepository) GetSessionByCode(code string) (*models.UploadSession, error) {
	var session models.UploadSession
	query := "SELECT * FROM upload_sessions WHERE session_code = ? LIMIT 1"
	err := r.db.Get(&session, query, code)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *UploadRepository) GetSessions(limit, offset int, userID int) ([]models.UploadSession, int, error) {
	var sessions []models.UploadSession
	var total int

	whereClause := ""
	args := []interface{}{}

	// if userID > 0 {
	// 	whereClause = "WHERE user_id = ?"
	// 	args = append(args, userID)
	// }

	countQuery := "SELECT COUNT(*) FROM upload_sessions " + whereClause
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT * FROM upload_sessions " + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err = r.db.Select(&sessions, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// GetSessionsOptimized - Ultra-fast query that only touches upload_sessions table
// No JOINs with transaction_data for maximum performance
func (r *UploadRepository) GetSessionsOptimized(limit, offset int, userID int) ([]models.UploadSession, int, error) {
	var sessions []models.UploadSession
	var total int

	whereClause := ""
	args := []interface{}{}

	// if userID > 0 {
	// 	whereClause = "WHERE user_id = ?"
	// 	args = append(args, userID)
	// }

	// Ultra-fast count query - only scans upload_sessions primary key/index
	countQuery := "SELECT COUNT(*) FROM upload_sessions " + whereClause
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Ultra-fast data query - only selects from upload_sessions, uses created_at index for ORDER BY
	query := "SELECT id, session_code, user_id, filename, file_path, total_rows, processed_rows, failed_rows, status, error_message, created_at, updated_at FROM upload_sessions " + whereClause + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err = r.db.Select(&sessions, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// GetSessionsBySessionCode gets sessions using session_code-based relationship
func (r *UploadRepository) GetSessionsBySessionCode(limit, offset int, userID int) ([]map[string]interface{}, int, error) {
	var sessions []map[string]interface{}
	var total int

	// Get upload sessions first
	sessionsQuery := `
		SELECT
			us.id,
			us.session_code,
			us.user_id,
			us.filename,
			us.file_path,
			us.total_rows,
			us.processed_rows,
			us.failed_rows,
			us.status,
			us.error_message,
			us.created_at,
			us.updated_at,
			COUNT(td.id) as transaction_count
		FROM upload_sessions us
		LEFT JOIN transaction_data td ON us.session_code = td.session_code
	`

	whereClause := ""
	args := []interface{}{}

	// if userID > 0 {
	// 	whereClause += " WHERE us.user_id = ?"
	// 	args = append(args, userID)
	// }

	groupClause := " GROUP BY us.id ORDER BY us.created_at DESC"
	limitClause := " LIMIT ? OFFSET ?"

	// Count query
	countQuery := "SELECT COUNT(DISTINCT us.id) FROM upload_sessions us LEFT JOIN transaction_data td ON us.session_code = td.session_code" + whereClause
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	args = append(args, limit, offset)
	finalQuery := sessionsQuery + whereClause + groupClause + limitClause
	err = r.db.Select(&sessions, finalQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// GetSessionDetailBySessionCode gets detailed session info using session_code
func (r *UploadRepository) GetSessionDetailBySessionCode(sessionCode string) (*map[string]interface{}, error) {
	// First get the basic session info
	var session models.UploadSession
	query := `SELECT id, session_code, user_id, filename, file_path, total_rows, processed_rows, failed_rows, status, error_message, created_at, updated_at FROM upload_sessions WHERE session_code = ?`

	err := r.db.Get(&session, query, sessionCode)
	if err != nil {
		return nil, err
	}

	// Get transaction statistics separately
	var stats struct {
		TransactionCount int `db:"transaction_count"`
		ProcessedCount   int `db:"processed_count"`
		PendingCount     int `db:"pending_count"`
	}

	statsQuery := `
		SELECT
			COUNT(id) as transaction_count,
			SUM(CASE WHEN is_processed = 1 THEN 1 ELSE 0 END) as processed_count,
			SUM(CASE WHEN is_processed = 0 THEN 1 ELSE 0 END) as pending_count
		FROM transaction_data
		WHERE session_code = ?
	`

	err = r.db.Get(&stats, statsQuery, sessionCode)
	if err != nil {
		// If stats query fails, continue with zero values
		stats.TransactionCount = 0
		stats.ProcessedCount = 0
		stats.PendingCount = 0
	}

	// Build response map
	sessionMap := map[string]interface{}{
		"id":                session.ID,
		"session_code":      session.SessionCode,
		"user_id":           session.UserID,
		"filename":          session.Filename,
		"file_path":         session.FilePath,
		"total_rows":        session.TotalRows,
		"processed_rows":    session.ProcessedRows,
		"failed_rows":       session.FailedRows,
		"status":            session.Status,
		"error_message":     session.ErrorMessage,
		"created_at":        session.CreatedAt,
		"updated_at":        session.UpdatedAt,
		"transaction_count": stats.TransactionCount,
		"processed_count":   stats.ProcessedCount,
		"pending_count":     stats.PendingCount,
	}

	return &sessionMap, nil
}

// GetBatchUploads gets batch uploads from transaction_data that don't have matching upload_sessions
func (r *UploadRepository) GetBatchUploads(limit, offset int, userID int) ([]models.BatchUploadSession, int, error) {
	var batches []models.BatchUploadSession
	var total int

	whereClause := "WHERE t.session_id = 0 AND t.session_code IS NOT NULL"
	args := []interface{}{}

	// if userID > 0 {
	// 	whereClause += " AND t.user_id = ?"
	// 	args = append(args, userID)
	// }

	// Count unique session codes
	countQuery := `
		SELECT COUNT(DISTINCT t.session_code)
		FROM transaction_data t
		LEFT JOIN upload_sessions s ON t.session_id = s.id
		` + whereClause

	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get batch upload summary
	query := `
		SELECT
			t.session_code,
			t.user_id,
			MIN(t.filename) as filename,
			COUNT(*) as total_rows,
			SUM(CASE WHEN t.is_processed = 1 THEN 1 ELSE 0 END) as processed_rows,
			SUM(CASE WHEN t.is_processed = 0 THEN 1 ELSE 0 END) as failed_rows,
			CASE
				WHEN SUM(CASE WHEN t.is_processed = 1 THEN 1 ELSE 0 END) = COUNT(*) THEN 'completed'
				WHEN SUM(CASE WHEN t.is_processed = 1 THEN 1 ELSE 0 END) > 0 THEN 'processing'
				ELSE 'uploaded'
			END as status,
			MIN(t.created_at) as created_at,
			MAX(t.updated_at) as updated_at
		FROM transaction_data t
		LEFT JOIN upload_sessions s ON t.session_id = s.id
		` + whereClause + `
		GROUP BY t.session_code, t.user_id
		ORDER BY MIN(t.created_at) DESC
		LIMIT ? OFFSET ?`

	args = append(args, limit, offset)

	fmt.Printf("DEBUG: Executing GetBatchUploads query for userID: %d, limit: %d, offset: %d\n", userID, limit, offset)
	fmt.Printf("DEBUG: Total count before: %d\n", total)

	err = r.db.Select(&batches, query, args...)
	if err != nil {
		fmt.Printf("DEBUG: GetBatchUploads query error: %v\n", err)
		fmt.Printf("DEBUG: Query: %s\n", query)
		fmt.Printf("DEBUG: Args: %v\n", args)
		return nil, 0, err
	}

	fmt.Printf("DEBUG: GetBatchUploads success, found %d batches\n", len(batches))
	return batches, total, nil
}

func (r *UploadRepository) UpdateSession(session *models.UploadSession) error {
	query := `UPDATE upload_sessions SET processed_rows = :processed_rows,
	          failed_rows = :failed_rows, status = :status, error_message = :error_message
	          WHERE id = :id`
	_, err := r.db.NamedExec(query, session)
	return err
}

func (r *UploadRepository) UpdateSessionStatus(id int, status string) error {
	query := "UPDATE upload_sessions SET status = ? WHERE id = ?"
	_, err := r.db.Exec(query, status, id)
	return err
}

// Transaction Data - Optimized for session_code only
func (r *UploadRepository) CreateMultipleTransactions(transactions []models.TransactionData) error {
	if len(transactions) == 0 {
		return nil
	}

	fmt.Printf("DEBUG: CreateMultipleTransactions: Processing %d transactions with session_code optimization\n", len(transactions))

	// Define chunk size to avoid MySQL placeholder limit (65535)
	// Each transaction has about 10 placeholders, so we use 5000 transactions per chunk
	const CHUNK_SIZE = 5000

	totalInserted := int64(0)
	sessionCode := transactions[0].SessionCode // All transactions should have same session_code

	for i := 0; i < len(transactions); i += CHUNK_SIZE {
		end := i + CHUNK_SIZE
		if end > len(transactions) {
			end = len(transactions)
		}

		chunk := transactions[i:end]

		fmt.Printf("DEBUG: Processing chunk %d-%d of %d total transactions for session_code: %s\n",
			i+1, end, len(transactions), sessionCode)

		// Optimized query - session_id = 0, rely on session_code for relation
		query := `INSERT INTO transaction_data (session_id, session_code, user_id, file_path, filename,
		          document_type, document_number, posting_date, account, account_name,
		          keterangan, debet, credit, net, created_at, updated_at)
		          VALUES (0, :session_code, :user_id, :file_path, :filename,
		          :document_type, :document_number, :posting_date, :account, :account_name,
		          :keterangan, :debet, :credit, :net, NOW(), NOW())`

		result, err := r.db.NamedExec(query, chunk)
		if err != nil {
			fmt.Printf("DEBUG: CreateMultipleTransactions ERROR in chunk %d-%d: %v\n", i+1, end, err)
			return err
		}

		rowsAffected, _ := result.RowsAffected()
		totalInserted += rowsAffected
		fmt.Printf("DEBUG: CreateMultipleTransactions SUCCESS: Inserted %d rows in chunk %d-%d (total: %d)\n",
			rowsAffected, i+1, end, totalInserted)
	}

	fmt.Printf("DEBUG: CreateMultipleTransactions FINAL SUCCESS: Total %d rows inserted in %d chunks for session_code: %s\n",
		totalInserted, (len(transactions)+CHUNK_SIZE-1)/CHUNK_SIZE, sessionCode)
	return nil
}

// UpdateTransactionsSessionID updates session_id for transactions with given session_code
func (r *UploadRepository) UpdateTransactionsSessionID(sessionCode string, sessionID int) error {
	// First check how many records will be updated
	var countBefore int
	checkQuery := `SELECT COUNT(*) FROM transaction_data WHERE session_code = ? AND session_id = 0`
	err := r.db.Get(&countBefore, checkQuery, sessionCode)
	if err != nil {
		fmt.Printf("DEBUG: Error checking update count: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: Found %d records to update with session_code: %s\n", countBefore, sessionCode)

	// Debug: Show all records with session_id = 0 to see what's there
	var allZeroSessions []struct {
		ID          int64  `db:"id"`
		SessionID   int    `db:"session_id"`
		SessionCode string `db:"session_code"`
		Filename    string `db:"filename"`
	}
	allZeroQuery := `SELECT id, session_id, session_code, filename FROM transaction_data WHERE session_id = 0 LIMIT 10`
	err = r.db.Select(&allZeroSessions, allZeroQuery)
	if err == nil {
		fmt.Printf("DEBUG: Found %d records with session_id = 0:\n", len(allZeroSessions))
		for _, record := range allZeroSessions {
			fmt.Printf("  - ID: %d, session_id: %d, session_code: '%s', filename: '%s'\n",
				record.ID, record.SessionID, record.SessionCode, record.Filename)
		}
	}

	// Debug: Show all records with any session_code
	var allSessionCodes []struct {
		ID          int64  `db:"id"`
		SessionID   int    `db:"session_id"`
		SessionCode string `db:"session_code"`
		Filename    string `db:"filename"`
	}
	allCodesQuery := `SELECT id, session_id, session_code, filename FROM transaction_data WHERE session_code IS NOT NULL LIMIT 10`
	err = r.db.Select(&allSessionCodes, allCodesQuery)
	if err == nil {
		fmt.Printf("DEBUG: Found %d records with non-NULL session_code:\n", len(allSessionCodes))
		for _, record := range allSessionCodes {
			fmt.Printf("  - ID: %d, session_id: %d, session_code: '%s', filename: '%s'\n",
				record.ID, record.SessionID, record.SessionCode, record.Filename)
		}
	}

	query := `UPDATE transaction_data SET session_id = ? WHERE session_code = ? AND session_id = 0`
	result, err := r.db.Exec(query, sessionID, sessionCode)
	if err != nil {
		fmt.Printf("DEBUG: Error updating transactions: %v\n", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("DEBUG: Updated %d rows with session_id: %d for session_code: %s\n", rowsAffected, sessionID, sessionCode)

	// Verify the update
	var countAfter int
	verifyQuery := `SELECT COUNT(*) FROM transaction_data WHERE session_code = ? AND session_id = ?`
	err = r.db.Get(&countAfter, verifyQuery, sessionCode, sessionID)
	if err != nil {
		fmt.Printf("DEBUG: Error verifying update: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: Verified %d records now have session_id: %d for session_code: %s\n", countAfter, sessionID, sessionCode)

	return nil
}

func (r *UploadRepository) BulkInsertTransactions(transactions []models.TransactionData) error {
	if len(transactions) == 0 {
		return nil
	}

	// Use chunking to avoid placeholder limit
	const CHUNK_SIZE = 5000

	for i := 0; i < len(transactions); i += CHUNK_SIZE {
		end := i + CHUNK_SIZE
		if end > len(transactions) {
			end = len(transactions)
		}

		chunk := transactions[i:end]

		query := `INSERT INTO transaction_data (session_id, document_type, document_number,
		          posting_date, account, account_name, keterangan, debet, credit, net)
		          VALUES (:session_id, :document_type, :document_number, :posting_date,
		          :account, :account_name, :keterangan, :debet, :credit, :net)`

		_, err := r.db.NamedExec(query, chunk)
		if err != nil {
			return fmt.Errorf("error inserting chunk %d-%d: %w", i+1, end, err)
		}
	}

	return nil
}

// CreateBackgroundJob creates a background job record for large uploads
func (r *UploadRepository) CreateBackgroundJob(sessionCode string, userID int, filename string, totalRows int) (*models.BackgroundJob, error) {
	job := &models.BackgroundJob{
		SessionCode: sessionCode,
		UserID:      userID,
		Filename:    filename,
		TotalRows:   totalRows,
		Status:      "pending",
	}

	query := `INSERT INTO background_jobs (session_code, user_id, filename, total_rows, status)
	          VALUES (:session_code, :user_id, :filename, :total_rows, :status)`

	result, err := r.db.NamedExec(query, job)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	job.ID = int(id)
	return job, nil
}

// UpdateBackgroundJobProgress updates background job progress
func (r *UploadRepository) UpdateBackgroundJobProgress(jobID int, processedRows int, status string, errorMsg *string) error {
	query := `UPDATE background_jobs SET processed_rows = ?, status = ?, updated_at = NOW()`
	args := []interface{}{processedRows, status}

	if errorMsg != nil {
		query += `, error_message = ?`
		args = append(args, *errorMsg)
	}

	query += ` WHERE id = ?`
	args = append(args, jobID)

	_, err := r.db.Exec(query, args...)
	return err
}

// GetBackgroundJobBySessionCode retrieves background job by session code
func (r *UploadRepository) GetBackgroundJobBySessionCode(sessionCode string) (*models.BackgroundJob, error) {
	var job models.BackgroundJob
	query := "SELECT * FROM background_jobs WHERE session_code = ? ORDER BY created_at DESC LIMIT 1"
	err := r.db.Get(&job, query, sessionCode)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *UploadRepository) GetTransactionsBySession(sessionID int, limit, offset int) ([]models.TransactionData, int, error) {
	var transactions []models.TransactionData
	var total int

	countQuery := "SELECT COUNT(*) FROM transaction_data WHERE session_id = ?"
	err := r.db.Get(&total, countQuery, sessionID)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT * FROM transaction_data WHERE session_id = ? ORDER BY id LIMIT ? OFFSET ?"
	err = r.db.Select(&transactions, query, sessionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetTransactionsBySessionCode retrieves transactions using session_code with JOIN to accounts
func (r *UploadRepository) GetTransactionsBySessionCode(sessionCode string, limit, offset int) ([]models.TransactionData, int, error) {
	var transactions []models.TransactionData
	var total int

	countQuery := "SELECT COUNT(*) FROM transaction_data WHERE session_code = ?"
	err := r.db.Get(&total, countQuery, sessionCode)
	if err != nil {
		return nil, 0, err
	}

	
	// Query dengan JOIN ke tabel accounts untuk mendapatkan nature_akun dan analisa_kot
	query := `SELECT
				td.document_type,
				td.document_number,
				td.posting_date,
				td.account,
				td.account_name,
				td.keterangan,
				td.debet,
				td.credit,
				td.net,
				td.id,
				td.session_id,
				td.session_code,
				td.user_id,
				td.file_path,
				td.filename,
				accounts.nature as nature_akun,
				-- Analisa KOT: propagate label from ANY row with same document_number
				(
					SELECT MAX(acc_with.koreksi_obyek)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number 
						AND td_sub.session_code = td.session_code
						AND acc_with.koreksi_obyek IS NOT NULL 
						AND acc_with.koreksi_obyek != ''
				) as analisa_kot,
				td.koreksi,
				td.obyek,
				td.um_pajak_db,
				td.pm_db,
				td.wth_21_cr,
				td.wth_23_cr,
				td.wth_26_cr,
				td.wth_4_2_cr,
				td.wth_15_cr,
				td.pk_cr,
				td.analisa_tambahan,
				td.is_processed,
				td.processing_error,
				td.created_at,
				td.updated_at,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 4.2 Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number AND td_sub.session_code = td.session_code
				) as withholding_pph_42,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 15 Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number AND td_sub.session_code = td.session_code
				) as withholding_pph_15,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 21 Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number 
						AND td_sub.session_code = td.session_code
						AND acc_with.koreksi_obyek = 'Wth 21 Cr'
				) as withholding_pph_21,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 23 Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number 
						AND td_sub.session_code = td.session_code
						AND acc_with.koreksi_obyek = 'Wth 23 Cr'
				) as withholding_pph_23,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 26 Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number 
						AND td_sub.session_code = td.session_code
						AND acc_with.koreksi_obyek = 'Wth 26 Cr'
				) as withholding_pph_26,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'PK Cr' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number AND td_sub.session_code = td.session_code
				) as pk_cr_account,
				(
					SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'PM DB' THEN acc_with.account_name ELSE NULL END)
					FROM transaction_data td_sub
					LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
					WHERE td_sub.document_number = td.document_number 
						AND td_sub.session_code = td.session_code
						AND acc_with.koreksi_obyek = 'PM DB'
				) as pm_db_account
			  FROM transaction_data td
			  LEFT JOIN accounts ON td.account = accounts.account_code
			  WHERE td.session_code = ?
			  ORDER BY td.id
			  LIMIT ? OFFSET ?`

	err = r.db.Select(&transactions, query, sessionCode, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetUnprocessedTransactionsBySessionCode gets unprocessed transactions by session_code
func (r *UploadRepository) GetUnprocessedTransactionsBySessionCode(sessionCode string, limit int) ([]models.TransactionData, error) {
	var transactions []models.TransactionData
	query := `SELECT * FROM transaction_data WHERE session_code = ? AND is_processed = FALSE
	          ORDER BY id LIMIT ?`
	err := r.db.Select(&transactions, query, sessionCode, limit)
	return transactions, err
}

func (r *UploadRepository) GetUnprocessedTransactions(sessionID int, limit int) ([]models.TransactionData, error) {
	var transactions []models.TransactionData
	query := `SELECT * FROM transaction_data WHERE session_id = ? AND is_processed = FALSE
	          ORDER BY id LIMIT ?`
	err := r.db.Select(&transactions, query, sessionID, limit)
	return transactions, err
}

func (r *UploadRepository) UpdateTransaction(transaction *models.TransactionData) error {
	query := `UPDATE transaction_data SET
	          analisa_nature_akun = :analisa_nature_akun,
	          analisa_koreksi_obyek = :analisa_koreksi_obyek,
	          koreksi = :koreksi,
	          obyek = :obyek,
	          um_pajak_db = :um_pajak_db,
	          pm_db = :pm_db,
	          wth_21_cr = :wth_21_cr,
	          wth_23_cr = :wth_23_cr,
	          wth_26_cr = :wth_26_cr,
	          wth_4_2_cr = :wth_4_2_cr,
	          wth_15_cr = :wth_15_cr,
	          pk_cr = :pk_cr,
	          analisa_tambahan = :analisa_tambahan,
	          is_processed = :is_processed,
	          processing_error = :processing_error
	          WHERE id = :id`
	_, err := r.db.NamedExec(query, transaction)
	return err
}

func (r *UploadRepository) BulkUpdateTransactions(transactions []models.TransactionData) error {
	if len(transactions) == 0 {
		return nil
	}

	// Update each transaction individually to handle NullableNumericFloat64 properly
	for _, tx := range transactions {
		query := `UPDATE transaction_data SET
		          analisa_nature_akun = ?,
		          analisa_koreksi_obyek = ?,
		          koreksi = ?,
		          obyek = ?,
		          um_pajak_db = ?,
		          pm_db = ?,
		          wth_21_cr = ?,
		          wth_23_cr = ?,
		          wth_26_cr = ?,
		          wth_4_2_cr = ?,
		          wth_15_cr = ?,
		          pk_cr = ?,
		          analisa_tambahan = ?,
		          is_processed = ?,
		          processing_error = ?
		          WHERE id = ?`

		// Convert NullableNumericFloat64 to proper SQL types
		var umPajakDB, pmDB, wth21Cr, wth23Cr, wth26Cr, wth42Cr, wth15Cr, pkCr interface{}

		if tx.UmPajakDB.Valid {
			umPajakDB = tx.UmPajakDB.Value
		}
		if tx.PmDB.Valid {
			pmDB = tx.PmDB.Value
		}
		if tx.Wth21Cr.Valid {
			wth21Cr = tx.Wth21Cr.Value
		}
		if tx.Wth23Cr.Valid {
			wth23Cr = tx.Wth23Cr.Value
		}
		if tx.Wth26Cr.Valid {
			wth26Cr = tx.Wth26Cr.Value
		}
		if tx.Wth42Cr.Valid {
			wth42Cr = tx.Wth42Cr.Value
		}
		if tx.Wth15Cr.Valid {
			wth15Cr = tx.Wth15Cr.Value
		}
		if tx.PkCr.Valid {
			pkCr = tx.PkCr.Value
		}

		_, err := r.db.Exec(query,
			tx.AnalisaNatureAkun,
			tx.AnalisaKoreksiObyek,
			tx.Koreksi,
			tx.Obyek,
			umPajakDB,
			pmDB,
			wth21Cr,
			wth23Cr,
			wth26Cr,
			wth42Cr,
			wth15Cr,
			pkCr,
			tx.AnalisaTambahan,
			tx.IsProcessed,
			tx.ProcessingError,
			tx.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateTransactionKoreksiObyek updates koreksi and obyek fields of a transaction
// Note: koreksi and obyek are NOT propagated - each row has its own values
func (r *UploadRepository) UpdateTransactionKoreksiObyek(transactionID int64, koreksi, obyek *string, userID int, userRole string) error {
	// First check if user has permission to update this transaction
	if userRole != "admin" {
		// For non-admin users, check if they own this transaction
		var count int
		checkQuery := "SELECT COUNT(*) FROM transaction_data WHERE id = ? AND user_id = ?"
		err := r.db.Get(&count, checkQuery, transactionID, userID)
		if err != nil {
			return fmt.Errorf("failed to check transaction ownership: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("transaction not found or access denied")
		}
	}

	// Update koreksi and obyek fields (only for this specific row)
	query := `
		UPDATE transaction_data
		SET
			koreksi = ?,
			obyek = ?,
			analisa_koreksi_obyek = CASE
				WHEN ? IS NOT NULL AND ? IS NOT NULL THEN CONCAT(?, ' - ', ?)
				WHEN ? IS NOT NULL AND ? IS NULL THEN ?
				WHEN ? IS NULL AND ? IS NOT NULL THEN ?
				ELSE NULL
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	var koreksiValue, obyekValue interface{}
	if koreksi != nil && *koreksi != "" {
		koreksiValue = *koreksi
	}
	if obyek != nil && *obyek != "" {
		obyekValue = *obyek
	}

	_, err := r.db.Exec(query,
		koreksiValue, obyekValue, // 1-2: SET koreksi = ?, obyek = ?
		koreksiValue, obyekValue, koreksiValue, obyekValue, // 3-6: WHEN both not null THEN CONCAT
		koreksiValue, obyekValue, // 7-8: WHEN koreksi not null and obyek null
		koreksiValue,             // 9: THEN koreksi
		koreksiValue, obyekValue, // 10-11: WHEN koreksi null and obyek not null
		obyekValue,    // 12: THEN obyek
		transactionID, // 13: WHERE id = ?
	)

	return err
}

// Delete operations
func (r *UploadRepository) DeleteSession(id int) error {
	query := "DELETE FROM upload_sessions WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UploadRepository) DeleteTransactionsBySession(sessionID int) error {
	query := "DELETE FROM transaction_data WHERE session_id = ?"
	_, err := r.db.Exec(query, sessionID)
	return err
}

// DeleteTransactionsBySessionCode deletes all transactions with the given session_code
func (r *UploadRepository) DeleteTransactionsBySessionCode(sessionCode string) error {
	query := "DELETE FROM transaction_data WHERE session_code = ?"
	_, err := r.db.Exec(query, sessionCode)
	return err
}

// Cursor Pagination Methods

// GetSessionsWithCursor - Cursor-based pagination for upload sessions (OPTIMIZED)
func (r *UploadRepository) GetSessionsWithCursor(params utils.PaginationParams, userID int, maxRecords int) ([]models.UploadSession, utils.PaginationMeta, error) {
	var sessions []models.UploadSession
	var total int64
	var hasMore bool
	var nextCursor *utils.Cursor

	// Validate limit against maximum allowed
	limit := utils.ValidatePaginationLimit(params.Limit, maxRecords)

	// OPTIMIZATION: Use separate queries for count and data to avoid complex joins

	// Step 1: Build base WHERE conditions (simplified)
	whereConditions := []string{}
	countArgs := []interface{}{}
	dataArgs := []interface{}{}

	// Add user filter if specified
	if userID > 0 {
		whereConditions = append(whereConditions, "user_id = ?")
		countArgs = append(countArgs, userID)
		dataArgs = append(dataArgs, userID)
	}

	// OPTIMIZATION: Simplified search - only search session_code (most common use case)
	if params.Search != "" {
		searchCondition := fmt.Sprintf("session_code LIKE '%%%s%%'", params.Search)
		whereConditions = append(whereConditions, searchCondition)
		countArgs = append(countArgs, params.Search)
		dataArgs = append(dataArgs, params.Search)
	}

	// OPTIMIZATION: Simplified column filters - only use most important ones
	if params.Filters != nil {
		if status, ok := params.Filters["status"]; ok && status != "" {
			whereConditions = append(whereConditions, "status = ?")
			countArgs = append(countArgs, status)
			dataArgs = append(dataArgs, status)
		}

		if sessionCode, ok := params.Filters["session_code"]; ok && sessionCode != "" {
			sessionCodeStr := fmt.Sprintf("%v", sessionCode)
			whereConditions = append(whereConditions, "session_code LIKE ?")
			countArgs = append(countArgs, "%"+sessionCodeStr+"%")
			dataArgs = append(dataArgs, "%"+sessionCodeStr+"%")
		}
	}

	// Step 2: Simple count query (no complex filters for performance)
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// OPTIMIZATION: Fast count query
	countQuery := "SELECT COUNT(*) FROM upload_sessions " + whereClause
	err := r.db.Get(&total, countQuery, countArgs...)
	if err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	// Apply total limit
	if total > int64(maxRecords) {
		total = int64(maxRecords)
	}

	// Step 3: Simplified cursor handling
	var cursor *utils.Cursor
	if params.Cursor != "" && params.Mode == "cursor" {
		cursor, err = utils.DecodeCursor(params.Cursor)
		if err == nil {
			// Simple cursor condition - only use created_at for consistency
			if whereClause == "" {
				whereClause = "WHERE created_at < ?"
			} else {
				whereClause += " AND created_at < ?"
			}
			dataArgs = append(dataArgs, cursor.CreatedAt)
		}
	}

	// OPTIMIZATION: Simplified order by - always use created_at DESC for consistency
	orderByClause := "ORDER BY created_at DESC, id DESC"

	// Step 4: Main data query (optimized)
	query := fmt.Sprintf(`
		SELECT id, session_code, user_id, filename, file_path,
		       total_rows, processed_rows, failed_rows, status,
		       error_message, created_at, updated_at
		FROM upload_sessions
		%s
		%s
		LIMIT ?`, whereClause, orderByClause)

	// Add limit with +1 for hasMore check
	dataArgs = append(dataArgs, limit+1)

	err = r.db.Select(&sessions, query, dataArgs...)
	if err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	// Step 5: Process results
	if len(sessions) > limit {
		hasMore = true
		sessions = sessions[:limit] // Remove the extra record
	} else {
		hasMore = false
	}

	// Build next cursor (simplified - only use created_at)
	if hasMore && len(sessions) > 0 {
		lastSession := sessions[len(sessions)-1]
		nextCursor = &utils.Cursor{
			ID:        lastSession.ID,
			CreatedAt: lastSession.CreatedAt,
		}
	}

	// Calculate pagination metadata
	paginationMeta := utils.CalculateCursorPagination(
		limit,
		total,
		hasMore,
		nextCursor,
		nil, // Simplified - no prev cursor for performance
		params.Mode,
	)

	return sessions, paginationMeta, nil
}

// GetTransactionsBySessionCodeWithCursor - Cursor-based pagination for transactions
func (r *UploadRepository) GetTransactionsBySessionCodeWithCursor(
	sessionCode string,
	params utils.PaginationParams,
	maxRecords int,
) ([]models.TransactionData, utils.PaginationMeta, error) {
	var transactions []models.TransactionData
	var total int64
	var hasMore bool
	var nextCursor *utils.Cursor
	var prevCursor *utils.Cursor

	// Validate limit against maximum allowed
	limit := utils.ValidatePaginationLimit(params.Limit, maxRecords)

	// Get total count - no max limit for session detail to show all records
	countQuery := "SELECT COUNT(*) FROM transaction_data WHERE session_code = ?"
	err := r.db.Get(&total, countQuery, sessionCode)
	if err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	// Only apply limit for pagination pages, not for total count
	// This allows users to see all records in session detail

	// Build cursor condition
	var cursor *utils.Cursor
	whereClause := "WHERE td.session_code = ?"
	args := []interface{}{sessionCode}

	if params.Cursor != "" && params.Mode == "cursor" {
		var err error
		cursor, err = utils.DecodeCursor(params.Cursor)
		if err != nil {
			// Invalid cursor, fall back to first page
			cursor = nil
		} else {
			// Add cursor condition
			cursorCondition := utils.BuildTransactionCursorCondition(cursor, "desc", sessionCode)
			if cursorCondition != "" {
				whereClause += " AND " + cursorCondition
			}
		}
	}

	// Build order by clause (default: by ID for consistency)
	orderByClause := "ORDER BY td.id DESC"
	if params.OrderBy != "" && params.OrderDir != "" {
		switch strings.ToLower(params.OrderBy) {
		case "id":
			orderByClause = fmt.Sprintf("ORDER BY td.id %s", params.OrderDir)
		case "posting_date":
			orderByClause = fmt.Sprintf("ORDER BY td.posting_date %s, td.id DESC", params.OrderDir)
		case "account":
			orderByClause = fmt.Sprintf("ORDER BY td.account %s, td.id DESC", params.OrderDir)
		case "debet":
			orderByClause = fmt.Sprintf("ORDER BY td.debet %s, td.id DESC", params.OrderDir)
		case "credit":
			orderByClause = fmt.Sprintf("ORDER BY td.credit %s, td.id DESC", params.OrderDir)
		default:
			orderByClause = fmt.Sprintf("ORDER BY td.id %s", params.OrderDir)
		}
	}

	// Main query with subqueries for propagation - ensures all rows with same document_number show same labels
	// This uses subqueries to find account names from ANY row with the same document_number
	query := fmt.Sprintf(`
		SELECT
			td.document_type,
			td.document_number,
			td.posting_date,
			td.account,
			td.account_name,
			td.keterangan,
			td.debet,
			td.credit,
			td.net,
			td.id,
			td.session_id,
			td.session_code,
			td.user_id,
			td.file_path,
			td.filename,
			td.analisa_nature_akun,
			td.analisa_koreksi_obyek,
			td.koreksi,
			td.obyek,
			td.um_pajak_db,
			td.pm_db,
			td.wth_21_cr,
			td.wth_23_cr,
			td.wth_26_cr,
			td.wth_4_2_cr,
			td.wth_15_cr,
			td.pk_cr,
			td.analisa_tambahan,
			td.is_processed,
			td.processing_error,
			td.created_at,
			td.updated_at,
			accounts.nature as nature_akun,
			-- Analisa KOT: propagate label from ANY row with same document_number
			(
				SELECT MAX(acc_with.koreksi_obyek)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek IS NOT NULL 
					AND acc_with.koreksi_obyek != ''
			) as analisa_kot,
			-- Withholding account names from ANY row with same document_number (propagation)
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 4.2 Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'Wth 4.2 Cr'
			) as withholding_pph_42,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 15 Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'Wth 15 Cr'
			) as withholding_pph_15,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 21 Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'Wth 21 Cr'
			) as withholding_pph_21,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 23 Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'Wth 23 Cr'
			) as withholding_pph_23,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'Wth 26 Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'Wth 26 Cr'
			) as withholding_pph_26,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'PK Cr' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'PK Cr'
			) as pk_cr_account,
			(
				SELECT MAX(CASE WHEN acc_with.koreksi_obyek = 'PM DB' THEN acc_with.account_name ELSE NULL END)
				FROM transaction_data td_sub
				LEFT JOIN accounts acc_with ON td_sub.account = acc_with.account_code
				WHERE td_sub.document_number = td.document_number 
					AND td_sub.session_code = td.session_code
					AND acc_with.koreksi_obyek = 'PM DB'
			) as pm_db_account
		FROM transaction_data td
		-- Main account join for current row's account
		LEFT JOIN accounts ON td.account = accounts.account_code
		%s
		%s
		LIMIT ?`, whereClause, orderByClause)

	args = append(args, limit+1) // +1 to check if there are more records

	err = r.db.Select(&transactions, query, args...)
	if err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	// Check if there are more records
	if len(transactions) > limit {
		hasMore = true
		transactions = transactions[:limit] // Remove the extra record
	} else {
		hasMore = false
	}

	// Build next cursor
	if hasMore && len(transactions) > 0 {
		lastTransaction := transactions[len(transactions)-1]
		nextCursor = &utils.Cursor{
			ID:         int(lastTransaction.ID),
			SessionCode: lastTransaction.SessionCode,
		}
	}

	// Build prev cursor
	if cursor != nil {
		if len(transactions) > 0 {
			firstTransaction := transactions[0]
			prevCursor = &utils.Cursor{
				ID:         int(firstTransaction.ID),
				SessionCode: firstTransaction.SessionCode,
			}
		}
	}

	// Calculate pagination metadata
	paginationMeta := utils.CalculateCursorPagination(
		limit,
		total,
		hasMore,
		nextCursor,
		prevCursor,
		params.Mode,
	)

	return transactions, paginationMeta, nil
}

// GetTotalSessionsCount returns total count of sessions (with maximum limit)
func (r *UploadRepository) GetTotalSessionsCount(userID int, maxRecords int) (int64, error) {
	var total int64

	query := "SELECT COUNT(*) FROM upload_sessions"
	args := []interface{}{}

	if userID > 0 {
		query += " WHERE user_id = ?"
		args = append(args, userID)
	}

	err := r.db.Get(&total, query, args...)
	if err != nil {
		return 0, err
	}

	// Apply maximum limit
	if total > int64(maxRecords) {
		total = int64(maxRecords)
	}

	return total, nil
}

// GetTotalTransactionsBySessionCodeCount returns total count of transactions for a session (with maximum limit)
func (r *UploadRepository) GetTotalTransactionsBySessionCodeCount(sessionCode string, maxRecords int) (int64, error) {
	var total int64

	query := "SELECT COUNT(*) FROM transaction_data WHERE session_code = ?"
	err := r.db.Get(&total, query, sessionCode)
	if err != nil {
		return 0, err
	}

	// Apply maximum limit
	if total > int64(maxRecords) {
		total = int64(maxRecords)
	}

	return total, nil
}

// PropagateDocumentNumberFields propagates field values to all rows with the same document_number
// If a document_number has a row with account that has koreksi_obyek matching a flag (PK Cr, Wth 15 Cr, etc.),
// then ALL rows with that document_number will have the value populated in the corresponding field.
// Logic:
// - For WTH/PK CR fields: If there's a calculated value, use it; otherwise use credit from the row with matching account
// - For PM DB/UM Pajak DB: If there's a calculated value, use it; otherwise use debet from the row with matching account
// Note: koreksi and obyek are NOT propagated - each row has its own values (independent per row)
func (r *UploadRepository) PropagateDocumentNumberFields(sessionCode string) error {
	// Use a single query to propagate all fields at once for better performance
	// This query finds values from rows that have accounts with matching koreksi_obyek flags
	// and propagates those values to all rows with the same document_number
	
	propagateQuery := `
		UPDATE transaction_data td_target
		INNER JOIN (
			SELECT 
				td.document_number,
				td.session_code,
				-- For WTH 4.2 Cr: if document_number has account with 'Wth 4.2 Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'Wth 4.2 Cr' THEN 
						COALESCE(
							NULLIF(td.wth_4_2_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as wth_4_2_cr_val,
				-- For WTH 15 Cr: if document_number has account with 'Wth 15 Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'Wth 15 Cr' THEN 
						COALESCE(
							NULLIF(td.wth_15_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as wth_15_cr_val,
				-- For WTH 21 Cr: if document_number has account with 'Wth 21 Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'Wth 21 Cr' THEN 
						COALESCE(
							NULLIF(td.wth_21_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as wth_21_cr_val,
				-- For WTH 23 Cr: if document_number has account with 'Wth 23 Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'Wth 23 Cr' THEN 
						COALESCE(
							NULLIF(td.wth_23_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as wth_23_cr_val,
				-- For WTH 26 Cr: if document_number has account with 'Wth 26 Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'Wth 26 Cr' THEN 
						COALESCE(
							NULLIF(td.wth_26_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as wth_26_cr_val,
				-- For PK CR: if document_number has account with 'PK Cr' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'PK Cr' THEN 
						COALESCE(
							NULLIF(td.pk_cr, 0), 
							CASE WHEN td.credit > 0 THEN td.credit ELSE NULL END
						)
					ELSE NULL 
				END) as pk_cr_val,
				-- For PM DB: if document_number has account with 'PM DB' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'PM DB' THEN 
						COALESCE(
							NULLIF(td.pm_db, 0), 
							CASE WHEN td.debet > 0 THEN td.debet ELSE NULL END
						)
					ELSE NULL 
				END) as pm_db_val,
				-- For UM Pajak DB: if document_number has account with 'UM Pajak DB' flag, propagate the value
				MAX(CASE 
					WHEN a.koreksi_obyek = 'UM Pajak DB' THEN 
						COALESCE(
							NULLIF(td.um_pajak_db, 0), 
							CASE WHEN td.debet > 0 THEN td.debet ELSE NULL END
						)
					ELSE NULL 
				END) as um_pajak_db_val
				-- Note: koreksi and obyek are NOT propagated - each row has its own values
			FROM transaction_data td
			INNER JOIN accounts a ON td.account = a.account_code
			WHERE td.session_code = ?
				AND td.document_number IS NOT NULL
				AND td.document_number != ''
			GROUP BY td.document_number, td.session_code
		) td_source ON td_target.document_number = td_source.document_number 
			AND td_target.session_code = td_source.session_code
		SET 
			-- If flag exists for this document_number, set the value (even if it's 0, it will override existing value)
			td_target.wth_4_2_cr = COALESCE(td_source.wth_4_2_cr_val, td_target.wth_4_2_cr),
			td_target.wth_15_cr = COALESCE(td_source.wth_15_cr_val, td_target.wth_15_cr),
			td_target.wth_21_cr = COALESCE(td_source.wth_21_cr_val, td_target.wth_21_cr),
			td_target.wth_23_cr = COALESCE(td_source.wth_23_cr_val, td_target.wth_23_cr),
			td_target.wth_26_cr = COALESCE(td_source.wth_26_cr_val, td_target.wth_26_cr),
			td_target.pk_cr = COALESCE(td_source.pk_cr_val, td_target.pk_cr),
			td_target.pm_db = COALESCE(td_source.pm_db_val, td_target.pm_db),
			td_target.um_pajak_db = COALESCE(td_source.um_pajak_db_val, td_target.um_pajak_db)
			-- Note: koreksi and obyek are NOT updated here - each row keeps its own values
		WHERE td_target.session_code = ?
			AND td_target.document_number IS NOT NULL
			AND td_target.document_number != ''
	`
	
	_, err := r.db.Exec(propagateQuery, sessionCode, sessionCode)
	if err != nil {
		return fmt.Errorf("failed to propagate document number fields: %w", err)
	}

	return nil
}
