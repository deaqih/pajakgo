package models

import (
	"database/sql"
	"time"
)

type KoreksiRule struct {
	ID        int       `db:"id" json:"id"`
	Keyword   string    `db:"keyword" json:"keyword"`
	Value     string    `db:"value" json:"value"`
	NotValue  sql.NullString `db:"not_value" json:"not_value"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ObyekRule struct {
	ID        int       `db:"id" json:"id"`
	Keyword   string    `db:"keyword" json:"keyword"`
	Value     string    `db:"value" json:"value"`
	NotValue  sql.NullString `db:"not_value" json:"not_value"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type WithholdingTaxRule struct {
	ID        int       `db:"id" json:"id"`
	Keyword   string    `db:"keyword" json:"keyword"`
	TaxType   string    `db:"tax_type" json:"tax_type"`
	TaxRate   float64   `db:"tax_rate" json:"tax_rate"`
	Priority  int       `db:"priority" json:"priority"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type TaxKeyword struct {
	ID          int       `db:"id" json:"id"`
	Keyword     string    `db:"keyword" json:"keyword"`
	TaxCategory string    `db:"tax_category" json:"tax_category"`
	Priority    int       `db:"priority" json:"priority"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type RuleRequest struct {
	Keyword     string  `json:"keyword" validate:"required"`
	Value       string  `json:"value"`
	TaxType     string  `json:"tax_type"`
	TaxRate     float64 `json:"tax_rate"`
	TaxCategory string  `json:"tax_category"`
	Priority    int     `json:"priority"`
	IsActive    bool    `json:"is_active"`
}

type KoreksiRuleRequest struct {
	Keyword  string `json:"keyword" validate:"required"`
	Value    string `json:"value" validate:"required"`
	NotValue string `json:"not_value"`
	IsActive bool   `json:"is_active"`
}

type KoreksiRuleValidationError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type KoreksiRuleImportResult struct {
	ValidRules       []KoreksiRule                 `json:"valid_rules"`
	ValidationErrors []KoreksiRuleValidationError  `json:"validation_errors"`
	TotalRows        int                           `json:"total_rows"`
	ValidCount       int                           `json:"valid_count"`
	ErrorCount       int                           `json:"error_count"`
	ErrorReportPath  string                        `json:"error_report_path,omitempty"`
	ImportTime       time.Time                     `json:"import_time"`
}

type ObyekRuleRequest struct {
	Keyword  string `json:"keyword" validate:"required"`
	Value    string `json:"value" validate:"required"`
	NotValue string `json:"not_value"`
	IsActive bool   `json:"is_active"`
}

type ObyekRuleValidationError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type ObyekRuleImportResult struct {
	ValidRules       []ObyekRule                 `json:"valid_rules"`
	ValidationErrors []ObyekRuleValidationError  `json:"validation_errors"`
	TotalRows        int                         `json:"total_rows"`
	ValidCount       int                         `json:"valid_count"`
	ErrorCount       int                         `json:"error_count"`
	ErrorReportPath  string                      `json:"error_report_path,omitempty"`
	ImportTime       time.Time                   `json:"import_time"`
}
