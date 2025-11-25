package middleware

import (
	"accounting-web/internal/config"
	"accounting-web/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header is required",
			})
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format",
			})
		}

		token := parts[1]

		// Development mode: accept dev tokens
		if strings.HasPrefix(token, "dev-token-") {
			c.Locals("user_id", 1)
			c.Locals("username", "admin")
			c.Locals("role", "admin")
			return c.Next()
		}

		// Validate token
		claims, err := utils.ValidateToken(token, cfg.JWTSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Admin access required",
			})
		}
		return c.Next()
	}
}

// WebAuthMiddleware untuk autentikasi halaman web menggunakan session
func WebAuthMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Development mode: Check for Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				// Development mode: accept dev tokens
				if strings.HasPrefix(token, "dev-token-") {
					c.Locals("user_id", 1)
					c.Locals("username", "admin")
					c.Locals("role", "admin")
					return c.Next()
				}
			}
		}

		// Check for auth cookie (set after successful login)
		authCookie := c.Cookies("auth_token")
		if authCookie != "" && strings.HasPrefix(authCookie, "dev-token-") {
			c.Locals("user_id", 1)
			c.Locals("username", "admin")
			c.Locals("role", "admin")
			return c.Next()
		}

		sess, err := store.Get(c)
		if err != nil {
			return c.Redirect("/login")
		}

		// Check if user is logged in
		userID := sess.Get("user_id")
		if userID == nil {
			return c.Redirect("/login")
		}

		// Check if session has expired (optional: implement session timeout)
		if sess.Get("expires_at") != nil {
			expiresAt := sess.Get("expires_at").(int64)
			if expiresAt < utils.GetCurrentTimestamp() {
				// Session expired, destroy and redirect to login
				sess.Destroy()
				return c.Redirect("/login")
			}
		}

		// Store user data in context
		c.Locals("user_id", userID)
		c.Locals("username", sess.Get("username"))
		c.Locals("role", sess.Get("role"))

		return c.Next()
	}
}

// GuestMiddleware untuk halaman yang hanya bisa diakses saat belum login
func GuestMiddleware(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Next()
		}

		// If user is already logged in, redirect to dashboard
		if sess.Get("user_id") != nil {
			return c.Redirect("/")
		}

		return c.Next()
	}
}