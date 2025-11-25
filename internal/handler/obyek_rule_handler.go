package handler

import (
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ObyekRuleHandler struct {
	rulesRepo    *repository.RulesRepository
	excelService *service.ExcelService
}

func NewObyekRuleHandler(rulesRepo *repository.RulesRepository) *ObyekRuleHandler {
	return &ObyekRuleHandler{
		rulesRepo:    rulesRepo,
		excelService: service.NewExcelService(),
	}
}

func (h *ObyekRuleHandler) GetObyekRules(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	rules, total, err := h.rulesRepo.GetObyekRules(params.Limit, offset, params.Search)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve obyek rules", err)
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(total))

	responseData := fiber.Map{
		"rules":      rules,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Koreksi rules retrieved successfully", responseData, pagination)
}

func (h *ObyekRuleHandler) GetObyekRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	rule, err := h.rulesRepo.GetObyekRuleByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Koreksi rule not found", err)
	}

	return utils.SuccessResponse(c, "Koreksi rule retrieved successfully", rule)
}

func (h *ObyekRuleHandler) CreateObyekRule(c *fiber.Ctx) error {
	var req models.ObyekRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validation
	if req.Keyword == "" || req.Value == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Keyword and value are required", nil)
	}

	rule := &models.ObyekRule{
		Keyword:  req.Keyword,
		Value:    req.Value,
		NotValue: sql.NullString{String: req.NotValue, Valid: req.NotValue != ""},
		IsActive: true, // Default to active
	}

	if err := h.rulesRepo.CreateObyekRule(rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create obyek rule", err)
	}

	return utils.SuccessResponse(c, "Koreksi rule created successfully", rule)
}

func (h *ObyekRuleHandler) UpdateObyekRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var req models.ObyekRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validation
	if req.Keyword == "" || req.Value == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Keyword and value are required", nil)
	}

	rule, err := h.rulesRepo.GetObyekRuleByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Koreksi rule not found", err)
	}

	rule.Keyword = req.Keyword
	rule.Value = req.Value
	rule.NotValue = sql.NullString{String: req.NotValue, Valid: req.NotValue != ""}
	rule.IsActive = req.IsActive

	if err := h.rulesRepo.UpdateObyekRule(rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update obyek rule", err)
	}

	return utils.SuccessResponse(c, "Koreksi rule updated successfully", rule)
}

func (h *ObyekRuleHandler) DeleteObyekRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	if err := h.rulesRepo.DeleteObyekRule(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete obyek rule", err)
	}

	return utils.SuccessResponse(c, "Koreksi rule deleted successfully", nil)
}

func (h *ObyekRuleHandler) ExportObyekRules(c *fiber.Ctx) error {
	// Get all active rules
	rules, err := h.rulesRepo.GetAllActiveObyekRules()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve obyek rules", err)
	}

	// Generate export filename
	exportFileName := fmt.Sprintf("obyek_rules_export_%s.xlsx", time.Now().Format("20060102_150405"))
	exportPath := filepath.Join("./storage/exports", exportFileName)

	// Export to Excel
	if err := h.excelService.ExportObyekRules(rules, exportPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export obyek rules", err)
	}

	// Send file
	return c.Download(exportPath, exportFileName)
}

func (h *ObyekRuleHandler) DownloadTemplate(c *fiber.Ctx) error {
	// Generate template filename
	templateFileName := "obyek_rules_import_template.xlsx"
	templatePath := filepath.Join("./storage/exports", templateFileName)

	// Create template with sample data
	sampleRules := []models.ObyekRule{
		{
			Keyword:  "Rawat Inap",
			Value:    "Rawat Inap",
			NotValue: sql.NullString{String: "", Valid: false},
			IsActive: true,
		},
		{
			Keyword:  "Consultation",
			Value:    "Consultation Fee",
			NotValue: sql.NullString{String: "exclude", Valid: true},
			IsActive: true,
		},
	}

	// Export template
	if err := h.excelService.ExportObyekRules(sampleRules, templatePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate template", err)
	}

	// Send file
	return c.Download(templatePath, templateFileName)
}

func (h *ObyekRuleHandler) ImportObyekRules(c *fiber.Ctx) error {
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
	result, err := h.excelService.ParseObyekRulesWithValidation(tempPath)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to parse Excel file: "+err.Error(), err)
	}

	// Clean up temp file
	defer os.Remove(tempPath)

	// If there are no valid rules
	if result.ValidCount == 0 {
		// Generate error report
		errorReportPath := ""
		if len(result.ValidationErrors) > 0 {
			errorReportPath = filepath.Join("./storage/exports", fmt.Sprintf("import_errors_%s.xlsx", time.Now().Format("20060102_150405")))
			if err := h.excelService.GenerateObyekRuleImportErrorReport(result, errorReportPath); err == nil {
				result.ErrorReportPath = errorReportPath
			}
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":           false,
			"message":           "No valid obyek rules found in the file",
			"total_rows":        result.TotalRows,
			"valid_count":       result.ValidCount,
			"error_count":       result.ErrorCount,
			"errors":            result.ValidationErrors,
			"error_report_path": result.ErrorReportPath,
		})
	}

	// If there are validation errors but some valid rules, import the valid ones
	if len(result.ValidationErrors) > 0 {
		// Generate error report
		errorReportPath := filepath.Join("./storage/exports", fmt.Sprintf("import_errors_%s.xlsx", time.Now().Format("20060102_150405")))
		if err := h.excelService.GenerateObyekRuleImportErrorReport(result, errorReportPath); err == nil {
			result.ErrorReportPath = errorReportPath
		}

		// Import only valid rules
		if err := h.rulesRepo.BulkInsertObyekRules(result.ValidRules); err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import valid obyek rules: "+err.Error(), err)
		}

		// Return partial success with error details
		return c.Status(fiber.StatusPartialContent).JSON(fiber.Map{
			"success":            true,
			"message":            fmt.Sprintf("Import completed with %d errors. %d obyek rules imported successfully.", result.ErrorCount, result.ValidCount),
			"total_rows":         result.TotalRows,
			"valid_count":        result.ValidCount,
			"error_count":        result.ErrorCount,
			"errors":             getFirstNObyekRuleErrors(result.ValidationErrors, 10), // Limit to first 10 errors for readability
			"error_report_path":  result.ErrorReportPath,
			"total_imported":     result.ValidCount,
		})
	}

	// If no validation errors, import all rules
	if err := h.rulesRepo.BulkInsertObyekRules(result.ValidRules); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import obyek rules: "+err.Error(), err)
	}

	return utils.SuccessResponse(c, "All obyek rules imported successfully", fiber.Map{
		"total_rows":     result.TotalRows,
		"valid_count":    result.ValidCount,
		"error_count":    result.ErrorCount,
		"total_imported": result.ValidCount,
	})
}

// DownloadErrorReport downloads an error report file
func (h *ObyekRuleHandler) DownloadErrorReport(c *fiber.Ctx) error {
	filename := c.Params("filename")
	if filename == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Filename is required", nil)
	}

	// Validate filename to prevent directory traversal
	if !isValidObyekRuleFilename(filename) {
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

// getFirstNObyekRuleErrors returns the first n errors from a slice
func getFirstNObyekRuleErrors(errors []models.ObyekRuleValidationError, n int) []models.ObyekRuleValidationError {
	if len(errors) <= n {
		return errors
	}
	return errors[:n]
}

// isValidObyekRuleFilename validates filename to prevent directory traversal
func isValidObyekRuleFilename(filename string) bool {
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
