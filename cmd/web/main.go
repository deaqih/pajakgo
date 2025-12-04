package main

import (
	"accounting-web/internal/config"
	"accounting-web/internal/database"
	"accounting-web/internal/router"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		log.Printf("Application will continue without database (read-only mode)")
		// Create a nil DB for development
		db = nil
	} else {
		defer db.Close()
	}

	// Initialize Redis (optional - for caching and background jobs)
	redisClient, err := database.NewRedis(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Printf("Application will continue without Redis (caching and background jobs disabled)")
	} else {
		defer redisClient.Close()
	}

	// Initialize template engine
	engine := html.New("./views", ".html")
	engine.Reload(cfg.AppEnv == "development")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		Views:        engine,
		BodyLimit:    cfg.UploadMaxSize,
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Static files
	app.Static("/static", "./public")
	app.Static("/shared", "./public/shared")

	// Setup routes
	router.Setup(app, db, redisClient, cfg)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nGracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	port := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Server starting on %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("Server exited")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Check if request expects JSON
	if c.Accepts("application/json") != "" {
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"message": message,
			"error":   err.Error(),
		})
	}

	// Return HTML error page
	return c.Status(code).Render("error", fiber.Map{
		"Code":    code,
		"Message": message,
	})
}
