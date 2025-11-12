package utils

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Search   string `json:"search"`
	OrderBy  string `json:"order_by"`
	OrderDir string `json:"order_dir"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
	From        int   `json:"from"`
	To          int   `json:"to"`
	HasMore     bool  `json:"has_more"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Data       interface{}  `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "25"))
	search := c.Query("search", "")
	orderBy := c.Query("order_by", "")
	orderDir := c.Query("order_dir", "asc")

	// Validate and set defaults
	if page < 1 {
		page = 1
	}

	// Validate limit options
	validLimits := []int{10, 25, 50, 100}
	isValidLimit := false
	for _, validLimit := range validLimits {
		if limit == validLimit {
			isValidLimit = true
			break
		}
	}
	if !isValidLimit {
		limit = 25
	}

	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "asc"
	}

	return PaginationParams{
		Page:     page,
		Limit:    limit,
		Search:   search,
		OrderBy:  orderBy,
		OrderDir: orderDir,
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}

	lastPage := int(math.Ceil(float64(total) / float64(limit)))
	from := (page-1)*limit + 1
	to := page * limit

	if total == 0 {
		from = 0
		to = 0
	} else if to > int(total) {
		to = int(total)
	}

	return PaginationMeta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       total,
		LastPage:    lastPage,
		From:        from,
		To:          to,
		HasMore:     page < lastPage,
	}
}

// PaginatedResponseBuilder creates a paginated response
func PaginatedResponseBuilder(c *fiber.Ctx, message string, data interface{}, pagination PaginationMeta) error {
	response := PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}

	return c.JSON(response)
}

// GetOffset calculates offset for SQL queries
func GetOffset(page, limit int) int {
	return (page - 1) * limit
}

// GetLimitOptions returns available limit options
func GetLimitOptions() []int {
	return []int{10, 25, 50, 100}
}