package models

import "time"

type Account struct {
	ID               int       `db:"id" json:"id"`
	AccountCode      string    `db:"account_code" json:"account_code"`
	AccountName      string    `db:"account_name" json:"account_name"`
	AccountType      string    `db:"account_type" json:"account_type"`
	Nature           string    `db:"nature" json:"nature"` // Asset, Liability, Equity, Revenue, Expense
	KoreksiObyek     string    `db:"koreksi_obyek" json:"koreksi_obyek"`
	AnalisaTambahan  string    `db:"analisa_tambahan" json:"analisa_tambahan"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type AccountRequest struct {
	AccountCode     string `json:"account_code" validate:"required"`
	AccountName     string `json:"account_name" validate:"required"`
	AccountType     string `json:"account_type"`
	Nature          string `json:"nature"`
	KoreksiObyek    string `json:"koreksi_obyek"`
	AnalisaTambahan string `json:"analisa_tambahan"`
	IsActive        bool   `json:"is_active"`
}
