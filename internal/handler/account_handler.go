package handler

import (
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	accountRepo  *repository.AccountRepository
	excelService *service.ExcelService
}

func NewAccountHandler(accountRepo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{
		accountRepo:  accountRepo,
		excelService: service.NewExcelService(),
	}
}

func (h *AccountHandler) GetAccounts(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	accounts, total, err := h.accountRepo.FindAll(params.Limit, offset, params.Search)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve accounts", err)
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(total))

	responseData := fiber.Map{
		"accounts": accounts,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Accounts retrieved successfully", responseData, pagination)
}

func (h *AccountHandler) GetAccount(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid account ID", err)
	}

	account, err := h.accountRepo.FindByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Account not found", err)
	}

	return utils.SuccessResponse(c, "Account retrieved successfully", account)
}

func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	var req models.AccountRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validation
	if req.AccountCode == "" || req.AccountName == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Account code and name are required", nil)
	}

	account := &models.Account{
		AccountCode:     req.AccountCode,
		AccountName:     req.AccountName,
		AccountType:     req.AccountType,
		Nature:          req.Nature,
		KoreksiObyek:    req.KoreksiObyek,
		AnalisaTambahan: req.AnalisaTambahan,
		IsActive:        true, // Default to active
	}

	if err := h.accountRepo.Create(account); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create account", err)
	}

	return utils.SuccessResponse(c, "Account created successfully", account)
}

func (h *AccountHandler) UpdateAccount(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid account ID", err)
	}

	var req models.AccountRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validation
	if req.AccountCode == "" || req.AccountName == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Account code and name are required", nil)
	}

	account, err := h.accountRepo.FindByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Account not found", err)
	}

	account.AccountCode = req.AccountCode
	account.AccountName = req.AccountName
	account.AccountType = req.AccountType
	account.Nature = req.Nature
	account.KoreksiObyek = req.KoreksiObyek
	account.AnalisaTambahan = req.AnalisaTambahan
	account.IsActive = req.IsActive

	if err := h.accountRepo.Update(account); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update account", err)
	}

	return utils.SuccessResponse(c, "Account updated successfully", account)
}

func (h *AccountHandler) DeleteAccount(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid account ID", err)
	}

	if err := h.accountRepo.Delete(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete account", err)
	}

	return utils.SuccessResponse(c, "Account deleted successfully", nil)
}

func (h *AccountHandler) ExportAccounts(c *fiber.Ctx) error {
	// Get all accounts
	accounts, err := h.accountRepo.GetAllActive()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve accounts", err)
	}

	// Generate export filename
	exportFileName := fmt.Sprintf("accounts_export_%s.xlsx", time.Now().Format("20060102_150405"))
	exportPath := filepath.Join("./storage/exports", exportFileName)

	// Export to Excel
	if err := h.excelService.ExportAccounts(accounts, exportPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export accounts", err)
	}

	// Send file
	return c.Download(exportPath, exportFileName)
}

func (h *AccountHandler) DownloadTemplate(c *fiber.Ctx) error {
	// Generate template filename
	templateFileName := "accounts_import_template.xlsx"
	templatePath := filepath.Join("./storage/exports", templateFileName)

	// Create template with sample data
	sampleAccounts := []models.Account{
		{
			AccountCode:     "1000",
			AccountName:     "Cash",
			AccountType:     "Current Asset",
			Nature:          "Asset",
			KoreksiObyek:    "N/A",
			AnalisaTambahan: "Bank Account",
			IsActive:        true,
		},
		{
			AccountCode:     "2000",
			AccountName:     "Accounts Payable",
			AccountType:     "Current Liability",
			Nature:          "Liability",
			KoreksiObyek:    "Standard",
			AnalisaTambahan: "Trade Payable",
			IsActive:        true,
		},
	}

	// Export template
	if err := h.excelService.ExportAccounts(sampleAccounts, templatePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate template", err)
	}

	// Send file
	return c.Download(templatePath, templateFileName)
}

func (h *AccountHandler) ImportAccounts(c *fiber.Ctx) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File is required", err)
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	if ext != ".xlsx" && ext != ".xls" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Only Excel files (.xlsx, .xls) are allowed", nil)
	}

	// Save file temporarily
	tempPath := filepath.Join("./storage/temp", fmt.Sprintf("import_%d%s", time.Now().Unix(), ext))
	if err := c.SaveFile(file, tempPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save file", err)
	}

	// Parse Excel file with validation
	result, err := h.excelService.ParseAccountsWithValidation(tempPath)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to parse Excel file: "+err.Error(), err)
	}

	// Clean up temp file
	defer os.Remove(tempPath)

	// If there are no valid accounts
	if result.ValidCount == 0 {
		// Generate error report
		errorReportPath := ""
		if len(result.ValidationErrors) > 0 {
			errorReportPath = filepath.Join("./storage/exports", fmt.Sprintf("import_errors_%s.xlsx", time.Now().Format("20060102_150405")))
			if err := h.excelService.GenerateImportErrorReport(result, errorReportPath); err == nil {
				result.ErrorReportPath = errorReportPath
			}
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":           false,
			"message":           "No valid accounts found in the file",
			"total_rows":        result.TotalRows,
			"valid_count":       result.ValidCount,
			"error_count":       result.ErrorCount,
			"errors":            result.ValidationErrors,
			"error_report_path": result.ErrorReportPath,
		})
	}

	// If there are validation errors but some valid accounts, import the valid ones
	if len(result.ValidationErrors) > 0 {
		// Generate error report
		errorReportPath := filepath.Join("./storage/exports", fmt.Sprintf("import_errors_%s.xlsx", time.Now().Format("20060102_150405")))
		if err := h.excelService.GenerateImportErrorReport(result, errorReportPath); err == nil {
			result.ErrorReportPath = errorReportPath
		}

		// Import only valid accounts
		if err := h.accountRepo.BulkInsert(result.ValidAccounts); err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import valid accounts: "+err.Error(), err)
		}

		// Return partial success with error details
		return c.Status(fiber.StatusPartialContent).JSON(fiber.Map{
			"success":            true,
			"message":            fmt.Sprintf("Import completed with %d errors. %d accounts imported successfully.", result.ErrorCount, result.ValidCount),
			"total_rows":         result.TotalRows,
			"valid_count":        result.ValidCount,
			"error_count":        result.ErrorCount,
			"errors":             getFirstNErrors(result.ValidationErrors, 10), // Limit to first 10 errors for readability
			"error_report_path":  result.ErrorReportPath,
			"total_imported":     result.ValidCount,
		})
	}

	// If no validation errors, import all accounts
	if err := h.accountRepo.BulkInsert(result.ValidAccounts); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import accounts: "+err.Error(), err)
	}

	return utils.SuccessResponse(c, "All accounts imported successfully", fiber.Map{
		"total_rows":     result.TotalRows,
		"valid_count":    result.ValidCount,
		"error_count":    result.ErrorCount,
		"total_imported": result.ValidCount,
	})
}

// DownloadErrorReport downloads an error report file
func (h *AccountHandler) DownloadErrorReport(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Filename is required", nil)
	}

	// Validate filename to prevent directory traversal
	if !isValidFilename(filename) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid filename", nil)
	}

	filePath := filepath.Join("./storage/exports", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Error report file not found", err)
	}

	// Send file
	return c.Download(filePath, filename)
}

// getFirstNErrors returns the first n errors from a slice
func getFirstNErrors(errors []models.AccountValidationError, n int) []models.AccountValidationError {
	if len(errors) <= n {
		return errors
	}
	return errors[:n]
}

// isValidFilename validates filename to prevent directory traversal
func isValidFilename(filename string) bool {
	// Basic validation - no path separators, no special chars
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	// Check for dangerous characters
	dangerousChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Check if it starts with "import_errors_" (our expected format)
	return strings.HasPrefix(filename, "import_errors_") && strings.HasSuffix(filename, ".xlsx")
}
