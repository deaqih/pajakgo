package handler

import (
	"accounting-web/internal/config"
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type UploadHandler struct {
	uploadRepo   *repository.UploadRepository
	excelService *service.ExcelService
	asynqClient  *asynq.Client
	cfg          *config.Config
}

func NewUploadHandler(
	uploadRepo *repository.UploadRepository,
	excelService *service.ExcelService,
	asynqClient *asynq.Client,
	cfg *config.Config,
) *UploadHandler {
	return &UploadHandler{
		uploadRepo:   uploadRepo,
		excelService: excelService,
		asynqClient:  asynqClient,
		cfg:          cfg,
	}
}

func (h *UploadHandler) UploadMultipleFiles(c *fiber.Ctx) error {
	// Get user ID with type assertion safety
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case string:
		// Try to parse string to int
		if id, err := strconv.Atoi(v); err == nil {
			userID = id
		} else {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID format", nil)
		}
	default:
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID type", nil)
	}

	// Get multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to parse multipart form", err)
	}

	// Get files from form
	files := form.File["files"]
	if len(files) == 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "No files selected", nil)
	}

	// Validate file count limit (20 files)
	const MAX_FILES = 20
	if len(files) > MAX_FILES {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Maximum %d files allowed per upload", MAX_FILES), nil)
	}

	// Validate total size limit (500MB)
	const MAX_TOTAL_SIZE = 500 * 1024 * 1024
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	if totalSize > MAX_TOTAL_SIZE {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Total size exceeds maximum limit of %s", formatFileSize(MAX_TOTAL_SIZE)), nil)
	}

	// Create upload session
	sessionCode := fmt.Sprintf("BATCH-%s", uuid.New().String()[:8])
	var uploadResults []map[string]interface{}

	// Process each file
	for i, file := range files {
		// Validate file type
		ext := filepath.Ext(file.Filename)
		if ext != ".xlsx" && ext != ".xls" {
			uploadResults = append(uploadResults, map[string]interface{}{
				"filename": file.Filename,
				"success":  false,
				"error":    "Only Excel files (.xlsx, .xls) are allowed",
			})
			continue
		}

		// Validate individual file size
		if file.Size > int64(h.cfg.UploadMaxSize) {
			uploadResults = append(uploadResults, map[string]interface{}{
				"filename": file.Filename,
				"success":  false,
				"error":    "File size exceeds maximum limit",
			})
			continue
		}

		// Save file with unique name to avoid conflicts
		filePath := filepath.Join(h.cfg.UploadPath, fmt.Sprintf("%s_%d%s", sessionCode, i+1, ext))
		if err := c.SaveFile(file, filePath); err != nil {
			uploadResults = append(uploadResults, map[string]interface{}{
				"filename": file.Filename,
				"success":  false,
				"error":    "Failed to save file",
			})
			continue
		}

		// Parse Excel file
		fmt.Printf("Starting to parse file: %s (size: %d bytes)\n", file.Filename, file.Size)
		startTime := time.Now()

		transactions, err := h.excelService.ParseTransactionFile(filePath)
		if err != nil {
			uploadResults = append(uploadResults, map[string]interface{}{
				"filename": file.Filename,
				"success":  false,
				"error":    fmt.Sprintf("Failed to parse Excel file: %v", err),
			})
			continue
		}

		parseTime := time.Since(startTime)
		fmt.Printf("Parsed %d rows in %v\n", len(transactions), parseTime)

		// Save to database
		for _, transaction := range transactions {
			transaction.SessionCode = sessionCode
			transaction.UserID = userID
			transaction.FilePath = filePath
			transaction.Filename = file.Filename
		}

		err = h.uploadRepo.CreateMultipleTransactions(transactions)
		if err != nil {
			uploadResults = append(uploadResults, map[string]interface{}{
				"filename": file.Filename,
				"success": false,
				"error":    fmt.Sprintf("Failed to save transactions: %v", err),
			})
			continue
		}

		uploadResults = append(uploadResults, map[string]interface{}{
			"filename": file.Filename,
			"success":  true,
			"rows":     len(transactions),
			"size":     file.Size,
			"parse_time": parseTime.String(),
		})
	}

	// Calculate statistics
	totalRows := 0
	totalFiles := 0
	totalErrors := 0
	var firstFileName string
	for _, result := range uploadResults {
		if result["success"].(bool) {
			totalRows += result["rows"].(int)
			totalFiles++
			if firstFileName == "" {
				firstFileName = result["filename"].(string)
			}
		} else {
			totalErrors++
		}
	}

	// Create upload session record if at least one file was processed successfully
	if totalFiles > 0 {
		status := "uploaded"
		if totalErrors > 0 {
			status = "uploaded" // Some files failed but batch has some success
		}

		session := &models.UploadSession{
			SessionCode: sessionCode,
			UserID:      userID,
			Filename:    fmt.Sprintf("Batch: %d files (%s)", totalFiles, firstFileName),
			FilePath:    h.cfg.UploadPath,
			TotalRows:   totalRows,
			Status:      status,
		}

		err = h.uploadRepo.CreateSession(session)
		if err != nil {
			// Log error but don't fail the upload - transactions are already saved
			fmt.Printf("Warning: Failed to create upload session record: %v\n", err)
		}

		// Now update transactions with the actual session_id
		err = h.uploadRepo.UpdateTransactionsSessionID(sessionCode, session.ID)
		if err != nil {
			// Log error but don't fail the upload
			fmt.Printf("Warning: Failed to update transactions with session_id: %v\n", err)
		}
	}

	return utils.SuccessResponse(c, "Files uploaded successfully", fiber.Map{
		"session_code":   sessionCode,
		"total_files":   totalFiles,
		"total_errors":   totalErrors,
		"total_rows":    totalRows,
		"upload_results": uploadResults,
	})
}

func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	// Get user ID with type assertion safety
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case string:
		// Try to parse string to int
		if id, err := strconv.Atoi(v); err == nil {
			userID = id
		} else {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID format", nil)
		}
	default:
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID type", nil)
	}

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

	// Validate file size
	if file.Size > int64(h.cfg.UploadMaxSize) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File size exceeds maximum limit", nil)
	}

	// Generate session code
	sessionCode := fmt.Sprintf("UPLOAD-%s", uuid.New().String()[:8])

	// Save file
	filePath := filepath.Join(h.cfg.UploadPath, fmt.Sprintf("%s%s", sessionCode, ext))
	if err := c.SaveFile(file, filePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save file", err)
	}

	// Parse Excel file
	fmt.Printf("Starting to parse file: %s (size: %d bytes)\n", file.Filename, file.Size)
	startTime := time.Now()

	transactions, err := h.excelService.ParseTransactionFile(filePath)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to parse Excel file", err)
	}

	parseTime := time.Since(startTime)
	fmt.Printf("Parsed %d rows in %v\n", len(transactions), parseTime)

	// Create upload session
	session := &models.UploadSession{
		SessionCode: sessionCode,
		UserID:      userID,
		Filename:    file.Filename,
		FilePath:    filePath,
		TotalRows:   len(transactions),
		Status:      "uploaded",
	}

	if err := h.uploadRepo.CreateSession(session); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create upload session", err)
	}

	// Insert transactions in batches
	batchSize := h.cfg.BatchSize
	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		for j := range batch {
			batch[j].SessionID = session.ID
		}

		if err := h.uploadRepo.BulkInsertTransactions(batch); err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to insert transactions", err)
		}
	}

	return utils.SuccessResponse(c, "File uploaded successfully", fiber.Map{
		"session":     session,
		"total_rows":  len(transactions),
		"preview":     getPreview(transactions, 10),
		"file_size":   file.Size,
		"processing_time": "completed",
	})
}

func (h *UploadHandler) GetSessions(c *fiber.Ctx) error {
	// Get user ID and role with type assertion safety
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			userID = id
		} else {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID format", nil)
		}
	default:
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID type", nil)
	}

	roleInterface := c.Locals("role")
	var role string
	if roleInterface != nil {
		role = fmt.Sprintf("%v", roleInterface)
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	// Admin can see all sessions, user can only see their own
	filterUserID := 0
	if role != "admin" {
		filterUserID = userID
	}

	// Get regular upload sessions
	var regularSessions []models.UploadSession
	var batchSessions []models.BatchUploadSession
	var totalRegular, totalBatch int
	var err error

	regularSessions, totalRegular, err = h.uploadRepo.GetSessions(params.Limit, offset, filterUserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve regular sessions", err)
	}

	// Get batch upload sessions (from transaction_data with session_id = 0)
	batchSessions, totalBatch, err = h.uploadRepo.GetBatchUploads(params.Limit, offset, filterUserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve batch sessions", err)
	}

	// Combine sessions for response
	var allSessions []interface{}
	for _, session := range regularSessions {
		sessionMap := map[string]interface{}{
			"id":             session.ID,
			"session_code":   session.SessionCode,
			"user_id":        session.UserID,
			"filename":       session.Filename,
			"file_path":      session.FilePath,
			"total_rows":     session.TotalRows,
			"processed_rows": session.ProcessedRows,
			"failed_rows":    session.FailedRows,
			"status":         session.Status,
			"error_message":  session.ErrorMessage,
			"created_at":     session.CreatedAt,
			"updated_at":     session.UpdatedAt,
			"is_batch":       false,
		}
		allSessions = append(allSessions, sessionMap)
	}

	for _, batch := range batchSessions {
		allSessions = append(allSessions, batch.ToUploadSession())
	}

	// Combine totals and calculate pagination
	totalCombined := totalRegular + totalBatch
	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(totalCombined))

	responseData := fiber.Map{
		"sessions": allSessions,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Sessions retrieved successfully", responseData, pagination)
}

func (h *UploadHandler) GetSessionDetail(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	return utils.SuccessResponse(c, "Session retrieved successfully", session)
}

func (h *UploadHandler) GetTransactions(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	transactions, total, err := h.uploadRepo.GetTransactionsBySession(id, params.Limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(total))

	responseData := fiber.Map{
		"transactions": transactions,
		"pagination": pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Transactions retrieved successfully", responseData, pagination)
}

func (h *UploadHandler) ProcessSession(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get session
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Check if already processing or completed
	if session.Status == "processing" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session is already being processed", nil)
	}
	if session.Status == "completed" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session is already completed", nil)
	}

	// Update status to processing
	if err := h.uploadRepo.UpdateSessionStatus(id, "processing"); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update session status", err)
	}

	// Create processing task
	if h.asynqClient == nil {
		return utils.ErrorResponse(c, fiber.StatusServiceUnavailable, "Background job processing is not available (Redis not connected)", nil)
	}

	payload, _ := json.Marshal(fiber.Map{
		"session_id":   session.ID,
		"session_code": session.SessionCode,
	})

	task := asynq.NewTask("transaction:process", payload)
	info, err := h.asynqClient.Enqueue(task)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to queue processing task", err)
	}

	return utils.SuccessResponse(c, "Processing started", fiber.Map{
		"job_id":  info.ID,
		"session": session,
	})
}

func (h *UploadHandler) CancelSession(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get session
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Check if session can be canceled
	if session.Status != "processing" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Only processing sessions can be canceled", nil)
	}

	// Update status to canceled
	if err := h.uploadRepo.UpdateSessionStatus(id, "canceled"); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update session status", err)
	}

	// Cancel active jobs in Asynq
	if h.asynqClient != nil {
		// Create inspector to check and cancel jobs
		inspector := asynq.NewInspector(asynq.RedisClientOpt{
			Addr:     h.cfg.AsynqRedisAddr,
			Password: h.cfg.AsynqRedisPassword,
			DB:       h.cfg.AsynqRedisDB,
		})

		// Get active queues - simplified approach
		// For now, we'll just mark the session as canceled
		// The worker will check the session status before processing
		_ = inspector
	}

	return utils.SuccessResponse(c, "Processing canceled successfully", fiber.Map{
		"session": session,
	})
}

func (h *UploadHandler) DeleteSession(c *fiber.Ctx) error {
	// Get user ID and role with type assertion safety
	userIDInterface := c.Locals("user_id")
	if userIDInterface == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			userID = id
		} else {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID format", nil)
		}
	default:
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID type", nil)
	}

	roleInterface := c.Locals("role")
	var role string
	if roleInterface != nil {
		role = fmt.Sprintf("%v", roleInterface)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get session to verify ownership or admin rights
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Check if user can delete this session
	if role != "admin" && session.UserID != userID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "You can only delete your own sessions", nil)
	}

	// Delete transactions first (due to foreign key constraint)
	if err := h.uploadRepo.DeleteTransactionsBySession(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete transactions", err)
	}

	// Delete the session
	if err := h.uploadRepo.DeleteSession(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete session", err)
	}

	return utils.SuccessResponse(c, "Session deleted successfully", nil)
}

func (h *UploadHandler) DownloadTemplate(c *fiber.Ctx) error {
	// Generate template filename
	templateFileName := "transaction_upload_template.xlsx"
	templatePath := filepath.Join("./storage/exports", templateFileName)

	// Generate transaction template
	if err := h.excelService.GenerateTransactionTemplate(templatePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate template", err)
	}

	// Send file
	return c.Download(templatePath, templateFileName)
}

func (h *UploadHandler) ExportSession(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get session
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Get all transactions (no pagination)
	transactions, _, err := h.uploadRepo.GetTransactionsBySession(id, 1000000, 0)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
	}

	// Generate export file
	exportFileName := fmt.Sprintf("export_%s_%s.xlsx", session.SessionCode, time.Now().Format("20060102_150405"))
	exportPath := filepath.Join("./storage/exports", exportFileName)

	if err := h.excelService.ExportTransactions(transactions, exportPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export data", err)
	}

	// Send file
	return c.Download(exportPath, exportFileName)
}

func getPreview(transactions []models.TransactionData, limit int) []models.TransactionData {
	if len(transactions) > limit {
		return transactions[:limit]
	}
	return transactions
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
