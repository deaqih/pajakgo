package worker

import (
	"accounting-web/internal/config"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func RegisterHandlers(mux *asynq.ServeMux, db *sqlx.DB, redis *redis.Client, cfg *config.Config) {
	// Register processing task handler
	processor := NewProcessingHandler(db, redis, cfg)
	mux.HandleFunc("transaction:process", processor.HandleProcessing)
}
