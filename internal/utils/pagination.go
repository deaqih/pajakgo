package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
	Search   string                 `json:"search"`
	OrderBy  string                 `json:"order_by"`
	OrderDir string                 `json:"order_dir"`
	Cursor   string                 `json:"cursor"`
	Mode     string                 `json:"mode"` // "offset" or "cursor"
	Filters  map[string]interface{} `json:"filters"` // Column-specific filters
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	CurrentPage int    `json:"current_page"`
	PerPage     int    `json:"per_page"`
	Total       int64  `json:"total"`
	LastPage    int    `json:"last_page"`
	From        int    `json:"from"`
	To          int    `json:"to"`
	HasMore     bool   `json:"has_more"`
	NextCursor  string `json:"next_cursor,omitempty"`
	PrevCursor  string `json:"prev_cursor,omitempty"`
	Mode        string `json:"mode"`
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

// GetEnhancedLimitOptions returns available limit options including show entries
func GetEnhancedLimitOptions() []int {
	return []int{10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}
}

// BuildColumnFilters builds SQL WHERE conditions for column filters
func BuildColumnFilters(filters map[string]interface{}, tableName string) string {
	if len(filters) == 0 {
		return ""
	}

	var conditions []string

	for column, value := range filters {
		if value == nil || value == "" {
			continue
		}

		// Sanitize column name to prevent SQL injection
		allowedColumns := map[string]bool{
			"session_code": true,
			"filename":     true,
			"status":       true,
			"user_id":      true,
			"total_rows":   true,
			"processed_rows": true,
			"failed_rows":  true,
		}

		if !allowedColumns[column] {
			continue // Skip invalid columns
		}

		// Build condition based on column type and value
		switch column {
		case "session_code", "filename":
			conditions = append(conditions, fmt.Sprintf("%s.%s LIKE '%%%s%%'", tableName, column, fmt.Sprintf("%v", value)))
		case "status":
			conditions = append(conditions, fmt.Sprintf("%s.%s = '%s'", tableName, column, fmt.Sprintf("%v", value)))
		case "user_id", "total_rows", "processed_rows", "failed_rows":
			// Numeric columns - support range and exact match
			valueStr := fmt.Sprintf("%v", value)
			if strings.Contains(valueStr, "-") {
				// Range filter (e.g., "100-500")
				parts := strings.Split(valueStr, "-")
				if len(parts) == 2 {
					conditions = append(conditions, fmt.Sprintf("%s.%s >= %s AND %s.%s <= %s", tableName, column, parts[0], tableName, column, parts[1]))
				}
			} else {
				// Exact match
				conditions = append(conditions, fmt.Sprintf("%s.%s = %s", tableName, column, valueStr))
			}
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return "(" + strings.Join(conditions, " AND ") + ")"
}

// GetAvailableFilters returns available column filters with their types
func GetAvailableFilters() map[string]string {
	return map[string]string{
		"session_code":   "text",
		"filename":       "text",
		"status":         "select",
		"user_id":        "number",
		"total_rows":     "number",
		"processed_rows": "number",
		"failed_rows":    "number",
	}
}

// GetStatusOptions returns available status filter options
func GetStatusOptions() []string {
	return []string{"uploaded", "processing", "completed", "failed"}
}

// Cursor represents a cursor for pagination
type Cursor struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int       `json:"user_id,omitempty"`
	SessionCode string  `json:"session_code,omitempty"`
}

// EncodeCursor encodes a cursor to base64 string
func EncodeCursor(cursor Cursor) (string, error) {
	jsonBytes, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// DecodeCursor decodes a base64 cursor string
func DecodeCursor(encodedCursor string) (*Cursor, error) {
	// Remove potential URL encoding
	decoded, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %v", err)
	}

	var cursor Cursor
	err = json.Unmarshal(decoded, &cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor data: %v", err)
	}

	return &cursor, nil
}

// GetPaginationParamsWithCursor extracts pagination parameters with cursor support
func GetPaginationParamsWithCursor(c *fiber.Ctx) PaginationParams {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "25"))
	search := c.Query("search", "")
	orderBy := c.Query("order_by", "created_at")
	orderDir := c.Query("order_dir", "desc")
	cursor := c.Query("cursor", "")
	mode := c.Query("mode", "cursor") // Default to cursor mode

	// Parse filters from query parameters
	filters := make(map[string]interface{})
	for key, values := range c.Queries() {
		if strings.HasPrefix(key, "filter_") {
			filterKey := strings.TrimPrefix(key, "filter_")
			if len(values) > 0 {
				filters[filterKey] = values[0]
			}
		}
	}

	// Validate and set defaults
	if page < 1 {
		page = 1
	}

	// Enhanced limit options with show entries support (max 10,000)
	validLimits := []int{10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}
	isValidLimit := false
	for _, validLimit := range validLimits {
		if limit == validLimit {
			isValidLimit = true
			break
		}
	}
	if !isValidLimit {
		limit = 25 // Default
	}

	if orderDir != "asc" && orderDir != "desc" {
		orderDir = "desc"
	}

	// Validate mode
	if mode != "offset" && mode != "cursor" {
		mode = "cursor"
	}

	// For cursor mode, ignore page parameter
	if mode == "cursor" {
		page = 1
	}

	return PaginationParams{
		Page:     page,
		Limit:    limit,
		Search:   search,
		OrderBy:  orderBy,
		OrderDir: orderDir,
		Cursor:   cursor,
		Mode:     mode,
		Filters:  filters,
	}
}

// CalculateCursorPagination calculates cursor-based pagination metadata
func CalculateCursorPagination(
	limit int,
	totalRecords int64,
	hasMore bool,
	nextCursor *Cursor,
	prevCursor *Cursor,
	mode string,
) PaginationMeta {
	// For cursor pagination, we don't calculate traditional page numbers
	// but we provide metadata for compatibility
	from := 1
	to := limit
	if totalRecords == 0 {
		from = 0
		to = 0
	} else if to > int(totalRecords) {
		to = int(totalRecords)
	}

	// Encode cursors
	var nextCursorStr, prevCursorStr string
	var err error

	if nextCursor != nil && hasMore {
		nextCursorStr, err = EncodeCursor(*nextCursor)
		if err != nil {
			nextCursorStr = ""
		}
	}

	if prevCursor != nil {
		prevCursorStr, err = EncodeCursor(*prevCursor)
		if err != nil {
			prevCursorStr = ""
		}
	}

	return PaginationMeta{
		CurrentPage: 1, // Cursor mode uses 1 for display
		PerPage:     limit,
		Total:       totalRecords,
		LastPage:    int(math.Ceil(float64(totalRecords) / float64(limit))),
		From:        from,
		To:          to,
		HasMore:     hasMore,
		NextCursor:  nextCursorStr,
		PrevCursor:  prevCursorStr,
		Mode:        mode,
	}
}

// ValidatePaginationLimit validates and adjusts limit based on maximum allowed
func ValidatePaginationLimit(limit int, maxAllowed int) int {
	if limit <= 0 {
		return 25 // Default
	}
	if limit > maxAllowed {
		return maxAllowed
	}
	return limit
}

// BuildCursorCondition builds SQL WHERE condition for cursor pagination
func BuildCursorCondition(params PaginationParams, cursor *Cursor, tableName string) string {
	if cursor == nil || params.Mode != "cursor" {
		return ""
	}

	var condition string
	var orderSymbol string

	if params.OrderDir == "desc" {
		orderSymbol = "<"
	} else {
		orderSymbol = ">"
	}

	// Build condition based on order by field
	switch strings.ToLower(params.OrderBy) {
	case "id":
		condition = fmt.Sprintf("%s.id %s %d", tableName, orderSymbol, cursor.ID)
	case "created_at":
		if params.OrderDir == "desc" {
			condition = fmt.Sprintf("%s.created_at < '%s' OR (%s.created_at = '%s' AND %s.id > %d)",
				tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"),
				tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"),
				tableName, cursor.ID)
		} else {
			condition = fmt.Sprintf("%s.created_at > '%s' OR (%s.created_at = '%s' AND %s.id > %d)",
				tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"),
				tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"),
				tableName, cursor.ID)
		}
	case "user_id":
		if params.OrderDir == "desc" {
			condition = fmt.Sprintf("%s.user_id < %d OR (%s.user_id = %d AND %s.id > %d)",
				tableName, cursor.UserID,
				tableName, cursor.UserID,
				tableName, cursor.ID)
		} else {
			condition = fmt.Sprintf("%s.user_id > %d OR (%s.user_id = %d AND %s.id > %d)",
				tableName, cursor.UserID,
				tableName, cursor.UserID,
				tableName, cursor.ID)
		}
	default:
		// Default to created_at
		if params.OrderDir == "desc" {
			condition = fmt.Sprintf("%s.created_at < '%s'", tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"))
		} else {
			condition = fmt.Sprintf("%s.created_at > '%s'", tableName, cursor.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	return condition
}

// BuildTransactionCursorCondition builds SQL WHERE condition for transaction cursor pagination
func BuildTransactionCursorCondition(cursor *Cursor, orderDir string, sessionCode string) string {
	if cursor == nil {
		return ""
	}

	var orderSymbol string
	if orderDir == "desc" {
		orderSymbol = "<"
	} else {
		orderSymbol = ">"
	}

	// For transactions, we always use ID for consistent ordering
	condition := fmt.Sprintf("td.id %s %d", orderSymbol, cursor.ID)

	// If we have a session_code, add it to the condition
	if sessionCode != "" {
		condition = fmt.Sprintf("td.session_code = '%s' AND %s", sessionCode, condition)
	}

	return condition
}