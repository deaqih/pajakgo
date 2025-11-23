package models

import "time"

type AdditionalAnalysis struct {
	ID            int       `db:"id" json:"id"`
	AccountCode   string    `db:"account_code" json:"account_code"`
	Account       *Account  `db:"-" json:"account,omitempty"` // For JOIN operations
	AnalysisType  string    `db:"analysis_type" json:"analysis_type"` // revenue_recognition, tax_treatment, etc.
	AnalysisTitle string    `db:"analysis_title" json:"analysis_title"`
	Status        string    `db:"status" json:"status"` // active, inactive
	Notes         *string   `db:"notes" json:"notes"`
	CreatedBy     *int      `db:"created_by" json:"created_by"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type AdditionalAnalysisRequest struct {
	AccountCode   string `json:"account_code" validate:"required"`
	AnalysisType  string `json:"analysis_type" validate:"required"`
	AnalysisTitle string `json:"analysis_title" validate:"required"`
	Status        string `json:"status"`
	Notes         string `json:"notes"`
}

type AdditionalAnalysisResponse struct {
	ID          int       `json:"id" db:"id"`
	AccountCode string    `json:"account_code" db:"account_code"`
	AccountName string    `json:"account_name,omitempty" db:"account_name"`
	AnalysisType string   `json:"analysis_type" db:"analysis_type"`
	AnalysisTitle string  `json:"analysis_title" db:"analysis_title"`
	Status      string    `json:"status" db:"status"`
	Notes       *string   `json:"notes" db:"notes"`
	CreatedBy   *int      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AdditionalAnalysisImportResult struct {
	Success      int                          `json:"success"`
	Failed       int                          `json:"failed"`
	Total        int                          `json:"total"`
	Errors       []AdditionalAnalysisValidationError `json:"errors"`
	ProcessedAt  time.Time                    `json:"processed_at"`
}

type AdditionalAnalysisValidationError struct {
	Row       int                    `json:"row"`
	Field     string                 `json:"field"`
	Value     string                 `json:"value"`
	Message   string                 `json:"message"`
	Data      AdditionalAnalysisRequest `json:"data,omitempty"`
}

type AdditionalAnalysisFilter struct {
	AccountCode  string `json:"account_code,omitempty"`
	AnalysisType string `json:"analysis_type,omitempty"`
	Status       string `json:"status,omitempty"`
	Search       string `json:"search,omitempty"`
	Page         int    `json:"page,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	SortBy       string `json:"sort_by,omitempty"`
	SortOrder    string `json:"sort_order,omitempty"`
}

type AdditionalAnalysisExportRequest struct {
	AccountCode  string   `json:"account_code,omitempty"`
	AnalysisType string   `json:"analysis_type,omitempty"`
	Status       string   `json:"status,omitempty"`
	Search       string   `json:"search,omitempty"`
	Format       string   `json:"format"` // excel, csv
	Columns      []string `json:"columns"` // columns to export
}