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

func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

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
	transactions, err := h.excelService.ParseTransactionFile(filePath)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to parse Excel file", err)
	}

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
	})
}

func (h *UploadHandler) GetSessions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)

	// Get pagination parameters
	params := utils.GetPaginationParams(c)
	offset := utils.GetOffset(params.Page, params.Limit)

	// Admin can see all sessions, user can only see their own
	filterUserID := 0
	if role != "admin" {
		filterUserID = userID
	}

	sessions, total, err := h.uploadRepo.GetSessions(params.Limit, offset, filterUserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve sessions", err)
	}

	pagination := utils.CalculatePagination(params.Page, params.Limit, int64(total))

	responseData := fiber.Map{
		"sessions": sessions,
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
