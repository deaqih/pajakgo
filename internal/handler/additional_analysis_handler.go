package handler

import (
	"accounting-web/internal/models"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AdditionalAnalysisHandler struct {
	analysisService *service.AdditionalAnalysisService
}

func NewAdditionalAnalysisHandler(analysisService *service.AdditionalAnalysisService) *AdditionalAnalysisHandler {
	return &AdditionalAnalysisHandler{
		analysisService: analysisService,
	}
}

// Create creates a new additional analysis
func (h *AdditionalAnalysisHandler) Create(c *fiber.Ctx) error {
	var req models.AdditionalAnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id").(int)

	// Validate input
	if req.AccountCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Account code is required", nil)
	}
	if req.AnalysisType == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Analysis type is required", nil)
	}
	if req.AnalysisTitle == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Analysis title is required", nil)
	}
	
	// Create analysis
	analysis, err := h.analysisService.Create(req, userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create additional analysis", err)
	}

	return utils.SuccessResponse(c, "Additional analysis created successfully", analysis)
}

// GetByID retrieves an additional analysis by ID
func (h *AdditionalAnalysisHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID format", err)
	}

	analysis, err := h.analysisService.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Additional analysis not found", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get additional analysis", err)
	}

	return utils.SuccessResponse(c, "Additional analysis retrieved successfully", analysis)
}

// GetAll retrieves all additional analyses with filtering and pagination
func (h *AdditionalAnalysisHandler) GetAll(c *fiber.Ctx) error {
	// Parse query parameters
	filter := models.AdditionalAnalysisFilter{}

	// Parse string parameters
	filter.AccountCode = c.Query("account_code")
	filter.AnalysisType = c.Query("analysis_type")
	filter.Status = c.Query("status")
	filter.Search = c.Query("search")
	filter.SortBy = c.Query("sort_by")
	filter.SortOrder = c.Query("sort_order")

	// Parse integer parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	analyses, meta, err := h.analysisService.GetAll(filter)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get additional analyses", err)
	}

	// Combine data and pagination meta
	response := map[string]interface{}{
		"data":       analyses,
		"pagination": meta,
	}

	return utils.SuccessResponse(c, "Additional analyses retrieved successfully", response)
}

// Update updates an existing additional analysis
func (h *AdditionalAnalysisHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID format", err)
	}

	var req models.AdditionalAnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate input
	if req.AccountCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Account code is required", nil)
	}
	if req.AnalysisType == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Analysis type is required", nil)
	}
	if req.AnalysisTitle == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Analysis title is required", nil)
	}
	
	// Update analysis
	analysis, err := h.analysisService.Update(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Additional analysis not found", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update additional analysis", err)
	}

	return utils.SuccessResponse(c, "Additional analysis updated successfully", analysis)
}

// Delete soft deletes an additional analysis by ID
func (h *AdditionalAnalysisHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID format", err)
	}

	err = h.analysisService.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Additional analysis not found", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete additional analysis", err)
	}

	return utils.SuccessResponse(c, "Additional analysis deleted successfully", nil)
}

// HardDelete permanently deletes an additional analysis by ID
func (h *AdditionalAnalysisHandler) HardDelete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid ID format", err)
	}

	err = h.analysisService.HardDelete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Additional analysis not found", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to hard delete additional analysis", err)
	}

	return utils.SuccessResponse(c, "Additional analysis hard deleted successfully", nil)
}

// GetByAccountCode retrieves all additional analyses for a specific account
func (h *AdditionalAnalysisHandler) GetByAccountCode(c *fiber.Ctx) error {
	accountCode := c.Params("accountCode")
	if accountCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Account code is required", nil)
	}

	analyses, err := h.analysisService.GetByAccountCode(accountCode)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Account not found", err)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get additional analyses for account", err)
	}

	return utils.SuccessResponse(c, "Additional analyses retrieved successfully", analyses)
}

// GetAnalysisTypes retrieves all distinct analysis types
func (h *AdditionalAnalysisHandler) GetAnalysisTypes(c *fiber.Ctx) error {
	types, err := h.analysisService.GetAnalysisTypes()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get analysis types", err)
	}

	return utils.SuccessResponse(c, "Analysis types retrieved successfully", types)
}

// ImportFromExcel imports additional analyses from Excel file
func (h *AdditionalAnalysisHandler) ImportFromExcel(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "No file uploaded", err)
	}

	// Validate file type (should be Excel)
	if !strings.HasSuffix(file.Filename, ".xlsx") && !strings.HasSuffix(file.Filename, ".xls") {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid file type. Please upload Excel file", nil)
	}

	// TODO: Parse Excel file and convert to AdditionalAnalysisRequest array
	// For now, return placeholder
	analyses := []models.AdditionalAnalysisRequest{}

	result, err := h.analysisService.ImportFromExcel(analyses)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import additional analyses", err)
	}

	return utils.SuccessResponse(c, "Additional analyses imported successfully", result)
}

// ExportToExcel exports additional analyses to Excel format
func (h *AdditionalAnalysisHandler) ExportToExcel(c *fiber.Ctx) error {
	var req models.AdditionalAnalysisExportRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Set default format
	if req.Format == "" {
		req.Format = "excel"
	}

	// Export data
	data, err := h.analysisService.ExportToExcel(req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export additional analyses", err)
	}

	// Set appropriate headers for file download
	filename := "additional_analyses"
	if req.Format == "csv" {
		c.Set("Content-Type", "text/csv")
		filename += ".csv"
	} else {
		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		filename += ".xlsx"
	}

	c.Set("Content-Disposition", "attachment; filename="+filename)

	return c.Send(data)
}

// DownloadTemplate downloads Excel template for import
func (h *AdditionalAnalysisHandler) DownloadTemplate(c *fiber.Ctx) error {
	// Create template data with headers
	templateData := `Account Code,Analysis Type,Analysis Title,Status,Notes
12060700,revenue_recognition,Sample Analysis Title,active,Sample notes here
11010001,tax_treatment,Another Analysis,active,Another note`

	// Set headers for Excel file download
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=additional_analyses_template.csv")

	return c.SendString(templateData)
}