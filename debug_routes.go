package main

import (
	"fmt"
	"accounting-web/internal/config"
	"accounting-web/internal/handler"
	"accounting-web/internal/middleware"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	app := fiber.New()

	// Mock database connection
	db, _ := sql.Open("mysql", "user:password@tcp(localhost:3306)/database")
	defer db.Close()

	// Mock Redis
	var redis *redis.Client

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	uploadRepo := repository.NewUploadRepository(db)
	rulesRepo := repository.NewRulesRepository(db)
	additionalAnalysisRepo := repository.NewAdditionalAnalysisRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, &config.Config{})
	excelService := service.NewExcelService()
	additionalAnalysisService := service.NewAdditionalAnalysisService(additionalAnalysisRepo, accountRepo, utils.GetLogger())

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountRepo)
	uploadHandler := handler.NewUploadHandler(uploadRepo, excelService, nil, &config.Config{})
	koreksiRuleHandler := handler.NewKoreksiRuleHandler(rulesRepo)
	obyekRuleHandler := handler.NewObyekRuleHandler(rulesRepo)
	ruleHandler := handler.NewGenericRuleHandler()
	additionalAnalysisHandler := handler.NewAdditionalAnalysisHandler(additionalAnalysisService)

	// Setup API routes
	protected := app.Group("", middleware.AuthMiddleware(&config.Config{}))

	// Upload routes
	uploads := protected.Group("/uploads")
	uploads.Get("/", uploadHandler.GetSessions)
	uploads.Get("/export", uploadHandler.ExportSessionsList)
	uploads.Get("/session/:session_code", uploadHandler.GetSessionDetailBySessionCode)
	uploads.Get("/session/:session_code/export", uploadHandler.ExportSessionByCode)

	// Print all routes
	fmt.Println("=== Registered Routes ===")
	app.Stack().ForEach(func(route fiber.Route) error {
		if len(route.Methods) > 0 && route.Path != "" {
			fmt.Printf("%-8s %s\n", route.Methods[0], route.Path)
		}
		return nil
	})

	fmt.Println("\n=== Upload Routes Specifically ===")
	fmt.Println("GET    /uploads/")
	fmt.Println("GET    /uploads/export")
	fmt.Println("GET    /uploads/session/:session_code")
	fmt.Println("GET    /uploads/session/:session_code/export")
}