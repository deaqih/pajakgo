package router

import (
	"accounting-web/internal/config"
	"accounting-web/internal/middleware"
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func Setup(app *fiber.App, db *sqlx.DB, redis *redis.Client, cfg *config.Config) {
	// Initialize session store
	store := session.New(session.Config{
		Expiration:   24 * time.Hour, // 24 hours
		KeyLookup:    "cookie:session_id",
		CookieDomain: "",
		CookiePath:   "/",
		CookieSecure: false, // Set to true in production with HTTPS
		CookieHTTPOnly: true,
		Storage: nil, // Use memory storage by default, can be replaced with Redis
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    cfg.AppName,
		})
	})

	// API routes (JSON) - Set up first to avoid conflicts
	api := app.Group("/api/v1")
	SetupAPIRoutes(api, db, redis, cfg)

	// Web routes (HTML) - Set up after API
	web := app.Group("")
	setupWebRoutes(web, db, redis, cfg, store)
}

func setupWebRoutes(router fiber.Router, db *sqlx.DB, redis *redis.Client, cfg *config.Config, store *session.Store) {
	// Initialize services
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg)

	// Authentication pages (guest only)
	router.Get("/login", middleware.GuestMiddleware(store), func(c *fiber.Ctx) error {
		return c.Render("auth/login", fiber.Map{
			"Title": "Login",
		})
	})

	router.Get("/register", middleware.GuestMiddleware(store), func(c *fiber.Ctx) error {
		return c.Render("auth/register", fiber.Map{
			"Title": "Register",
		})
	})

	// Authentication POST routes (guest only)
	router.Post("/login", middleware.GuestMiddleware(store), func(c *fiber.Ctx) error {
		var req models.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username and password are required", nil)
		}

		// Check if this is an AJAX request (from JavaScript)
		isAjax := c.Get("X-Requested-With") == "XMLHttpRequest" ||
				  c.Get("Content-Type") == "application/json" ||
				  c.Get("Accept") == "application/json"

		if isAjax {
			// API-style login - create session AND return JWT token
			user, err := authService.WebLogin(req, c, store)
			if err != nil {
				return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
			}

			// Generate JWT tokens for API calls
			accessToken := fmt.Sprintf("dev-token-%s", req.Username)

			// Return JWT tokens like API login
			return c.JSON(fiber.Map{
				"success": true,
				"message": "Login successful",
				"data": fiber.Map{
					"access_token":  accessToken,
					"refresh_token": "dev-refresh-token",
					"user": models.User{
						ID:       user.ID,
						Name:     user.Name,
						Username: user.Username,
						Email:    user.Email,
						Role:     user.Role,
						IsActive: true,
					},
				},
			})
		} else {
			// Web form login - create session and redirect
			_, err := authService.WebLogin(req, c, store)
			if err != nil {
				return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
			}

			// Redirect to dashboard on success
			return c.Redirect("/")
		}
	})

	router.Post("/logout", middleware.WebAuthMiddleware(store), func(c *fiber.Ctx) error {
		err := authService.WebLogout(c, store)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Logout failed", err)
		}

		// Redirect to login page
		return c.Redirect("/login")
	})

	// Protected routes (require login)
	protected := router.Group("", middleware.WebAuthMiddleware(store))

	// Dashboard
	protected.Get("/", func(c *fiber.Ctx) error {
		return c.Render("dashboard/index", fiber.Map{
			"Title": "Dashboard",
		})
	})

	// Master data pages
	protected.Get("/accounts", func(c *fiber.Ctx) error {
		return c.Render("master/accounts", fiber.Map{
			"Title": "Accounts",
		})
	})

	protected.Get("/koreksi-rules", func(c *fiber.Ctx) error {
		return c.Render("master/koreksi-rules", fiber.Map{
			"Title": "Koreksi Rules",
		})
	})

	protected.Get("/obyek-rules", func(c *fiber.Ctx) error {
		return c.Render("master/obyek-rules", fiber.Map{
			"Title": "Obyek Rules",
		})
	})

	protected.Get("/additional-analyses", func(c *fiber.Ctx) error {
		return c.Render("master/additional-analyses", fiber.Map{
			"Title": "Additional Analyses",
		})
	})

	protected.Get("/withholding-tax-rules", func(c *fiber.Ctx) error {
		return c.Render("master/withholding-tax-rules", fiber.Map{
			"Title": "Withholding Tax Rules",
		})
	})

	protected.Get("/tax-keywords", func(c *fiber.Ctx) error {
		return c.Render("master/tax-keywords", fiber.Map{
			"Title": "Tax Keywords",
		})
	})

	// Upload pages
	protected.Get("/uploads", func(c *fiber.Ctx) error {
		return c.Render("uploads/index", fiber.Map{
			"Title": "Upload Sessions",
		})
	})

	protected.Get("/uploads/new", func(c *fiber.Ctx) error {
		return c.Render("uploads/new", fiber.Map{
			"Title": "New Upload",
		})
	})

	protected.Get("/uploads/:id", func(c *fiber.Ctx) error {
		return c.Render("uploads/detail", fiber.Map{
			"Title": "Upload Detail",
		})
	})

	// New session code-based detail route for optimized performance
	protected.Get("/uploads/session/:session_code", func(c *fiber.Ctx) error {
		return c.Render("uploads/detail", fiber.Map{
			"Title": "Upload Detail",
		})
	})
}

