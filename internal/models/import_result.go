package models

import "time"

// AccountValidationError represents a validation error for an account row
type AccountValidationError struct {
	Row         int    `json:"row"`
	AccountCode string `json:"account_code"`
	Field       string `json:"field"`
	Error       string `json:"error"`
	Value       string `json:"value"`
}

// AccountImportResult represents the result of account import with validation details
type AccountImportResult struct {
	ValidAccounts    []Account               `json:"valid_accounts"`
	ValidationErrors []AccountValidationError `json:"validation_errors"`
	TotalRows        int                     `json:"total_rows"`
	ValidCount       int                     `json:"valid_count"`
	ErrorCount       int                     `json:"error_count"`
	ErrorReportPath  string                  `json:"error_report_path,omitempty"`
	ImportTime       time.Time               `json:"import_time"`
}