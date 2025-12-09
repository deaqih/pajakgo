package worker

import (
	"accounting-web/internal/config"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type ProcessingTaskHandler struct {
	db              *sqlx.DB
	redis           *redis.Client
	cfg             *config.Config
	processingEngine *service.ProcessingEngine
	uploadRepo      *repository.UploadRepository
}

func NewProcessingTaskHandler(db *sqlx.DB, redis *redis.Client, cfg *config.Config) *ProcessingTaskHandler {
	accountRepo := repository.NewAccountRepository(db)
	rulesRepo := repository.NewRulesRepository(db)
	uploadRepo := repository.NewUploadRepository(db)

	processingEngine := service.NewProcessingEngine(accountRepo, rulesRepo, uploadRepo)

	return &ProcessingTaskHandler{
		db:              db,
		redis:           redis,
		cfg:             cfg,
		processingEngine: processingEngine,
		uploadRepo:      uploadRepo,
	}
}

type ProcessingTaskPayload struct {
	SessionID   int    `json:"session_id"`
	SessionCode string `json:"session_code"`
}

func (h *ProcessingTaskHandler) Handle(ctx context.Context, task *asynq.Task) error {
	var payload ProcessingTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting processing for session %s (ID: %d)", payload.SessionCode, payload.SessionID)

	// Get session
	session, err := h.uploadRepo.GetSessionByID(payload.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has been canceled
	if session.Status == "canceled" {
		log.Printf("Session %s has been canceled, skipping processing", payload.SessionCode)
		return nil // Don't return error, just skip processing
	}

	// Check if session is already completed or failed
	if session.Status == "completed" || session.Status == "failed" {
		log.Printf("Session %s is already %s, skipping processing", payload.SessionCode, session.Status)
		return nil // Don't return error, just skip processing
	}

	// Load rules into processing engine
	if err := h.processingEngine.LoadRules(); err != nil {
		log.Printf("Failed to load rules: %v", err)
		h.uploadRepo.UpdateSessionStatus(payload.SessionID, "failed")
		return fmt.Errorf("failed to load rules: %w", err)
	}

	// Process in batches
	batchSize := h.cfg.BatchSize
	totalProcessed := 0
	totalFailed := 0

	for {
		// Get batch of unprocessed transactions
		transactions, err := h.uploadRepo.GetUnprocessedTransactions(payload.SessionID, batchSize)
		if err != nil {
			log.Printf("Failed to get unprocessed transactions: %v", err)
			break
		}

		if len(transactions) == 0 {
			break // All transactions processed
		}

		// Process batch
		if err := h.processingEngine.ProcessBatch(transactions); err != nil {
			log.Printf("Failed to process batch: %v", err)
			totalFailed += len(transactions)
		} else {
			totalProcessed += len(transactions)
		}

		// Update session progress
		session.ProcessedRows = totalProcessed
		session.FailedRows = totalFailed
		h.uploadRepo.UpdateSession(session)

		// Update progress in Redis
		progressKey := fmt.Sprintf("processing:progress:%d", payload.SessionID)
		progress := float64(totalProcessed) / float64(session.TotalRows) * 100
		h.redis.Set(ctx, progressKey, fmt.Sprintf("%.2f", progress), 0)

		log.Printf("Processed %d/%d transactions (%.2f%%)", totalProcessed, session.TotalRows, progress)
	}

	// Propagate field values to all rows with the same document_number
	log.Printf("Propagating field values for document_numbers in session %s", payload.SessionCode)
	if err := h.uploadRepo.PropagateDocumentNumberFields(payload.SessionCode); err != nil {
		log.Printf("Warning: Failed to propagate document number fields: %v", err)
		// Don't fail the entire processing, just log the warning
	} else {
		log.Printf("Successfully propagated field values for session %s", payload.SessionCode)
	}

	// Mark session as completed
	session.ProcessedRows = totalProcessed
	session.FailedRows = totalFailed
	session.Status = "completed"
	if err := h.uploadRepo.UpdateSession(session); err != nil {
		log.Printf("Failed to update session status: %v", err)
	}

	log.Printf("Processing completed for session %s. Processed: %d, Failed: %d",
		payload.SessionCode, totalProcessed, totalFailed)

	return nil
}
