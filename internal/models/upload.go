package models

import (
	"encoding/json"
	"time"
)

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
	AnalisaNatureAkun    *string `db:"analisa_nature_akun" json:"analisa_nature_akun,omitempty"`
	AnalisaKoreksiObyek  *string `db:"analisa_koreksi_obyek" json:"analisa_koreksi_obyek,omitempty"`
	Koreksi              *string `db:"koreksi" json:"koreksi,omitempty"`
	Obyek                *string `db:"obyek" json:"obyek,omitempty"`
	UmPajakDB            *float64 `db:"um_pajak_db" json:"um_pajak_db,omitempty"`
	PmDB                 *float64 `db:"pm_db" json:"pm_db,omitempty"`
	Wth21Cr              *float64 `db:"wth_21_cr" json:"wth_21_cr,omitempty"`
	Wth23Cr              *float64 `db:"wth_23_cr" json:"wth_23_cr,omitempty"`
	Wth26Cr              *float64 `db:"wth_26_cr" json:"wth_26_cr,omitempty"`
	Wth42Cr              *float64 `db:"wth_4_2_cr" json:"wth_4_2_cr,omitempty"`
	Wth15Cr              *float64 `db:"wth_15_cr" json:"wth_15_cr,omitempty"`
	PkCr                 *float64 `db:"pk_cr" json:"pk_cr,omitempty"`
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
		AnalisaNatureAkun    *string  `json:"analisa_nature_akun,omitempty"`
		AnalisaKoreksiObyek  *string  `json:"analisa_koreksi_obyek,omitempty"`
		Koreksi              *string  `json:"koreksi,omitempty"`
		Obyek                *string  `json:"obyek,omitempty"`
		UmPajakDB            *float64 `json:"um_pajak_db,omitempty"`
		PmDB                 *float64 `json:"pm_db,omitempty"`
		Wth21Cr              *float64 `json:"wth_21_cr,omitempty"`
		Wth23Cr              *float64 `json:"wth_23_cr,omitempty"`
		Wth26Cr              *float64 `json:"wth_26_cr,omitempty"`
		Wth42Cr              *float64 `json:"wth_4_2_cr,omitempty"`
		Wth15Cr              *float64 `json:"wth_15_cr,omitempty"`
		PkCr                 *float64 `json:"pk_cr,omitempty"`
		AnalisaTambahan      *string  `json:"analisa_tambahan,omitempty"`
		ProcessingError      *string  `json:"processing_error,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(td),
	}

	aux.AnalisaNatureAkun = td.AnalisaNatureAkun
	aux.AnalisaKoreksiObyek = td.AnalisaKoreksiObyek
	aux.Koreksi = td.Koreksi
	aux.Obyek = td.Obyek
	aux.UmPajakDB = td.UmPajakDB
	aux.PmDB = td.PmDB
	aux.Wth21Cr = td.Wth21Cr
	aux.Wth23Cr = td.Wth23Cr
	aux.Wth26Cr = td.Wth26Cr
	aux.Wth42Cr = td.Wth42Cr
	aux.Wth15Cr = td.Wth15Cr
	aux.PkCr = td.PkCr
	aux.AnalisaTambahan = td.AnalisaTambahan
	aux.ProcessingError = td.ProcessingError

	return json.Marshal(aux)
}
