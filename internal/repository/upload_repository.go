package repository

import (
	"accounting-web/internal/models"

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

	if userID > 0 {
		whereClause = "WHERE user_id = ?"
		args = append(args, userID)
	}

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

// GetBatchUploads gets batch uploads from transaction_data that don't have matching upload_sessions
func (r *UploadRepository) GetBatchUploads(limit, offset int, userID int) ([]models.BatchUploadSession, int, error) {
	var batches []models.BatchUploadSession
	var total int

	whereClause := "WHERE t.session_id = 0"
	args := []interface{}{}

	if userID > 0 {
		whereClause += " AND t.user_id = ?"
		args = append(args, userID)
	}

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
	err = r.db.Select(&batches, query, args...)
	if err != nil {
		return nil, 0, err
	}

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

// Transaction Data
func (r *UploadRepository) CreateMultipleTransactions(transactions []models.TransactionData) error {
	if len(transactions) == 0 {
		return nil
	}

	// Since session_id is NOT NULL in database, we need to create a dummy session
	// or use a special value to identify batch uploads
	// We'll use session_id = 0 and session_code for batch identification
	query := `INSERT INTO transaction_data (session_id, session_code, user_id, file_path, filename,
	          document_type, document_number, posting_date, account, account_name,
	          keterangan, debet, credit, net)
	          VALUES (0, :session_code, :user_id, :file_path, :filename,
	          :document_type, :document_number, :posting_date, :account, :account_name,
	          :keterangan, :debet, :credit, :net)`

	_, err := r.db.NamedExec(query, transactions)
	return err
}

// UpdateTransactionsSessionID updates session_id for transactions with given session_code
func (r *UploadRepository) UpdateTransactionsSessionID(sessionCode string, sessionID int) error {
	query := `UPDATE transaction_data SET session_id = ? WHERE session_code = ? AND session_id = 0`
	_, err := r.db.Exec(query, sessionID, sessionCode)
	return err
}

func (r *UploadRepository) BulkInsertTransactions(transactions []models.TransactionData) error {
	if len(transactions) == 0 {
		return nil
	}

	query := `INSERT INTO transaction_data (session_id, document_type, document_number,
	          posting_date, account, account_name, keterangan, debet, credit, net)
	          VALUES (:session_id, :document_type, :document_number, :posting_date,
	          :account, :account_name, :keterangan, :debet, :credit, :net)`

	_, err := r.db.NamedExec(query, transactions)
	return err
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

		_, err := r.db.Exec(query,
			tx.AnalisaNatureAkun,
			tx.AnalisaKoreksiObyek,
			tx.Koreksi,
			tx.Obyek,
			tx.UmPajakDB,
			tx.PmDB,
			tx.Wth21Cr,
			tx.Wth23Cr,
			tx.Wth26Cr,
			tx.Wth42Cr,
			tx.Wth15Cr,
			tx.PkCr,
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
