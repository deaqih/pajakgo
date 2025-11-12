package models

import "time"

type UploadSession struct {
	ID            int       `db:"id" json:"id"`
	SessionCode   string    `db:"session_code" json:"session_code"`
	UserID        int       `db:"user_id" json:"user_id"`
	Filename      string    `db:"filename" json:"filename"`
	FilePath      string    `db:"file_path" json:"file_path"`
	TotalRows     int       `db:"total_rows" json:"total_rows"`
	ProcessedRows int       `db:"processed_rows" json:"processed_rows"`
	FailedRows    int       `db:"failed_rows" json:"failed_rows"`
	Status        string    `db:"status" json:"status"`
	ErrorMessage  string    `db:"error_message" json:"error_message"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type TransactionData struct {
	ID         int64     `db:"id" json:"id"`
	SessionID  int       `db:"session_id" json:"session_id"`

	// Input Fields
	DocumentType   string    `db:"document_type" json:"document_type"`
	DocumentNumber string    `db:"document_number" json:"document_number"`
	PostingDate    time.Time `db:"posting_date" json:"posting_date"`
	Account        string    `db:"account" json:"account"`
	AccountName    string    `db:"account_name" json:"account_name"`
	Keterangan     string    `db:"keterangan" json:"keterangan"`
	Debet          float64   `db:"debet" json:"debet"`
	Credit         float64   `db:"credit" json:"credit"`
	Net            float64   `db:"net" json:"net"`

	// Output Fields
	AnalisaNatureAkun    string  `db:"analisa_nature_akun" json:"analisa_nature_akun"`
	AnalisaKoreksiObyek  string  `db:"analisa_koreksi_obyek" json:"analisa_koreksi_obyek"`
	Koreksi              string  `db:"koreksi" json:"koreksi"`
	Obyek                string  `db:"obyek" json:"obyek"`
	UmPajakDB            float64 `db:"um_pajak_db" json:"um_pajak_db"`
	PmDB                 float64 `db:"pm_db" json:"pm_db"`
	Wth21Cr              float64 `db:"wth_21_cr" json:"wth_21_cr"`
	Wth23Cr              float64 `db:"wth_23_cr" json:"wth_23_cr"`
	Wth26Cr              float64 `db:"wth_26_cr" json:"wth_26_cr"`
	Wth42Cr              float64 `db:"wth_4_2_cr" json:"wth_4_2_cr"`
	Wth15Cr              float64 `db:"wth_15_cr" json:"wth_15_cr"`
	PkCr                 float64 `db:"pk_cr" json:"pk_cr"`
	AnalisaTambahan      string  `db:"analisa_tambahan" json:"analisa_tambahan"`

	// Processing
	IsProcessed      bool      `db:"is_processed" json:"is_processed"`
	ProcessingError  string    `db:"processing_error" json:"processing_error"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
