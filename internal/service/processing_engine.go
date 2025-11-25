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
		if account.Nature != "" {
			tx.AnalisaNatureAkun = &account.Nature
		} else if account.AccountName != "" {
			// If nature is empty, fallback to account name
			tx.AnalisaNatureAkun = &account.AccountName
		} else {
			// If both are empty, use account code instead
			tx.AnalisaNatureAkun = &tx.Account
		}
	}

	// STEP 2: Koreksi - Match keterangan with koreksi_rules
	koreksiValue := e.matchKoreksiRule(keterangan)
	if koreksiValue != "" {
		tx.Koreksi = &koreksiValue
	}

	// STEP 3: Obyek - Match keterangan with obyek_rules
	obyekValue := e.matchObyekRule(keterangan)
	if obyekValue != "" {
		tx.Obyek = &obyekValue
	}

	// STEP 4: Analisa Koreksi - Obyek
	if (tx.Koreksi != nil && *tx.Koreksi != "") && (tx.Obyek != nil && *tx.Obyek != "") {
		combinedValue := fmt.Sprintf("%s - %s", *tx.Koreksi, *tx.Obyek)
		tx.AnalisaKoreksiObyek = &combinedValue
	} else if tx.Koreksi != nil && *tx.Koreksi != "" {
		tx.AnalisaKoreksiObyek = tx.Koreksi
	} else if tx.Obyek != nil && *tx.Obyek != "" {
		tx.AnalisaKoreksiObyek = tx.Obyek
	}

	// STEP 5: Withholding Tax (21, 23, 26, 4.2, 15)
	e.calculateWithholdingTax(tx, keterangan)

	// STEP 6: PM DB (Input Tax)
	e.calculateInputTax(tx, keterangan)

	// STEP 7: PK CR (Output Tax)
	e.calculateOutputTax(tx, keterangan)

	// STEP 8: UM Pajak DB (TBD - not implemented yet)
	tx.UmPajakDB = models.NullableNumericFloat64{Value: 0.0, Valid: false}

	// STEP 9: Analisa Tambahan (TBD - not implemented yet)
	analisaTambahanValue := ""
	tx.AnalisaTambahan = &analisaTambahanValue

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
				tx.Wth21Cr = models.NullableNumericFloat64{Value: amount, Valid: true}
			case "wth_23":
				tx.Wth23Cr = models.NullableNumericFloat64{Value: amount, Valid: true}
			case "wth_26":
				tx.Wth26Cr = models.NullableNumericFloat64{Value: amount, Valid: true}
			case "wth_4_2":
				tx.Wth42Cr = models.NullableNumericFloat64{Value: amount, Valid: true}
			case "wth_15":
				tx.Wth15Cr = models.NullableNumericFloat64{Value: amount, Valid: true}
			}
		}
	}
}

// calculateInputTax calculates PM DB (Input Tax)
func (e *ProcessingEngine) calculateInputTax(tx *models.TransactionData, keterangan string) {
	if tx.Debet <= 0 {
		tx.PmDB = models.NullableNumericFloat64{Value: 0.0, Valid: false}
		return
	}

	// Check if keterangan contains input tax keyword
	for _, keyword := range e.inputTaxKeywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(keterangan, keywordLower) {
			tx.PmDB = models.NullableNumericFloat64{Value: tx.Debet, Valid: true}
			return
		}
	}

	tx.PmDB = models.NullableNumericFloat64{Value: 0.0, Valid: false}
}

// calculateOutputTax calculates PK CR (Output Tax)
func (e *ProcessingEngine) calculateOutputTax(tx *models.TransactionData, keterangan string) {
	if tx.Credit <= 0 {
		tx.PkCr = models.NullableNumericFloat64{Value: 0.0, Valid: false}
		return
	}

	// Check if keterangan contains output tax keyword
	for _, keyword := range e.outputTaxKeywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(keterangan, keywordLower) {
			tx.PkCr = models.NullableNumericFloat64{Value: tx.Credit, Valid: true}
			return
		}
	}

	tx.PkCr = models.NullableNumericFloat64{Value: 0.0, Valid: false}
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
			errMsg := err.Error()
			transactions[i].ProcessingError = &errMsg
			continue
		}
	}

	// Bulk update transactions
	return e.uploadRepo.BulkUpdateTransactions(transactions)
}
