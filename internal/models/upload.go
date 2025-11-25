package models

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// NumericFloat64 handles conversion from SQL string/decimal to float64
type NumericFloat64 float64

// Scan implements sql.Scanner interface for NumericFloat64
func (n *NumericFloat64) Scan(value interface{}) error {
	if value == nil {
		*n = 0
		return nil
	}

	switch v := value.(type) {
	case float64:
		*n = NumericFloat64(v)
		return nil
	case int, int32, int64:
		*n = NumericFloat64(float64(v.(int64)))
		return nil
	case string:
		// Handle empty string as 0
		if strings.TrimSpace(v) == "" {
			*n = 0
			return nil
		}
		// Try to parse as float
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			*n = NumericFloat64(f)
			return nil
		}
		*n = 0
		return nil
	case []byte:
		// Handle empty bytes as 0
		str := strings.TrimSpace(string(v))
		if str == "" {
			*n = 0
			return nil
		}
		// Try to parse as float
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			*n = NumericFloat64(f)
			return nil
		}
		*n = 0
		return nil
	case sql.NullFloat64:
		if v.Valid {
			*n = NumericFloat64(v.Float64)
		} else {
			*n = 0
		}
		return nil
	default:
		*n = 0
		return nil
	}
}

// Float64 returns the float64 value
func (n NumericFloat64) Float64() float64 {
	return float64(n)
}

// MarshalJSON implements json.Marshaler
func (n NumericFloat64) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Float64())
}

// NullableNumericFloat64 handles nullable numeric fields
type NullableNumericFloat64 struct {
	Value float64
	Valid  bool
}

// Scan implements sql.Scanner interface for NullableNumericFloat64
func (n *NullableNumericFloat64) Scan(value interface{}) error {
	if value == nil {
		n.Value = 0
		n.Valid = false
		return nil
	}

	switch v := value.(type) {
	case float64:
		n.Value = v
		n.Valid = true
		return nil
	case int, int32, int64:
		n.Value = float64(v.(int64))
		n.Valid = true
		return nil
	case string:
		if strings.TrimSpace(v) == "" {
			n.Value = 0
			n.Valid = false
			return nil
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			n.Value = f
			n.Valid = true
			return nil
		}
		n.Value = 0
		n.Valid = false
		return nil
	case []byte:
		str := strings.TrimSpace(string(v))
		if str == "" {
			n.Value = 0
			n.Valid = false
			return nil
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			n.Value = f
			n.Valid = true
			return nil
		}
		n.Value = 0
		n.Valid = false
		return nil
	case sql.NullFloat64:
		n.Value = v.Float64
		n.Valid = v.Valid
		return nil
	default:
		n.Value = 0
		n.Valid = false
		return nil
	}
}

// Float64 returns the float64 value
func (n NullableNumericFloat64) Float64() float64 {
	return n.Value
}

// IsNull returns true if the value is null/invalid
func (n NullableNumericFloat64) IsNull() bool {
	return !n.Valid || n.Value == 0
}

// MarshalJSON implements json.Marshaler
func (n NullableNumericFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

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
	ErrorMessage  *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// MarshalJSON custom JSON marshaling for UploadSession to handle nullable strings
func (us *UploadSession) MarshalJSON() ([]byte, error) {
	type Alias UploadSession
	aux := &struct {
		ErrorMessage string `json:"error_message,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(us),
	}

	if us.ErrorMessage != nil {
		aux.ErrorMessage = *us.ErrorMessage
	}

	// Handle FilePath as well
	if us.FilePath == "" {
		aux.FilePath = ""
	}

	return json.Marshal(aux)
}

type TransactionData struct {
	ID         int64     `db:"id" json:"id"`
	SessionID  int       `db:"session_id" json:"session_id"`

	// Batch Upload Fields (for multiple file uploads without session)
	SessionCode string `db:"session_code" json:"session_code,omitempty"`
	UserID      int    `db:"user_id" json:"user_id,omitempty"`
	FilePath    string `db:"file_path" json:"file_path,omitempty"`
	Filename    string `db:"filename" json:"filename,omitempty"`

	// Input Fields
	DocumentType   string    `db:"document_type" json:"document_type"`
	DocumentNumber string    `db:"document_number" json:"document_number"`
	PostingDate    *time.Time `db:"posting_date" json:"posting_date,omitempty"`
	Account        string    `db:"account" json:"account"`
	AccountName    string    `db:"account_name" json:"account_name"`
	Keterangan     string    `db:"keterangan" json:"keterangan"`
	Debet          float64   `db:"debet" json:"debet"`
	Credit         float64   `db:"credit" json:"credit"`
	Net            float64   `db:"net" json:"net"`

	// Output Fields
	AnalisaNatureAkun    *string `db:"analisa_nature_akun" json:"analisa_nature_akun,omitempty"`
	AnalisaKoreksiObyek  *string `db:"analisa_koreksi_obyek" json:"analisa_koreksi_obyek,omitempty"`

	// Fields from JOIN with accounts table
	NatureAkun           *string `db:"nature_akun" json:"nature_akun,omitempty"`
	AnalisaKOT           *string `db:"analisa_kot" json:"analisa_kot,omitempty"`

	// Withholding account names from window function queries
	WithholdingPph42     *string `db:"withholding_pph_42" json:"withholding_pph_42,omitempty"`
	WithholdingPph15     *string `db:"withholding_pph_15" json:"withholding_pph_15,omitempty"`
	WithholdingPph21     *string `db:"withholding_pph_21" json:"withholding_pph_21,omitempty"`
	WithholdingPph23     *string `db:"withholding_pph_23" json:"withholding_pph_23,omitempty"`
	WithholdingPph26     *string `db:"withholding_pph_26" json:"withholding_pph_26,omitempty"`
	PkCrAccount          *string `db:"pk_cr_account" json:"pk_cr_account,omitempty"`
	PmDbAccount          *string `db:"pm_db_account" json:"pm_db_account,omitempty"`
	Koreksi              *string `db:"koreksi" json:"koreksi,omitempty"`
	Obyek                *string `db:"obyek" json:"obyek,omitempty"`
	UmPajakDB            NullableNumericFloat64 `db:"um_pajak_db" json:"um_pajak_db,omitempty"`
	PmDB                 NullableNumericFloat64 `db:"pm_db" json:"pm_db,omitempty"`
	Wth21Cr              NullableNumericFloat64 `db:"wth_21_cr" json:"wth_21_cr,omitempty"`
	Wth23Cr              NullableNumericFloat64 `db:"wth_23_cr" json:"wth_23_cr,omitempty"`
	Wth26Cr              NullableNumericFloat64 `db:"wth_26_cr" json:"wth_26_cr,omitempty"`
	Wth42Cr              NullableNumericFloat64 `db:"wth_4_2_cr" json:"wth_4_2_cr,omitempty"`
	Wth15Cr              NullableNumericFloat64 `db:"wth_15_cr" json:"wth_15_cr,omitempty"`
	PkCr                 NullableNumericFloat64 `db:"pk_cr" json:"pk_cr,omitempty"`
	AnalisaTambahan      *string `db:"analisa_tambahan" json:"analisa_tambahan,omitempty"`

	// Processing
	IsProcessed      bool       `db:"is_processed" json:"is_processed"`
	ProcessingError  *string    `db:"processing_error" json:"processing_error,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// MarshalJSON custom JSON marshaling for TransactionData to handle nullable pointers
func (td *TransactionData) MarshalJSON() ([]byte, error) {
	type Alias TransactionData
	aux := &struct {
		AnalisaNatureAkun    *string                `json:"analisa_nature_akun,omitempty"`
		AnalisaKoreksiObyek  *string                `json:"analisa_koreksi_obyek,omitempty"`
		NatureAkun           *string                `json:"nature_akun,omitempty"`
		AnalisaKOT           *string                `json:"analisa_kot,omitempty"`
		WithholdingPph42     *string                `json:"withholding_pph_42,omitempty"`
		WithholdingPph15     *string                `json:"withholding_pph_15,omitempty"`
		WithholdingPph21     *string                `json:"withholding_pph_21,omitempty"`
		WithholdingPph23     *string                `json:"withholding_pph_23,omitempty"`
		WithholdingPph26     *string                `json:"withholding_pph_26,omitempty"`
		PkCrAccount          *string                `json:"pk_cr_account,omitempty"`
		PmDbAccount          *string                `json:"pm_db_account,omitempty"`
		Koreksi              *string                `json:"koreksi,omitempty"`
		Obyek                *string                `json:"obyek,omitempty"`
		UmPajakDB            *float64               `json:"um_pajak_db,omitempty"`
		PmDB                 *float64               `json:"pm_db,omitempty"`
		Wth21Cr              *float64               `json:"wth_21_cr,omitempty"`
		Wth23Cr              *float64               `json:"wth_23_cr,omitempty"`
		Wth26Cr              *float64               `json:"wth_26_cr,omitempty"`
		Wth42Cr              *float64               `json:"wth_4_2_cr,omitempty"`
		Wth15Cr              *float64               `json:"wth_15_cr,omitempty"`
		PkCr                 *float64               `json:"pk_cr,omitempty"`
		AnalisaTambahan      *string                `json:"analisa_tambahan,omitempty"`
		ProcessingError      *string                `json:"processing_error,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(td),
	}

	// Handle nullable numeric fields
	var umPajakDB *float64
	if td.UmPajakDB.Valid {
		umPajakDB = &td.UmPajakDB.Value
	}

	var pmDB *float64
	if td.PmDB.Valid {
		pmDB = &td.PmDB.Value
	}

	var wth21Cr *float64
	if td.Wth21Cr.Valid {
		wth21Cr = &td.Wth21Cr.Value
	}

	var wth23Cr *float64
	if td.Wth23Cr.Valid {
		wth23Cr = &td.Wth23Cr.Value
	}

	var wth26Cr *float64
	if td.Wth26Cr.Valid {
		wth26Cr = &td.Wth26Cr.Value
	}

	var wth42Cr *float64
	if td.Wth42Cr.Valid {
		wth42Cr = &td.Wth42Cr.Value
	}

	var wth15Cr *float64
	if td.Wth15Cr.Valid {
		wth15Cr = &td.Wth15Cr.Value
	}

	var pkCr *float64
	if td.PkCr.Valid {
		pkCr = &td.PkCr.Value
	}

	aux.AnalisaNatureAkun = td.AnalisaNatureAkun
	aux.AnalisaKoreksiObyek = td.AnalisaKoreksiObyek
	aux.NatureAkun = td.NatureAkun
	aux.AnalisaKOT = td.AnalisaKOT
	aux.WithholdingPph42 = td.WithholdingPph42
	aux.WithholdingPph15 = td.WithholdingPph15
	aux.WithholdingPph21 = td.WithholdingPph21
	aux.WithholdingPph23 = td.WithholdingPph23
	aux.WithholdingPph26 = td.WithholdingPph26
	aux.PkCrAccount = td.PkCrAccount
	aux.PmDbAccount = td.PmDbAccount
	aux.Koreksi = td.Koreksi
	aux.Obyek = td.Obyek
	aux.UmPajakDB = umPajakDB
	aux.PmDB = pmDB
	aux.Wth21Cr = wth21Cr
	aux.Wth23Cr = wth23Cr
	aux.Wth26Cr = wth26Cr
	aux.Wth42Cr = wth42Cr
	aux.Wth15Cr = wth15Cr
	aux.PkCr = pkCr
	aux.AnalisaTambahan = td.AnalisaTambahan
	aux.ProcessingError = td.ProcessingError

	return json.Marshal(aux)
}

// BatchUploadSession represents a batch upload session (from transaction_data with session_id = 0)
type BatchUploadSession struct {
	SessionCode   string    `db:"session_code" json:"session_code"`
	UserID        int       `db:"user_id" json:"user_id"`
	Filename      string    `db:"filename" json:"filename"`
	TotalRows     int       `db:"total_rows" json:"total_rows"`
	ProcessedRows int       `db:"processed_rows" json:"processed_rows"`
	FailedRows    int       `db:"failed_rows" json:"failed_rows"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// GetBatchID returns a consistent ID for pagination (based on session_code hash)
func (b *BatchUploadSession) GetBatchID() int {
	// Simple hash of session_code to create a consistent ID
	hash := 0
	for _, char := range b.SessionCode {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// ToUploadSession converts BatchUploadSession to UploadSession-like format for UI compatibility
func (b *BatchUploadSession) ToUploadSession() map[string]interface{} {
	return map[string]interface{}{
		"id":             b.GetBatchID(),
		"session_code":   b.SessionCode,
		"user_id":        b.UserID,
		"filename":       b.Filename,
		"file_path":      "",
		"total_rows":     b.TotalRows,
		"processed_rows": b.ProcessedRows,
		"failed_rows":    b.FailedRows,
		"status":         b.Status,
		"error_message":  nil,
		"created_at":     b.CreatedAt,
		"updated_at":     b.UpdatedAt,
		"is_batch":       true,
	}
}

// BackgroundJob represents a background processing job for large uploads
type BackgroundJob struct {
	ID            int        `db:"id" json:"id"`
	SessionCode   string     `db:"session_code" json:"session_code"`
	UserID        int        `db:"user_id" json:"user_id"`
	Filename      string     `db:"filename" json:"filename"`
	TotalRows     int        `db:"total_rows" json:"total_rows"`
	ProcessedRows int        `db:"processed_rows" json:"processed_rows"`
	Status        string     `db:"status" json:"status"` // pending, processing, completed, failed
	ErrorMessage  *string    `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// GetProgressPercentage returns the progress as a percentage
func (j *BackgroundJob) GetProgressPercentage() float64 {
	if j.TotalRows == 0 {
		return 0
	}
	return float64(j.ProcessedRows) / float64(j.TotalRows) * 100
}
