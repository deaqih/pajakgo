package service

import (
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"fmt"
	"strings"
)

type ProcessingEngine struct {
	accountRepo *repository.AccountRepository
	rulesRepo   *repository.RulesRepository
	uploadRepo  *repository.UploadRepository

	// Cached rules
	accounts       map[string]models.Account
	koreksiRules   []models.KoreksiRule
	obyekRules     []models.ObyekRule
	whtRules       []models.WithholdingTaxRule
	taxKeywords    []models.TaxKeyword
	inputTaxKeywords  []string
	outputTaxKeywords []string
}

func NewProcessingEngine(
	accountRepo *repository.AccountRepository,
	rulesRepo *repository.RulesRepository,
	uploadRepo *repository.UploadRepository,
) *ProcessingEngine {
	return &ProcessingEngine{
		accountRepo: accountRepo,
		rulesRepo:   rulesRepo,
		uploadRepo:  uploadRepo,
	}
}

// LoadRules loads all active rules into memory for processing
func (e *ProcessingEngine) LoadRules() error {
	// Load accounts
	accounts, err := e.accountRepo.GetAllActive()
	if err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}
	e.accounts = make(map[string]models.Account)
	for _, acc := range accounts {
		e.accounts[acc.AccountCode] = acc
	}

	// Load koreksi rules
	e.koreksiRules, err = e.rulesRepo.GetActiveKoreksiRules()
	if err != nil {
		return fmt.Errorf("failed to load koreksi rules: %w", err)
	}

	// Load obyek rules
	e.obyekRules, err = e.rulesRepo.GetActiveObyekRules()
	if err != nil {
		return fmt.Errorf("failed to load obyek rules: %w", err)
	}

	// Load withholding tax rules
	e.whtRules, err = e.rulesRepo.GetActiveWithholdingTaxRules()
	if err != nil {
		return fmt.Errorf("failed to load WHT rules: %w", err)
	}

	// Load tax keywords
	taxKeywords, err := e.rulesRepo.GetActiveTaxKeywords()
	if err != nil {
		return fmt.Errorf("failed to load tax keywords: %w", err)
	}
	e.taxKeywords = taxKeywords

	// Separate input and output tax keywords
	e.inputTaxKeywords = []string{}
	e.outputTaxKeywords = []string{}
	for _, kw := range taxKeywords {
		if kw.TaxCategory == "input_tax" {
			e.inputTaxKeywords = append(e.inputTaxKeywords, kw.Keyword)
		} else if kw.TaxCategory == "output_tax" {
			e.outputTaxKeywords = append(e.outputTaxKeywords, kw.Keyword)
		}
	}

	return nil
}

// ProcessTransaction processes a single transaction according to PRD rules
func (e *ProcessingEngine) ProcessTransaction(tx *models.TransactionData) error {
	keterangan := strings.ToLower(tx.Keterangan)

	// STEP 1: Analisa Nature Akun
	if account, exists := e.accounts[tx.Account]; exists {
		tx.AnalisaNatureAkun = account.Nature
	}

	// STEP 2: Koreksi - Match keterangan with koreksi_rules
	tx.Koreksi = e.matchKoreksiRule(keterangan)

	// STEP 3: Obyek - Match keterangan with obyek_rules
	tx.Obyek = e.matchObyekRule(keterangan)

	// STEP 4: Analisa Koreksi - Obyek
	if tx.Koreksi != "" && tx.Obyek != "" {
		tx.AnalisaKoreksiObyek = fmt.Sprintf("%s - %s", tx.Koreksi, tx.Obyek)
	} else if tx.Koreksi != "" {
		tx.AnalisaKoreksiObyek = tx.Koreksi
	} else if tx.Obyek != "" {
		tx.AnalisaKoreksiObyek = tx.Obyek
	}

	// STEP 5: Withholding Tax (21, 23, 26, 4.2, 15)
	e.calculateWithholdingTax(tx, keterangan)

	// STEP 6: PM DB (Input Tax)
	e.calculateInputTax(tx, keterangan)

	// STEP 7: PK CR (Output Tax)
	e.calculateOutputTax(tx, keterangan)

	// STEP 8: UM Pajak DB (TBD - not implemented yet)
	tx.UmPajakDB = 0

	// STEP 9: Analisa Tambahan (TBD - not implemented yet)
	tx.AnalisaTambahan = ""

	// Mark as processed
	tx.IsProcessed = true

	return nil
}

// matchKoreksiRule finds the first matching koreksi rule based on priority
func (e *ProcessingEngine) matchKoreksiRule(keterangan string) string {
	for _, rule := range e.koreksiRules {
		keyword := strings.ToLower(rule.Keyword)
		if strings.Contains(keterangan, keyword) {
			return rule.Value
		}
	}
	return ""
}

// matchObyekRule finds the first matching obyek rule based on priority
func (e *ProcessingEngine) matchObyekRule(keterangan string) string {
	for _, rule := range e.obyekRules {
		keyword := strings.ToLower(rule.Keyword)
		if strings.Contains(keterangan, keyword) {
			return rule.Value
		}
	}
	return ""
}

// calculateWithholdingTax calculates all withholding taxes if credit > 0
func (e *ProcessingEngine) calculateWithholdingTax(tx *models.TransactionData, keterangan string) {
	if tx.Credit <= 0 {
		return
	}

	// Check each WHT rule
	for _, rule := range e.whtRules {
		keyword := strings.ToLower(rule.Keyword)
		if strings.Contains(keterangan, keyword) {
			amount := tx.Credit * rule.TaxRate

			switch rule.TaxType {
			case "wth_21":
				tx.Wth21Cr = amount
			case "wth_23":
				tx.Wth23Cr = amount
			case "wth_26":
				tx.Wth26Cr = amount
			case "wth_4_2":
				tx.Wth42Cr = amount
			case "wth_15":
				tx.Wth15Cr = amount
			}
		}
	}
}

// calculateInputTax calculates PM DB (Input Tax)
func (e *ProcessingEngine) calculateInputTax(tx *models.TransactionData, keterangan string) {
	if tx.Debet <= 0 {
		tx.PmDB = 0
		return
	}

	// Check if keterangan contains input tax keyword
	for _, keyword := range e.inputTaxKeywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(keterangan, keywordLower) {
			tx.PmDB = tx.Debet
			return
		}
	}

	tx.PmDB = 0
}

// calculateOutputTax calculates PK CR (Output Tax)
func (e *ProcessingEngine) calculateOutputTax(tx *models.TransactionData, keterangan string) {
	if tx.Credit <= 0 {
		tx.PkCr = 0
		return
	}

	// Check if keterangan contains output tax keyword
	for _, keyword := range e.outputTaxKeywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(keterangan, keywordLower) {
			tx.PkCr = tx.Credit
			return
		}
	}

	tx.PkCr = 0
}

// ProcessBatch processes a batch of transactions
func (e *ProcessingEngine) ProcessBatch(transactions []models.TransactionData) error {
	// Load rules if not already loaded
	if e.koreksiRules == nil || e.obyekRules == nil {
		if err := e.LoadRules(); err != nil {
			return err
		}
	}

	// Process each transaction
	for i := range transactions {
		if err := e.ProcessTransaction(&transactions[i]); err != nil {
			transactions[i].ProcessingError = err.Error()
			continue
		}
	}

	// Bulk update transactions
	return e.uploadRepo.BulkUpdateTransactions(transactions)
}
