package router

import (
	"accounting-web/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func Setup(app *fiber.App, db *sqlx.DB, redis *redis.Client, cfg *config.Config) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    cfg.AppName,
		})
	})

	// Web routes (HTML)
	web := app.Group("")
	setupWebRoutes(web, db, redis, cfg)

	// API routes (JSON)
	api := app.Group("/api/v1")
	SetupAPIRoutes(api, db, redis, cfg)
}

func setupWebRoutes(router fiber.Router, db *sqlx.DB, redis *redis.Client, cfg *config.Config) {
	// Authentication pages
	router.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("auth/login", fiber.Map{
			"Title": "Login",
		})
	})

	router.Get("/register", func(c *fiber.Ctx) error {
		return c.Render("auth/register", fiber.Map{
			"Title": "Register",
		})
	})

	// Dashboard (protected)
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Render("dashboard/index", fiber.Map{
			"Title": "Dashboard",
		})
	})

	// Master data pages
	router.Get("/accounts", func(c *fiber.Ctx) error {
		return c.Render("master/accounts", fiber.Map{
			"Title": "Accounts",
		})
	})

	router.Get("/koreksi-rules", func(c *fiber.Ctx) error {
		return c.Render("master/koreksi-rules", fiber.Map{
			"Title": "Koreksi Rules",
		})
	})

	router.Get("/obyek-rules", func(c *fiber.Ctx) error {
		return c.Render("master/obyek-rules", fiber.Map{
			"Title": "Obyek Rules",
		})
	})

	router.Get("/withholding-tax-rules", func(c *fiber.Ctx) error {
		return c.Render("master/withholding-tax-rules", fiber.Map{
			"Title": "Withholding Tax Rules",
		})
	})

	router.Get("/tax-keywords", func(c *fiber.Ctx) error {
		return c.Render("master/tax-keywords", fiber.Map{
			"Title": "Tax Keywords",
		})
	})

	// Upload pages
	router.Get("/uploads", func(c *fiber.Ctx) error {
		return c.Render("uploads/index", fiber.Map{
			"Title": "Upload Sessions",
		})
	})

	router.Get("/uploads/new", func(c *fiber.Ctx) error {
		return c.Render("uploads/new", fiber.Map{
			"Title": "New Upload",
		})
	})

	router.Get("/uploads/:id", func(c *fiber.Ctx) error {
		return c.Render("uploads/detail", fiber.Map{
			"Title": "Upload Detail",
		})
	})
}

