package service

import (
	"accounting-web/internal/models"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

// ParseTransactionFile parses an Excel file and returns transaction data
func (s *ExcelService) ParseTransactionFile(filePath string) ([]models.TransactionData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header row and one data row")
	}

	// Validate header
	header := rows[0]
	expectedHeaders := []string{
		"Document Type", "Document Number", "Posting Date", "Account",
		"Account Name", "Keterangan", "Debet", "Credit", "Net",
	}

	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("invalid header format")
	}

	// Parse data rows
	var transactions []models.TransactionData
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 9 {
			continue // Skip incomplete rows
		}

		tx := models.TransactionData{}

		// Parse basic fields
		tx.DocumentType = getCellValue(row, 0)
		tx.DocumentNumber = getCellValue(row, 1)

		// Parse posting date
		dateStr := getCellValue(row, 2)
		if dateStr != "" {
			parsedDate, err := parseDate(dateStr)
			if err == nil {
				tx.PostingDate = parsedDate
			}
		}

		tx.Account = getCellValue(row, 3)
		tx.AccountName = getCellValue(row, 4)
		tx.Keterangan = getCellValue(row, 5)

		// Parse numeric fields
		tx.Debet = parseFloat(getCellValue(row, 6))
		tx.Credit = parseFloat(getCellValue(row, 7))
		tx.Net = parseFloat(getCellValue(row, 8))

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// ExportTransactions exports processed transactions to Excel
func (s *ExcelService) ExportTransactions(transactions []models.TransactionData, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Processed Data"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Document Type", "Document Number", "Posting Date", "Account", "Account Name",
		"Keterangan", "Debet", "Credit", "Net",
		"Analisa Nature Akun", "Koreksi", "Obyek", "Analisa Koreksi-Obyek",
		"UM Pajak DB", "PM DB", "WHT 21 CR", "WHT 23 CR", "WHT 26 CR", "WHT 4.2 CR",
		"WHT 15 CR", "PK CR", "Analisa Tambahan",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for rowIdx, tx := range transactions {
		row := rowIdx + 2
		values := []interface{}{
			tx.DocumentType,
			tx.DocumentNumber,
			tx.PostingDate.Format("2006-01-02"),
			tx.Account,
			tx.AccountName,
			tx.Keterangan,
			tx.Debet,
			tx.Credit,
			tx.Net,
			tx.AnalisaNatureAkun,
			tx.Koreksi,
			tx.Obyek,
			tx.AnalisaKoreksiObyek,
			tx.UmPajakDB,
			tx.PmDB,
			tx.Wth21Cr,
			tx.Wth23Cr,
			tx.Wth26Cr,
			tx.Wth42Cr,
			tx.Wth15Cr,
			tx.PkCr,
			tx.AnalisaTambahan,
		}

		for colIdx, value := range values {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Save file
	return f.SaveAs(outputPath)
}

// Helper functions
func getCellValue(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"02-01-2006",
		"01/02/2006",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func getColumnName(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+(index%26))) + result
		index = index/26 - 1
	}
	return result
}

// ExportAccounts exports accounts to Excel file
func (s *ExcelService) ExportAccounts(accounts []models.Account, filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Accounts"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Account Code", "Account Name", "Account Type", "Nature",
		"Koreksi Obyek", "Analisa Tambahan", "Is Active",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write data
	for i, account := range accounts {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), account.AccountCode)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), account.AccountName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), account.AccountType)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), account.Nature)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), account.KoreksiObyek)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), account.AnalisaTambahan)

		isActiveStr := "No"
		if account.IsActive {
			isActiveStr = "Yes"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), isActiveStr)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 20)
	f.SetColWidth(sheetName, "G", "G", 12)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(filePath)
}

// ParseAccountsFile parses an Excel file and returns account data
func (s *ExcelService) ParseAccountsFile(filePath string) ([]models.Account, error) {
	result, err := s.ParseAccountsWithValidation(filePath)
	if err != nil {
		return nil, err
	}
	return result.ValidAccounts, nil
}

// ParseAccountsWithValidation parses an Excel file and returns detailed validation result
func (s *ExcelService) ParseAccountsWithValidation(filePath string) (*models.AccountImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header row and one data row")
	}

	// Validate headers
	expectedHeaders := []string{
		"Account Code", "Account Name", "Account Type", "Nature",
		"Koreksi Obyek", "Analisa Tambahan", "Is Active",
	}

	header := rows[0]
	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("invalid header format. Expected columns: %v", expectedHeaders)
	}

	result := &models.AccountImportResult{
		ValidAccounts:    []models.Account{},
		ValidationErrors: []models.AccountValidationError{},
		TotalRows:        len(rows) - 1, // Exclude header
		ValidCount:       0,
		ErrorCount:       0,
		ImportTime:       time.Now(),
	}

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Skip completely empty rows
		if len(row) == 0 || (len(row) > 0 && row[0] == "") {
			continue
		}

		// Extract values
		accountCode := getStringValue(row, 0)
		accountName := getStringValue(row, 1)
		accountType := getStringValue(row, 2)
		nature := getStringValue(row, 3)
		koreksiObyek := getStringValue(row, 4)
		analisaTambahan := getStringValue(row, 5)
		isActiveStr := getStringValue(row, 6)

		// Validate fields
		rowErrors := s.validateAccountRow(i+1, accountCode, accountName, accountType, nature, koreksiObyek, analisaTambahan, isActiveStr)

		if len(rowErrors) > 0 {
			result.ValidationErrors = append(result.ValidationErrors, rowErrors...)
			result.ErrorCount++
		} else {
			// Create valid account
			account := models.Account{
				AccountCode:     accountCode,
				AccountName:     accountName,
				AccountType:     accountType,
				Nature:          nature,
				KoreksiObyek:    koreksiObyek,
				AnalisaTambahan: analisaTambahan,
				IsActive:        parseBoolValue(isActiveStr),
			}
			result.ValidAccounts = append(result.ValidAccounts, account)
			result.ValidCount++
		}
	}

	return result, nil
}

// validateAccountRow validates a single account row and returns validation errors
func (s *ExcelService) validateAccountRow(rowNum int, accountCode, accountName, accountType, nature, koreksiObyek, analisaTambahan, isActiveStr string) []models.AccountValidationError {
	var errors []models.AccountValidationError

	// Validate Account Code (Required)
	if accountCode == "" {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Account Code",
			Error:       "Account Code is required",
			Value:       accountCode,
		})
	} else if len(accountCode) > 50 {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Account Code",
			Error:       "Account Code cannot exceed 50 characters",
			Value:       accountCode,
		})
	}

	// Validate Account Name (Required)
	if accountName == "" {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Account Name",
			Error:       "Account Name is required",
			Value:       accountName,
		})
	} else if len(accountName) > 200 {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Account Name",
			Error:       "Account Name cannot exceed 200 characters",
			Value:       accountName,
		})
	}

	// Validate Account Type (Optional, max 100 chars)
	if len(accountType) > 100 {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Account Type",
			Error:       "Account Type cannot exceed 100 characters",
			Value:       accountType,
		})
	}

	// Validate Nature (Optional, must be valid value)
	validNatures := []string{"Asset", "Liability", "Equity", "Revenue", "Expense", ""}
	if nature != "" && !contains(validNatures, nature) {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Nature",
			Error:       fmt.Sprintf("Nature must be one of: %v", validNatures),
			Value:       nature,
		})
	}

	// Validate Koreksi Obyek (Optional, max 100 chars)
	if len(koreksiObyek) > 100 {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Koreksi Obyek",
			Error:       "Koreksi Obyek cannot exceed 100 characters",
			Value:       koreksiObyek,
		})
	}

	// Validate Analisa Tambahan (Optional, max 200 chars)
	if len(analisaTambahan) > 200 {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Analisa Tambahan",
			Error:       "Analisa Tambahan cannot exceed 200 characters",
			Value:       analisaTambahan,
		})
	}

	// Validate Is Active (Optional, must be boolean-like)
	if isActiveStr != "" && !parseBoolValue(isActiveStr) && !isBooleanLike(isActiveStr) {
		errors = append(errors, models.AccountValidationError{
			Row:         rowNum,
			AccountCode: accountCode,
			Field:       "Is Active",
			Error:       "Is Active must be Yes/No, Y/N, 1/0, or true/false",
			Value:       isActiveStr,
		})
	}

	return errors
}

// GenerateImportErrorReport creates an Excel report with import validation errors
func (s *ExcelService) GenerateImportErrorReport(result *models.AccountImportResult, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Import Errors"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Row Number", "Account Code", "Field", "Error Message", "Invalid Value",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFE6E6"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write error data
	for rowIdx, error := range result.ValidationErrors {
		row := rowIdx + 2
		values := []interface{}{
			error.Row,
			error.AccountCode,
			error.Field,
			error.Error,
			error.Value,
		}

		for colIdx, value := range values {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f.SetCellValue(sheetName, cell, value)
		}

		// Set error row style
		errorStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFFCC"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%s%d", getColumnName(len(headers)-1), row), errorStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 50)
	f.SetColWidth(sheetName, "E", "E", 25)

	// Add summary section
	summaryStartRow := len(result.ValidationErrors) + 4
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow), "Import Summary")
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+1), "Total Rows Processed:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+1), result.TotalRows)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+2), "Valid Accounts:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+2), result.ValidCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+3), "Errors Found:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+3), result.ErrorCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+4), "Success Rate:")
	successRate := float64(result.ValidCount) / float64(result.TotalRows) * 100
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+4), fmt.Sprintf("%.1f%%", successRate))

	// Style summary section
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryStartRow), fmt.Sprintf("A%d", summaryStartRow), summaryStyle)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(outputPath)
}

// Helper function to check if a string contains in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to check if a string looks like a boolean value
func isBooleanLike(s string) bool {
	booleanValues := []string{"yes", "no", "y", "n", "1", "0", "true", "false", "YES", "NO", "Y", "N", "TRUE", "FALSE"}
	for _, val := range booleanValues {
		if s == val {
			return true
		}
	}
	return false
}

func getStringValue(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func parseBoolValue(s string) bool {
	s = fmt.Sprintf("%v", s)
	return s == "Yes" || s == "yes" || s == "Y" || s == "y" || s == "1" || s == "true" || s == "TRUE"
}

// ExportKoreksiRules exports koreksi rules to Excel file
func (s *ExcelService) ExportKoreksiRules(rules []models.KoreksiRule, filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Koreksi Rules"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Keyword", "Value", "Is Active",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write data
	for i, rule := range rules {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rule.Keyword)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rule.Value)

		isActiveStr := "No"
		if rule.IsActive {
			isActiveStr = "Yes"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), isActiveStr)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 30)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 12)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(filePath)
}

// ParseKoreksiRulesWithValidation parses an Excel file and returns detailed validation result
func (s *ExcelService) ParseKoreksiRulesWithValidation(filePath string) (*models.KoreksiRuleImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header row and one data row")
	}

	// Validate headers
	expectedHeaders := []string{
		"Keyword", "Value", "Is Active",
	}

	header := rows[0]
	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("invalid header format. Expected columns: %v", expectedHeaders)
	}

	result := &models.KoreksiRuleImportResult{
		ValidRules:       []models.KoreksiRule{},
		ValidationErrors: []models.KoreksiRuleValidationError{},
		TotalRows:        len(rows) - 1, // Exclude header
		ValidCount:       0,
		ErrorCount:       0,
		ImportTime:       time.Now(),
	}

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Skip completely empty rows
		if len(row) == 0 || (len(row) > 0 && row[0] == "") {
			continue
		}

		// Extract values
		keyword := getStringValue(row, 0)
		value := getStringValue(row, 1)
		isActiveStr := getStringValue(row, 2)

		// Validate fields
		rowErrors := s.validateKoreksiRuleRow(i+1, keyword, value, isActiveStr)

		if len(rowErrors) > 0 {
			result.ValidationErrors = append(result.ValidationErrors, rowErrors...)
			result.ErrorCount++
		} else {
			// Create valid rule
			rule := models.KoreksiRule{
				Keyword:  keyword,
				Value:    value,
				IsActive: parseBoolValue(isActiveStr),
			}
			result.ValidRules = append(result.ValidRules, rule)
			result.ValidCount++
		}
	}

	return result, nil
}

// validateKoreksiRuleRow validates a single koreksi rule row and returns validation errors
func (s *ExcelService) validateKoreksiRuleRow(rowNum int, keyword, value, isActiveStr string) []models.KoreksiRuleValidationError {
	var errors []models.KoreksiRuleValidationError

	// Validate Keyword (Required)
	if keyword == "" {
		errors = append(errors, models.KoreksiRuleValidationError{
			Row:     rowNum,
			Field:   "Keyword",
			Value:   keyword,
			Message: "Keyword is required",
		})
	} else if len(keyword) > 255 {
		errors = append(errors, models.KoreksiRuleValidationError{
			Row:     rowNum,
			Field:   "Keyword",
			Value:   keyword,
			Message: "Keyword cannot exceed 255 characters",
		})
	}

	// Validate Value (Required)
	if value == "" {
		errors = append(errors, models.KoreksiRuleValidationError{
			Row:     rowNum,
			Field:   "Value",
			Value:   value,
			Message: "Value is required",
		})
	} else if len(value) > 255 {
		errors = append(errors, models.KoreksiRuleValidationError{
			Row:     rowNum,
			Field:   "Value",
			Value:   value,
			Message: "Value cannot exceed 255 characters",
		})
	}

	// Validate Is Active (Optional, must be boolean-like)
	if isActiveStr != "" && !parseBoolValue(isActiveStr) && !isBooleanLike(isActiveStr) {
		errors = append(errors, models.KoreksiRuleValidationError{
			Row:     rowNum,
			Field:   "Is Active",
			Value:   isActiveStr,
			Message: "Is Active must be Yes/No, Y/N, 1/0, or true/false",
		})
	}

	return errors
}

// GenerateKoreksiRuleImportErrorReport creates an Excel report with import validation errors
func (s *ExcelService) GenerateKoreksiRuleImportErrorReport(result *models.KoreksiRuleImportResult, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Import Errors"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Row Number", "Field", "Error Message", "Invalid Value",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFE6E6"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write error data
	for rowIdx, error := range result.ValidationErrors {
		row := rowIdx + 2
		values := []interface{}{
			error.Row,
			error.Field,
			error.Message,
			error.Value,
		}

		for colIdx, value := range values {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f.SetCellValue(sheetName, cell, value)
		}

		// Set error row style
		errorStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFFCC"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%s%d", getColumnName(len(headers)-1), row), errorStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 50)
	f.SetColWidth(sheetName, "D", "D", 25)

	// Add summary section
	summaryStartRow := len(result.ValidationErrors) + 4
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow), "Import Summary")
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+1), "Total Rows Processed:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+1), result.TotalRows)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+2), "Valid Rules:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+2), result.ValidCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+3), "Errors Found:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+3), result.ErrorCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+4), "Success Rate:")
	successRate := float64(result.ValidCount) / float64(result.TotalRows) * 100
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+4), fmt.Sprintf("%.1f%%", successRate))

	// Style summary section
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryStartRow), fmt.Sprintf("A%d", summaryStartRow), summaryStyle)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(outputPath)
}

// ExportObyekRules exports obyek rules to Excel file
func (s *ExcelService) ExportObyekRules(rules []models.ObyekRule, filePath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Obyek Rules"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Keyword", "Value", "Is Active",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write data
	for i, rule := range rules {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rule.Keyword)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rule.Value)

		isActiveStr := "No"
		if rule.IsActive {
			isActiveStr = "Yes"
		}
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), isActiveStr)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 30)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 12)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(filePath)
}

// ParseObyekRulesWithValidation parses an Excel file and returns detailed validation result
func (s *ExcelService) ParseObyekRulesWithValidation(filePath string) (*models.ObyekRuleImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header row and one data row")
	}

	// Validate headers
	expectedHeaders := []string{
		"Keyword", "Value", "Is Active",
	}

	header := rows[0]
	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("invalid header format. Expected columns: %v", expectedHeaders)
	}

	result := &models.ObyekRuleImportResult{
		ValidRules:       []models.ObyekRule{},
		ValidationErrors: []models.ObyekRuleValidationError{},
		TotalRows:        len(rows) - 1,
		ValidCount:       0,
		ErrorCount:       0,
		ImportTime:       time.Now(),
	}

	// Process data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Skip completely empty rows
		if len(row) == 0 || (len(row) > 0 && row[0] == "") {
			continue
		}

		// Extract values
		keyword := getStringValue(row, 0)
		value := getStringValue(row, 1)
		isActiveStr := getStringValue(row, 2)

		// Validate fields
		rowErrors := s.validateObyekRuleRow(i+1, keyword, value, isActiveStr)

		if len(rowErrors) > 0 {
			result.ValidationErrors = append(result.ValidationErrors, rowErrors...)
			result.ErrorCount++
		} else {
			// Create valid rule
			rule := models.ObyekRule{
				Keyword:  keyword,
				Value:    value,
				IsActive: parseBoolValue(isActiveStr),
			}
			result.ValidRules = append(result.ValidRules, rule)
			result.ValidCount++
		}
	}

	return result, nil
}

// validateObyekRuleRow validates a single obyek rule row and returns validation errors
func (s *ExcelService) validateObyekRuleRow(rowNum int, keyword, value, isActiveStr string) []models.ObyekRuleValidationError {
	var errors []models.ObyekRuleValidationError

	// Validate Keyword (Required)
	if keyword == "" {
		errors = append(errors, models.ObyekRuleValidationError{
			Row:     rowNum,
			Field:   "Keyword",
			Value:   keyword,
			Message: "Keyword is required",
		})
	} else if len(keyword) > 255 {
		errors = append(errors, models.ObyekRuleValidationError{
			Row:     rowNum,
			Field:   "Keyword",
			Value:   keyword,
			Message: "Keyword cannot exceed 255 characters",
		})
	}

	// Validate Value (Required)
	if value == "" {
		errors = append(errors, models.ObyekRuleValidationError{
			Row:     rowNum,
			Field:   "Value",
			Value:   value,
			Message: "Value is required",
		})
	} else if len(value) > 255 {
		errors = append(errors, models.ObyekRuleValidationError{
			Row:     rowNum,
			Field:   "Value",
			Value:   value,
			Message: "Value cannot exceed 255 characters",
		})
	}

	// Validate Is Active (Optional, must be boolean-like)
	if isActiveStr != "" && !parseBoolValue(isActiveStr) && !isBooleanLike(isActiveStr) {
		errors = append(errors, models.ObyekRuleValidationError{
			Row:     rowNum,
			Field:   "Is Active",
			Value:   isActiveStr,
			Message: "Is Active must be Yes/No, Y/N, 1/0, or true/false",
		})
	}

	return errors
}

// GenerateObyekRuleImportErrorReport creates an Excel report with import validation errors
func (s *ExcelService) GenerateObyekRuleImportErrorReport(result *models.ObyekRuleImportResult, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Import Errors"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Set headers
	headers := []string{
		"Row Number", "Field", "Error Message", "Invalid Value",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", getColumnName(i))
		f.SetCellValue(sheetName, cell, header)
	}

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFE6E6"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", getColumnName(len(headers)-1)), headerStyle)

	// Write error data
	for rowIdx, error := range result.ValidationErrors {
		row := rowIdx + 2
		values := []interface{}{
			error.Row,
			error.Field,
			error.Message,
			error.Value,
		}

		for colIdx, value := range values {
			cell := fmt.Sprintf("%s%d", getColumnName(colIdx), row)
			f.SetCellValue(sheetName, cell, value)
		}

		// Set error row style
		errorStyle, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#FFFFCC"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%s%d", getColumnName(len(headers)-1), row), errorStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 12)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 50)
	f.SetColWidth(sheetName, "D", "D", 25)

	// Add summary section
	summaryStartRow := len(result.ValidationErrors) + 4
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow), "Import Summary")
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+1), "Total Rows Processed:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+1), result.TotalRows)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+2), "Valid Rules:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+2), result.ValidCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+3), "Errors Found:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+3), result.ErrorCount)
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryStartRow+4), "Success Rate:")
	successRate := float64(result.ValidCount) / float64(result.TotalRows) * 100
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryStartRow+4), fmt.Sprintf("%.1f%%", successRate))

	// Style summary section
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryStartRow), fmt.Sprintf("A%d", summaryStartRow), summaryStyle)

	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	return f.SaveAs(outputPath)
}
