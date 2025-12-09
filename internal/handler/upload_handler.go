package handler

import (
	"accounting-web/internal/config"
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"encoding/json"
	"fmt"
	"os"
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

	// No file count limit - unlimited files per batch

	// Validate total size limit (2GB for large uploads)
	const MAX_TOTAL_SIZE = 2 * 1024 * 1024 * 1024
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	if totalSize > MAX_TOTAL_SIZE {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Total size exceeds maximum limit of %s", formatFileSize(MAX_TOTAL_SIZE)), nil)
	}

	// Create upload session - one session for all files
	sessionCode := fmt.Sprintf("BATCH-%s", uuid.New().String()[:8])
	var uploadResults []map[string]interface{}
	var allTransactions []models.TransactionData
	var totalRows int

	// Check if this should be processed in background (large uploads)
	const BACKGROUND_THRESHOLD = 50000 // 50k rows trigger background processing

	// Create upload session first with estimated total
	session := &models.UploadSession{
		SessionCode: sessionCode,
		UserID:      userID,
		Filename:    "Processing...", // Will be updated later
		FilePath:    h.cfg.UploadPath,
		TotalRows:   0, // Will be updated after parsing
		Status:      "processing", // Set to processing immediately
	}

	// Create session record
	err = h.uploadRepo.CreateSession(session)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create upload session", err)
	}

	fmt.Printf("Created upload session: %s (ID: %d)\n", sessionCode, session.ID)

	// Process each file
	for i, file := range files {
		fmt.Printf("PROCESSING FILE %d: %s (session_code: %s)\n", i+1, file.Filename, sessionCode)

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
		fmt.Printf("Starting to parse file: %s (size: %d bytes, session_code: %s)\n", file.Filename, file.Size, sessionCode)
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
		fmt.Printf("Parsed %d rows in %v (session_code: %s, filename: %s)\n", len(transactions), parseTime, sessionCode, file.Filename)

		// Add to total rows for background processing decision
		totalRows += len(transactions)

		// Prepare transactions for saving with session_code only (session_id = 0)
		transactionsToSave := make([]models.TransactionData, len(transactions))
		for i, transaction := range transactions {
			transactionsToSave[i] = transaction
			transactionsToSave[i].SessionID = 0 // Explicitly set to 0
			transactionsToSave[i].SessionCode = sessionCode
			transactionsToSave[i].UserID = userID
			transactionsToSave[i].FilePath = filePath
			transactionsToSave[i].Filename = file.Filename

			// Debug first few transactions per file
			if i < 3 {
				fmt.Printf("DEBUG FILE %d: Transaction %d will use session_code='%s', filename='%s'\n",
					i+1, i+1, sessionCode, file.Filename)
			}
		}

		allTransactions = append(allTransactions, transactionsToSave...)

		uploadResults = append(uploadResults, map[string]interface{}{
			"filename":   file.Filename,
			"success":    true,
			"rows":       len(transactions),
			"size":       file.Size,
			"parse_time": parseTime.String(),
			"session_code": sessionCode, // Add session_code to response for debugging
		})
	}

	// Calculate statistics
	totalFiles := 0
	totalErrors := 0
	var firstFileName string
	var fileNames []string
	for _, result := range uploadResults {
		if result["success"].(bool) {
			totalFiles++
			fileNames = append(fileNames, result["filename"].(string))
			if firstFileName == "" {
				firstFileName = result["filename"].(string)
			}
		} else {
			totalErrors++
		}
	}

	if totalFiles == 0 {
		// Delete the session since no valid files
		h.uploadRepo.DeleteSession(session.ID)
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "No valid files were processed", nil)
	}

	// Update session with final details
	session.TotalRows = totalRows
	if totalFiles == 1 {
		session.Filename = firstFileName
	} else {
		session.Filename = fmt.Sprintf("Batch: %d files (%s)", totalFiles, firstFileName)
	}
	err = h.uploadRepo.UpdateSession(session)
	if err != nil {
		fmt.Printf("WARNING: Failed to update session: %v\n", err)
	}

	// Process immediately (faster - skip background processing for now)
	return h.processUploadOptimized(c, sessionCode, session.ID, userID, session.Filename, totalRows, allTransactions, uploadResults)
}

// processLargeUploadInBackground handles large uploads with background processing
func (h *UploadHandler) processLargeUploadInBackground(c *fiber.Ctx, sessionCode string, userID int, firstFileName string, totalRows int, transactions []models.TransactionData, uploadResults []map[string]interface{}) error {
	fmt.Printf("Processing large upload in background: %s (%d rows)\n", sessionCode, totalRows)

	// Create background job record
	backgroundJob, err := h.uploadRepo.CreateBackgroundJob(sessionCode, userID, firstFileName, totalRows)
	if err != nil {
		fmt.Printf("ERROR: Failed to create background job: %v\n", err)
		// Fall back to immediate processing
		return h.processUploadImmediately(c, sessionCode, userID, firstFileName, totalRows, transactions, uploadResults)
	}

	// Insert all transactions in chunks using the improved chunking function
	fmt.Printf("Inserting %d transactions to database...\n", len(transactions))
	insertStart := time.Now()

	err = h.uploadRepo.CreateMultipleTransactions(transactions)
	if err != nil {
		// Update background job with error
		errorMsg := fmt.Sprintf("Failed to insert transactions: %v", err)
		h.uploadRepo.UpdateBackgroundJobProgress(backgroundJob.ID, 0, "failed", &errorMsg)
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save transactions", err)
	}

	insertTime := time.Since(insertStart)
	fmt.Printf("Successfully inserted %d transactions in %v\n", len(transactions), insertTime)

	// Update background job with initial progress
	h.uploadRepo.UpdateBackgroundJobProgress(backgroundJob.ID, len(transactions), "uploaded", nil)

	// Queue background processing task
	if h.asynqClient != nil {
		payload, _ := json.Marshal(fiber.Map{
			"session_code": sessionCode,
			"background_job_id": backgroundJob.ID,
		})

		task := asynq.NewTask("upload:process_large", payload, asynq.MaxRetry(3))
		_, err = h.asynqClient.Enqueue(task)
		if err != nil {
			fmt.Printf("WARNING: Failed to queue background processing: %v\n", err)
			// Don't fail the upload, transactions are already saved
		}
	}

	// Return immediate response
	return utils.SuccessResponse(c, "Large upload queued for processing", fiber.Map{
		"session_code":   sessionCode,
		"total_files":   len(uploadResults),
		"total_errors":   0,
		"total_rows":    totalRows,
		"upload_results": uploadResults,
		"background_job": backgroundJob,
		"processing_mode": "background",
		"message":        "File berhasil diupload dan akan diproses di background. Halaman ini dapat direfresh untuk melihat progress.",
	})
}

// processUploadOptimized handles uploads with optimized processing using session_code only
func (h *UploadHandler) processUploadOptimized(c *fiber.Ctx, sessionCode string, sessionID int, userID int, filename string, totalRows int, transactions []models.TransactionData, uploadResults []map[string]interface{}) error {
	fmt.Printf("Processing upload optimized: %s (%d rows) using session_code only\n", sessionCode, totalRows)

	// Insert all transactions using the improved chunking function
	fmt.Printf("Inserting %d transactions to database with session_code: %s\n", len(transactions), sessionCode)
	insertStart := time.Now()

	err := h.uploadRepo.CreateMultipleTransactions(transactions)
	if err != nil {
		// Update session status to failed
		h.uploadRepo.UpdateSessionStatus(sessionID, "failed")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save transactions", err)
	}

	insertTime := time.Since(insertStart)
	fmt.Printf("Successfully inserted %d transactions in %v\n", len(transactions), insertTime)

	// Update session with final data: status, filename, and total_rows
	session := &models.UploadSession{
		ID:          sessionID,
		Filename:    filename,
		TotalRows:   totalRows,
		Status:      "uploaded",
	}

	err = h.uploadRepo.UpdateSession(session)
	if err != nil {
		fmt.Printf("WARNING: Failed to update session with final data: %v\n", err)
		// Fallback to status-only update
		err = h.uploadRepo.UpdateSessionStatus(sessionID, "uploaded")
		if err != nil {
			fmt.Printf("WARNING: Failed to update session status: %v\n", err)
		}
	} else {
		fmt.Printf("SUCCESS: Updated session with filename='%s', total_rows=%d, status='uploaded'\n", filename, totalRows)
	}

	return utils.SuccessResponse(c, "Files uploaded successfully", fiber.Map{
		"session_code":    sessionCode,
		"session_id":      sessionID,
		"total_files":    len(uploadResults),
		"total_errors":    0,
		"total_rows":     totalRows,
		"upload_results":  uploadResults,
		"processing_mode": "optimized",
		"message":        "Files successfully uploaded and ready for processing. Use session_code for all operations.",
	})
}

// processUploadImmediately handles smaller uploads with immediate processing
func (h *UploadHandler) processUploadImmediately(c *fiber.Ctx, sessionCode string, userID int, firstFileName string, totalRows int, transactions []models.TransactionData, uploadResults []map[string]interface{}) error {
	fmt.Printf("Processing upload immediately: %s (%d rows)\n", sessionCode, totalRows)

	// Insert all transactions using the improved chunking function
	fmt.Printf("Inserting %d transactions to database...\n", len(transactions))
	insertStart := time.Now()

	err := h.uploadRepo.CreateMultipleTransactions(transactions)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to save transactions", err)
	}

	insertTime := time.Since(insertStart)
	fmt.Printf("Successfully inserted %d transactions in %v\n", len(transactions), insertTime)

	// Create upload session record
	session := &models.UploadSession{
		SessionCode: sessionCode,
		UserID:      userID,
		Filename:    fmt.Sprintf("Batch: %d files (%s)", len(uploadResults), firstFileName),
		FilePath:    h.cfg.UploadPath,
		TotalRows:   totalRows,
		Status:      "uploaded",
	}

	fmt.Printf("Creating upload session for session_code: %s, user_id: %d\n", sessionCode, userID)
	err = h.uploadRepo.CreateSession(session)
	if err != nil {
		fmt.Printf("ERROR: Failed to create upload session record: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Created upload session with ID: %d\n", session.ID)

		// Update transactions with the actual session_id
		err = h.uploadRepo.UpdateTransactionsSessionID(sessionCode, session.ID)
		if err != nil {
			fmt.Printf("ERROR: Failed to update transactions with session_id: %v\n", err)
		} else {
			fmt.Printf("SUCCESS: Updated transactions with session_id: %d\n", session.ID)
		}
	}

	return utils.SuccessResponse(c, "Files uploaded successfully", fiber.Map{
		"session_code":    sessionCode,
		"total_files":    len(uploadResults),
		"total_errors":    0,
		"total_rows":     totalRows,
		"upload_results":  uploadResults,
		"processing_mode": "immediate",
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

	// Get pagination parameters with cursor support
	params := utils.GetPaginationParamsWithCursor(c)

	// Admin can see all sessions, user can only see their own
	filterUserID := 0
	if role != "admin" {
		filterUserID = userID
	}

	// Maximum records limit for cursor pagination
	maxRecords := 10000

	var sessions []models.UploadSession
	var pagination utils.PaginationMeta
	var err error

	// Use cursor pagination by default, fall back to offset if mode=offset
	if params.Mode == "cursor" {
		sessions, pagination, err = h.uploadRepo.GetSessionsWithCursor(params, filterUserID, maxRecords)
	} else {
		// Fallback to original offset-based pagination
		offset := utils.GetOffset(params.Page, params.Limit)
		var total int
		sessions, total, err = h.uploadRepo.GetSessionsOptimized(params.Limit, offset, filterUserID)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve sessions", err)
		}

		// Apply maximum limit
		if total > maxRecords {
			total = maxRecords
		}

		pagination = utils.CalculatePagination(params.Page, params.Limit, int64(total))
		pagination.Mode = "offset"
	}

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve sessions", err)
	}

	// Convert sessions to map format for response
	var allSessions []map[string]interface{}
	for _, session := range sessions {
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

	// Get session by ID first to get session_code
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Use session_code to get enhanced session detail with transaction statistics
	sessionDetail, err := h.uploadRepo.GetSessionDetailBySessionCode(session.SessionCode)
	if err != nil {
		// Fallback to basic session info if detailed query fails
		return utils.SuccessResponse(c, "Session retrieved successfully", session)
	}

	return utils.SuccessResponse(c, "Session detail retrieved successfully", sessionDetail)
}

// GetSessionDetailBySessionCode gets session details using session_code
func (h *UploadHandler) GetSessionDetailBySessionCode(c *fiber.Ctx) error {
	sessionCode := c.Params("session_code")
	if sessionCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session code is required", nil)
	}

	// Get session detail using session_code
	sessionDetail, err := h.uploadRepo.GetSessionDetailBySessionCode(sessionCode)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	return utils.SuccessResponse(c, "Session detail retrieved successfully", sessionDetail)
}

func (h *UploadHandler) GetTransactions(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid session ID", err)
	}

	// Get session first to get session_code
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Use session_code for transaction lookup - this is the optimized approach
	sessionCode := session.SessionCode

	// Get pagination parameters
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	transactions, total, err := h.uploadRepo.GetTransactionsBySessionCode(sessionCode, params.Limit, offset)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(total))

	responseData := fiber.Map{
		"session_id":    id,
		"session_code":  sessionCode,
		"transactions": transactions,
		"pagination":   pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Transactions retrieved successfully", responseData, pagination)
}

func (h *UploadHandler) GetTransactionsBySessionCode(c *fiber.Ctx) error {
	sessionCode := c.Params("session_code")
	if sessionCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session code is required", nil)
	}

	// Get pagination parameters with cursor support
	params := utils.GetPaginationParamsWithCursor(c)

	// Maximum records limit for cursor pagination - increased for session detail
	maxRecords := 1000000 // Allow up to 1 million records for session detail

	var transactions []models.TransactionData
	var pagination utils.PaginationMeta
	var err error

	// Use cursor pagination by default, fall back to offset if mode=offset
	if params.Mode == "cursor" {
		transactions, pagination, err = h.uploadRepo.GetTransactionsBySessionCodeWithCursor(sessionCode, params, maxRecords)
	} else {
		// Fallback to original offset-based pagination
		offset := utils.GetOffset(params.Page, params.Limit)
		var total int
		transactions, total, err = h.uploadRepo.GetTransactionsBySessionCode(sessionCode, params.Limit, offset)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
		}

		// Apply maximum limit
		if total > maxRecords {
			total = maxRecords
		}

		pagination = utils.CalculatePagination(params.Page, params.Limit, int64(total))
		pagination.Mode = "offset"
	}

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
	}

	responseData := fiber.Map{
		"session_code":  sessionCode,
		"transactions": transactions,
		"pagination":   pagination,
	}

	return utils.PaginatedResponseBuilder(c, "Transactions retrieved successfully", responseData, pagination)
}

func (h *UploadHandler) UpdateTransaction(c *fiber.Ctx) error {
	transactionID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid transaction ID", err)
	}

	var request struct {
		Koreksi *string `json:"koreksi"`
		Obyek   *string `json:"obyek"`
	}

	if err := c.BodyParser(&request); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Get user ID for authorization
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

	// Get role for authorization
	roleInterface := c.Locals("role")
	var role string
	if roleInterface != nil {
		role = fmt.Sprintf("%v", roleInterface)
	}

	// Update transaction
	err = h.uploadRepo.UpdateTransactionKoreksiObyek(transactionID, request.Koreksi, request.Obyek, userID, role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update transaction", err)
	}

	return utils.SuccessResponse(c, "Transaction updated successfully", nil)
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

// PropagateDocumentNumberFields triggers propagation of field values for a session
// This ensures all rows with the same document_number have the same values for WTH fields, PK CR, PM DB, etc.
func (h *UploadHandler) PropagateDocumentNumberFields(c *fiber.Ctx) error {
	sessionCode := c.Params("session_code")
	if sessionCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session code is required", nil)
	}

	// Verify session exists
	session, err := h.uploadRepo.GetSessionByCode(sessionCode)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Trigger propagation
	if err := h.uploadRepo.PropagateDocumentNumberFields(sessionCode); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to propagate document number fields", err)
	}

	return utils.SuccessResponse(c, "Document number fields propagated successfully", fiber.Map{
		"session_code": sessionCode,
		"session_id":   session.ID,
		"message":      "All rows with the same document_number now have propagated field values",
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

	// Use session_code to delete transactions - optimized approach
	sessionCode := session.SessionCode

	// Delete all transactions with this session_code
	if err := h.uploadRepo.DeleteTransactionsBySessionCode(sessionCode); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete transactions", err)
	}

	// Delete the session
	if err := h.uploadRepo.DeleteSession(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete session", err)
	}

	return utils.SuccessResponse(c, "Session deleted successfully", nil)
}

func (h *UploadHandler) GetUploadProgress(c *fiber.Ctx) error {
	sessionCode := c.Params("session_code")
	if sessionCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session code is required", nil)
	}

	// Try to get background job first (for large uploads)
	backgroundJob, err := h.uploadRepo.GetBackgroundJobBySessionCode(sessionCode)
	if err == nil && backgroundJob != nil {
		// Return background job progress
		progress := backgroundJob.GetProgressPercentage()
		return utils.SuccessResponse(c, "Progress retrieved successfully", fiber.Map{
			"session_code":    sessionCode,
			"total_rows":      backgroundJob.TotalRows,
			"processed_rows":  backgroundJob.ProcessedRows,
			"progress_percentage": progress,
			"status":          backgroundJob.Status,
			"error_message":   backgroundJob.ErrorMessage,
			"processing_mode": "background",
			"background_job":  backgroundJob,
			"created_at":      backgroundJob.CreatedAt,
			"updated_at":      backgroundJob.UpdatedAt,
		})
	}

	// If no background job, try to get regular session info
	// First try upload_sessions table
	session, err := h.uploadRepo.GetSessionByCode(sessionCode)
	if err == nil && session != nil {
		progress := float64(0)
		if session.TotalRows > 0 {
			progress = float64(session.ProcessedRows) / float64(session.TotalRows) * 100
		}

		return utils.SuccessResponse(c, "Progress retrieved successfully", fiber.Map{
			"session_code":    sessionCode,
			"total_rows":      session.TotalRows,
			"processed_rows":  session.ProcessedRows,
			"failed_rows":     session.FailedRows,
			"progress_percentage": progress,
			"status":          session.Status,
			"error_message":   session.ErrorMessage,
			"processing_mode": "immediate",
			"session":         session,
			"created_at":      session.CreatedAt,
			"updated_at":      session.UpdatedAt,
		})
	}

	// If no session found, check if it's a batch upload
	batchSessions, _, err := h.uploadRepo.GetBatchUploads(1, 0, 0) // Get all batches for admin
	if err == nil {
		for _, batch := range batchSessions {
			if batch.SessionCode == sessionCode {
				progress := float64(0)
				if batch.TotalRows > 0 {
					progress = float64(batch.ProcessedRows) / float64(batch.TotalRows) * 100
				}

				return utils.SuccessResponse(c, "Progress retrieved successfully", fiber.Map{
					"session_code":    sessionCode,
					"total_rows":      batch.TotalRows,
					"processed_rows":  batch.ProcessedRows,
					"failed_rows":     batch.FailedRows,
					"progress_percentage": progress,
					"status":          batch.Status,
					"processing_mode": "batch",
					"batch":           batch,
					"created_at":      batch.CreatedAt,
					"updated_at":      batch.UpdatedAt,
				})
			}
		}
	}

	return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", nil)
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

	// Get session to get session_code
	session, err := h.uploadRepo.GetSessionByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Session not found", err)
	}

	// Use session_code for transaction lookup - optimized approach
	sessionCode := session.SessionCode

	// Get all transactions using session_code (no pagination)
	transactions, _, err := h.uploadRepo.GetTransactionsBySessionCode(sessionCode, 1000000, 0)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
	}

	// Generate export file
	exportFileName := fmt.Sprintf("export_%s_%s.xlsx", sessionCode, time.Now().Format("20060102_150405"))
	exportPath := filepath.Join("./storage/exports", exportFileName)

	if err := h.excelService.ExportTransactions(transactions, exportPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export data", err)
	}

	// Send file
	return c.Download(exportPath, exportFileName)
}

// ExportSessionByCode exports session transactions using session_code (optimized method)
func (h *UploadHandler) ExportSessionByCode(c *fiber.Ctx) error {
	sessionCode := c.Params("session_code")
	if sessionCode == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Session code is required", nil)
	}

	// Log export request for debugging
	utils.GetLogger().Info("ExportSessionByCode called", map[string]interface{}{
		"session_code": sessionCode,
		"user_id":      c.Locals("user_id"),
		"role":         c.Locals("role"),
	})

	// Get all transactions using session_code - NO LIMIT for export
	// Use cursor pagination with very high limit to get ALL data
	maxRecords := 1000000 // Very high limit for export
	params := utils.PaginationParams{
		Mode:  "offset",
		Limit: maxRecords,
		Page:  1,
	}

	// Use the new cursor-based method to get all transactions
	transactions, _, err := h.uploadRepo.GetTransactionsBySessionCodeWithCursor(sessionCode, params, maxRecords)
	if err != nil {
		utils.GetLogger().Error("Failed to retrieve transactions with cursor method, trying fallback", map[string]interface{}{
			"session_code": sessionCode,
			"error":        err.Error(),
		})

		// Fallback to original method if cursor method fails
		transactions, _, err = h.uploadRepo.GetTransactionsBySessionCode(sessionCode, maxRecords, 0)
		if err != nil {
			utils.GetLogger().Error("Failed to retrieve transactions with fallback method", map[string]interface{}{
				"session_code": sessionCode,
				"error":        err.Error(),
			})
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve transactions", err)
		}
	}

	utils.GetLogger().Info("Retrieved transactions for export", map[string]interface{}{
		"session_code": sessionCode,
		"count":        len(transactions),
	})

	if len(transactions) == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "No transactions found for this session", nil)
	}

	// Generate export file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	exportFileName := fmt.Sprintf("transactions_%s_%s.xlsx", sessionCode, timestamp)
	exportPath := filepath.Join("./storage/exports", exportFileName)

	// Ensure the exports directory exists
	if err := os.MkdirAll("./storage/exports", 0755); err != nil {
		utils.GetLogger().Error("Failed to create exports directory", map[string]interface{}{
			"error": err.Error(),
		})
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create exports directory", err)
	}

	// Export transactions to Excel
	if err := h.excelService.ExportTransactions(transactions, exportPath); err != nil {
		utils.GetLogger().Error("Failed to export transactions to Excel", map[string]interface{}{
			"session_code": sessionCode,
			"error":        err.Error(),
		})
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export data", err)
	}

	utils.GetLogger().Info("Successfully exported transactions", map[string]interface{}{
		"session_code": sessionCode,
		"file_name":    exportFileName,
		"count":        len(transactions),
	})

	return c.Download(exportPath, exportFileName)
}

// ExportSessionsList exports filtered list of upload sessions
func (h *UploadHandler) ExportSessionsList(c *fiber.Ctx) error {
	// Get pagination and filter parameters
	params := utils.GetPaginationParamsWithCursor(c)

	// Get user ID and role
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

	// Admin can see all sessions, user can only see their own
	filterUserID := 0
	if role != "admin" {
		filterUserID = userID
	}

	// Maximum records for export (use maximum allowed)
	maxRecords := 10000
	params.Limit = maxRecords

	// Get sessions with filters
	sessions, _, err := h.uploadRepo.GetSessionsWithCursor(params, filterUserID, maxRecords)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve sessions", err)
	}

	// Convert sessions to export format
	exportData := make([]map[string]interface{}, len(sessions))
	for i, session := range sessions {
		exportData[i] = map[string]interface{}{
			"ID":           session.ID,
			"Session Code": session.SessionCode,
			"User ID":      session.UserID,
			"Filename":     session.Filename,
			"Total Rows":   session.TotalRows,
			"Processed":    session.ProcessedRows,
			"Failed":       session.FailedRows,
			"Status":       session.Status,
			"Error Message": session.ErrorMessage,
			"Created At":   session.CreatedAt.Format("2006-01-02 15:04:05"),
			"Updated At":   session.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Generate export file
	exportFileName := fmt.Sprintf("upload_sessions_export_%s.xlsx", time.Now().Format("20060102_150405"))
	exportPath := filepath.Join("./storage/exports", exportFileName)

	// Ensure the exports directory exists
	if err := os.MkdirAll("./storage/exports", 0755); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create exports directory", err)
	}

	// Export sessions to Excel using ExcelService
	if err := h.excelService.ExportSessionsList(exportData, exportPath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to export sessions", err)
	}

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
