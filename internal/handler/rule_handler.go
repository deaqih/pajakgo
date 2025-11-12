package handler

import (
	"accounting-web/internal/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GenericRuleHandler provides generic pagination for all rule types
type GenericRuleHandler struct{}

func NewGenericRuleHandler() *GenericRuleHandler {
	return &GenericRuleHandler{}
}

// Mock data - In real application, these would come from repositories
var mockKoreksiRules = []fiber.Map{
	{"id": 1, "code": "K001", "description": "Koreksi Rule 1", "account_pattern": "1*", "tax_code": "VAT"},
	{"id": 2, "code": "K002", "description": "Koreksi Rule 2", "account_pattern": "2*", "tax_code": "WHT"},
	{"id": 3, "code": "K003", "description": "Koreksi Rule 3", "account_pattern": "3*", "tax_code": "PPH"},
	{"id": 4, "code": "K004", "description": "Koreksi Rule 4", "account_pattern": "4*", "tax_code": "VAT"},
	{"id": 5, "code": "K005", "description": "Koreksi Rule 5", "account_pattern": "5*", "tax_code": "WHT"},
}

var mockObyekRules = []fiber.Map{
	{"id": 1, "code": "O001", "description": "Obyek Rule 1", "account_pattern": "1*", "obyek_code": "100"},
	{"id": 2, "code": "O002", "description": "Obyek Rule 2", "account_pattern": "2*", "obyek_code": "200"},
	{"id": 3, "code": "O003", "description": "Obyek Rule 3", "account_pattern": "3*", "obyek_code": "300"},
	{"id": 4, "code": "O004", "description": "Obyek Rule 4", "account_pattern": "4*", "obyek_code": "400"},
	{"id": 5, "code": "O005", "description": "Obyek Rule 5", "account_pattern": "5*", "obyek_code": "500"},
}

var mockWithholdingTaxRules = []fiber.Map{
	{"id": 1, "code": "WHT001", "description": "WHT Rule 1", "rate": 2.5, "account_pattern": "1*"},
	{"id": 2, "code": "WHT002", "description": "WHT Rule 2", "rate": 10.0, "account_pattern": "2*"},
	{"id": 3, "code": "WHT003", "description": "WHT Rule 3", "rate": 15.0, "account_pattern": "3*"},
	{"id": 4, "code": "WHT004", "description": "WHT Rule 4", "rate": 20.0, "account_pattern": "4*"},
	{"id": 5, "code": "WHT005", "description": "WHT Rule 5", "rate": 25.0, "account_pattern": "5*"},
}

var mockTaxKeywords = []fiber.Map{
	{"id": 1, "keyword": "ppn", "description": "PPN Keyword", "category": "VAT"},
	{"id": 2, "keyword": "pph", "description": "PPH Keyword", "category": "WHT"},
	{"id": 3, "keyword": "pph21", "description": "PPH 21 Keyword", "category": "WHT"},
	{"id": 4, "keyword": "pph23", "description": "PPH 23 Keyword", "category": "WHT"},
	{"id": 5, "keyword": "pph26", "description": "PPH 26 Keyword", "category": "WHT"},
}

// Helper function to paginate mock data
func paginateMockData(data []fiber.Map, page, limit int) ([]fiber.Map, int64) {
	total := int64(len(data))
	if total == 0 {
		return data, total
	}

	start := (page - 1) * limit
	if start >= len(data) {
		return []fiber.Map{}, total
	}

	end := start + limit
	if end > len(data) {
		end = len(data)
	}

	return data[start:end], total
}

func (h *GenericRuleHandler) GetKoreksiRules(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	// Apply search filter if provided
	filteredData := mockKoreksiRules
	if params.Search != "" {
		filteredData = []fiber.Map{}
		for _, rule := range mockKoreksiRules {
			code := rule["code"].(string)
			desc := rule["description"].(string)
			if containsSearch(code, params.Search) || containsSearch(desc, params.Search) {
				filteredData = append(filteredData, rule)
			}
		}
	}

	rules, total := paginateMockData(filteredData, params.Page, params.Limit)
	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	responseData := fiber.Map{
		"rules": rules,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Koreksi rules retrieved successfully", responseData, pagination)
}

func (h *GenericRuleHandler) GetObyekRules(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	// Apply search filter if provided
	filteredData := mockObyekRules
	if params.Search != "" {
		filteredData = []fiber.Map{}
		for _, rule := range mockObyekRules {
			code := rule["code"].(string)
			desc := rule["description"].(string)
			if containsSearch(code, params.Search) || containsSearch(desc, params.Search) {
				filteredData = append(filteredData, rule)
			}
		}
	}

	rules, total := paginateMockData(filteredData, params.Page, params.Limit)
	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	responseData := fiber.Map{
		"rules": rules,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Obyek rules retrieved successfully", responseData, pagination)
}

func (h *GenericRuleHandler) GetWithholdingTaxRules(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	// Apply search filter if provided
	filteredData := mockWithholdingTaxRules
	if params.Search != "" {
		filteredData = []fiber.Map{}
		for _, rule := range mockWithholdingTaxRules {
			code := rule["code"].(string)
			desc := rule["description"].(string)
			if containsSearch(code, params.Search) || containsSearch(desc, params.Search) {
				filteredData = append(filteredData, rule)
			}
		}
	}

	rules, total := paginateMockData(filteredData, params.Page, params.Limit)
	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	responseData := fiber.Map{
		"rules": rules,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Withholding tax rules retrieved successfully", responseData, pagination)
}

func (h *GenericRuleHandler) GetTaxKeywords(c *fiber.Ctx) error {
	params := utils.GetPaginationParams(c)
	
	// Apply search filter if provided
	filteredData := mockTaxKeywords
	if params.Search != "" {
		filteredData = []fiber.Map{}
		for _, keyword := range mockTaxKeywords {
			word := keyword["keyword"].(string)
			desc := keyword["description"].(string)
			if containsSearch(word, params.Search) || containsSearch(desc, params.Search) {
				filteredData = append(filteredData, keyword)
			}
		}
	}

	keywords, total := paginateMockData(filteredData, params.Page, params.Limit)
	pagination := utils.CalculatePagination(params.Page, params.Limit, total)

	responseData := fiber.Map{
		"keywords": keywords,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Tax keywords retrieved successfully", responseData, pagination)
}

func (h *GenericRuleHandler) CreateKoreksiRule(c *fiber.Ctx) error {
	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// In real application, save to database
	return utils.SuccessResponse(c, "Koreksi rule created successfully", rule)
}

func (h *GenericRuleHandler) CreateObyekRule(c *fiber.Ctx) error {
	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// In real application, save to database
	return utils.SuccessResponse(c, "Obyek rule created successfully", rule)
}

func (h *GenericRuleHandler) CreateWithholdingTaxRule(c *fiber.Ctx) error {
	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// In real application, save to database
	return utils.SuccessResponse(c, "Withholding tax rule created successfully", rule)
}

func (h *GenericRuleHandler) CreateTaxKeyword(c *fiber.Ctx) error {
	var keyword map[string]interface{}
	if err := c.BodyParser(&keyword); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// In real application, save to database
	return utils.SuccessResponse(c, "Tax keyword created successfully", keyword)
}

func (h *GenericRuleHandler) UpdateKoreksiRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	rule["id"] = id
	return utils.SuccessResponse(c, "Koreksi rule updated successfully", rule)
}

func (h *GenericRuleHandler) UpdateObyekRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	rule["id"] = id
	return utils.SuccessResponse(c, "Obyek rule updated successfully", rule)
}

func (h *GenericRuleHandler) UpdateWithholdingTaxRule(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	var rule map[string]interface{}
	if err := c.BodyParser(&rule); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	rule["id"] = id
	return utils.SuccessResponse(c, "Withholding tax rule updated successfully", rule)
}

func (h *GenericRuleHandler) UpdateTaxKeyword(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid keyword ID", err)
	}

	var keyword map[string]interface{}
	if err := c.BodyParser(&keyword); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	keyword["id"] = id
	return utils.SuccessResponse(c, "Tax keyword updated successfully", keyword)
}

func (h *GenericRuleHandler) DeleteKoreksiRule(c *fiber.Ctx) error {
	_, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	return utils.SuccessResponse(c, "Koreksi rule deleted successfully", nil)
}

func (h *GenericRuleHandler) DeleteObyekRule(c *fiber.Ctx) error {
	_, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	return utils.SuccessResponse(c, "Obyek rule deleted successfully", nil)
}

func (h *GenericRuleHandler) DeleteWithholdingTaxRule(c *fiber.Ctx) error {
	_, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid rule ID", err)
	}

	return utils.SuccessResponse(c, "Withholding tax rule deleted successfully", nil)
}

func (h *GenericRuleHandler) DeleteTaxKeyword(c *fiber.Ctx) error {
	_, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid keyword ID", err)
	}

	return utils.SuccessResponse(c, "Tax keyword deleted successfully", nil)
}

// Helper function for case-insensitive search
func containsSearch(text, search string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(search))
}