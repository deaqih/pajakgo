package worker

import (
	"accounting-web/internal/config"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func RegisterHandlers(mux *asynq.ServeMux, db *sqlx.DB, redis *redis.Client, cfg *config.Config) {
	// Create processing task handler
	processingHandler := NewProcessingTaskHandler(db, redis, cfg)

	// Register task handlers
	mux.HandleFunc("transaction:process", processingHandler.Handle)
}
