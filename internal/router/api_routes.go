package router

import (
	"accounting-web/internal/config"
	"accounting-web/internal/handler"
	"accounting-web/internal/middleware"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func SetupAPIRoutes(
	router fiber.Router,
	db *sqlx.DB,
	redis *redis.Client,
	cfg *config.Config,
) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	uploadRepo := repository.NewUploadRepository(db)
	rulesRepo := repository.NewRulesRepository(db)
	additionalAnalysisRepo := repository.NewAdditionalAnalysisRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)
	excelService := service.NewExcelService()
	additionalAnalysisService := service.NewAdditionalAnalysisService(additionalAnalysisRepo, accountRepo, utils.GetLogger())

	// Initialize Asynq client (optional - only if Redis is available)
	var asynqClient *asynq.Client
	if redis != nil {
		asynqClient = asynq.NewClient(asynq.RedisClientOpt{
			Addr:     cfg.AsynqRedisAddr,
			Password: cfg.AsynqRedisPassword,
			DB:       cfg.AsynqRedisDB,
		})
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountRepo)
	uploadHandler := handler.NewUploadHandler(uploadRepo, excelService, asynqClient, cfg)
	koreksiRuleHandler := handler.NewKoreksiRuleHandler(rulesRepo)
	obyekRuleHandler := handler.NewObyekRuleHandler(rulesRepo)
	ruleHandler := handler.NewGenericRuleHandler()
	additionalAnalysisHandler := handler.NewAdditionalAnalysisHandler(additionalAnalysisService)

	// Public routes
	auth := router.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)
	auth.Post("/logout", authHandler.Logout)

	// Protected routes
	protected := router.Group("", middleware.AuthMiddleware(cfg))

	// Auth routes
	protected.Get("/auth/me", authHandler.Me)

	// Dashboard routes
	protected.Get("/dashboard/stats", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"total_uploads":     0,
				"completed_uploads": 0,
				"processing":        0,
				"failed":            0,
			},
		})
	})

	// Account routes
	accounts := protected.Group("/accounts")
	accounts.Get("/", accountHandler.GetAccounts)
	accounts.Get("/export", accountHandler.ExportAccounts)
	accounts.Get("/template", accountHandler.DownloadTemplate)
	accounts.Post("/import", accountHandler.ImportAccounts)
	accounts.Get("/error-report/:filename", accountHandler.DownloadErrorReport)
	accounts.Get("/:id", accountHandler.GetAccount)
	accounts.Post("/", accountHandler.CreateAccount)
	accounts.Put("/:id", accountHandler.UpdateAccount)
	accounts.Delete("/:id", accountHandler.DeleteAccount)

	// Additional Analysis routes
	additionalAnalysis := protected.Group("/additional-analyses")
	additionalAnalysis.Get("/", additionalAnalysisHandler.GetAll)
	additionalAnalysis.Get("/export", additionalAnalysisHandler.ExportToExcel)
	additionalAnalysis.Get("/template", additionalAnalysisHandler.DownloadTemplate)
	additionalAnalysis.Get("/types", additionalAnalysisHandler.GetAnalysisTypes)
	additionalAnalysis.Post("/import", additionalAnalysisHandler.ImportFromExcel)
	additionalAnalysis.Get("/:id", additionalAnalysisHandler.GetByID)
	additionalAnalysis.Post("/", additionalAnalysisHandler.Create)
	additionalAnalysis.Put("/:id", additionalAnalysisHandler.Update)
	additionalAnalysis.Delete("/:id", additionalAnalysisHandler.Delete)
	additionalAnalysis.Delete("/:id/hard", additionalAnalysisHandler.HardDelete)
	additionalAnalysis.Get("/account/:accountCode", additionalAnalysisHandler.GetByAccountCode)

	// Koreksi Rules routes
	koreksi := protected.Group("/koreksi-rules")
	koreksi.Get("/", koreksiRuleHandler.GetKoreksiRules)
	koreksi.Get("/export", koreksiRuleHandler.ExportKoreksiRules)
	koreksi.Get("/template", koreksiRuleHandler.DownloadTemplate)
	koreksi.Post("/import", koreksiRuleHandler.ImportKoreksiRules)
	koreksi.Get("/error-report/:filename", koreksiRuleHandler.DownloadErrorReport)
	koreksi.Get("/:id", koreksiRuleHandler.GetKoreksiRule)
	koreksi.Post("/", koreksiRuleHandler.CreateKoreksiRule)
	koreksi.Put("/:id", koreksiRuleHandler.UpdateKoreksiRule)
	koreksi.Delete("/:id", koreksiRuleHandler.DeleteKoreksiRule)

	// Obyek Rules routes
	obyek := protected.Group("/obyek-rules")
	obyek.Get("/", obyekRuleHandler.GetObyekRules)
	obyek.Get("/export", obyekRuleHandler.ExportObyekRules)
	obyek.Get("/template", obyekRuleHandler.DownloadTemplate)
	obyek.Post("/import", obyekRuleHandler.ImportObyekRules)
	obyek.Get("/error-report/:filename", obyekRuleHandler.DownloadErrorReport)
	obyek.Get("/:id", obyekRuleHandler.GetObyekRule)
	obyek.Post("/", obyekRuleHandler.CreateObyekRule)
	obyek.Put("/:id", obyekRuleHandler.UpdateObyekRule)
	obyek.Delete("/:id", obyekRuleHandler.DeleteObyekRule)

	// Withholding Tax Rules routes
	wht := protected.Group("/withholding-tax-rules")
	wht.Get("/", ruleHandler.GetWithholdingTaxRules)
	wht.Post("/", ruleHandler.CreateWithholdingTaxRule)
	wht.Put("/:id", ruleHandler.UpdateWithholdingTaxRule)
	wht.Delete("/:id", ruleHandler.DeleteWithholdingTaxRule)

	// Tax Keywords routes
	taxKeywords := protected.Group("/tax-keywords")
	taxKeywords.Get("/", ruleHandler.GetTaxKeywords)
	taxKeywords.Post("/", ruleHandler.CreateTaxKeyword)
	taxKeywords.Put("/:id", ruleHandler.UpdateTaxKeyword)
	taxKeywords.Delete("/:id", ruleHandler.DeleteTaxKeyword)

	// Upload routes
	uploads := protected.Group("/uploads")
	uploads.Post("/", uploadHandler.UploadFile)
	uploads.Post("/multiple", uploadHandler.UploadMultipleFiles)
	uploads.Get("/", uploadHandler.GetSessions)
	uploads.Get("/template", uploadHandler.DownloadTemplate)
	uploads.Get("/:id", uploadHandler.GetSessionDetail)
	uploads.Get("/session/:session_code", uploadHandler.GetSessionDetailBySessionCode) // New session code-based detail
	uploads.Get("/session/:session_code/transactions", uploadHandler.GetTransactionsBySessionCode) // New optimized route - MOVED UP
	uploads.Get("/:id/transactions", uploadHandler.GetTransactions)
	uploads.Post("/:id/process", uploadHandler.ProcessSession)
	uploads.Post("/:id/cancel", uploadHandler.CancelSession)
	uploads.Get("/:id/export", uploadHandler.ExportSession)
	uploads.Get("/session/:session_code/export", uploadHandler.ExportSessionByCode)
	uploads.Delete("/:id", uploadHandler.DeleteSession)
	uploads.Get("/progress/:session_code", uploadHandler.GetUploadProgress)

	// Transaction routes
	protected.Put("/transactions/:id", uploadHandler.UpdateTransaction)

	// Job progress routes
	jobs := protected.Group("/jobs")
	jobs.Get("/:job_id/progress", func(c *fiber.Ctx) error {
		jobID := c.Params("job_id")
		// TODO: Implement actual progress tracking
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"job_id":   jobID,
				"progress": 50.0,
				"status":   "processing",
			},
		})
	})
}
