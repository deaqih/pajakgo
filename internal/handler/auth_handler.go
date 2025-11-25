package handler

import (
	"accounting-web/internal/models"
	"accounting-web/internal/service"
	"accounting-web/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Username and password are required", nil)
	}

	var resp *models.LoginResponse
	var err error

	// Development mode: accept any credentials
	if req.Username == "admin" && req.Password == "admin" {
		// Create mock response
		resp = &models.LoginResponse{
			AccessToken:  "dev-token-" + req.Username,
			RefreshToken: "dev-refresh-token",
			User: models.User{
				ID:       1,
				Name:     "Development User",
				Username: req.Username,
				Email:    "dev@example.com",
				Role:     "admin",
				IsActive: true,
			},
		}
	} else {
		// Perform login
		resp, err = h.authService.Login(req)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error(), nil)
		}
	}

	// For development mode, create session by checking user-agent
	userAgent := c.Get("User-Agent")
	if userAgent != "" && (strings.Contains(userAgent, "Mozilla") || strings.Contains(userAgent, "Chrome") || strings.Contains(userAgent, "Safari")) {
		// This is a web browser request - create a mock session for development
		c.Locals("web_login", true)
	}

	return utils.SuccessResponse(c, "Login successful", resp)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// For JWT, logout is handled client-side by removing the token
	return utils.SuccessResponse(c, "Logout successful", nil)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err)
	}

	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Validate input
	if req.Name == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "All fields are required", nil)
	}

	// Validate password length
	if len(req.Password) < 6 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Password must be at least 6 characters", nil)
	}

	// Perform registration
	user, err := h.authService.Register(req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Registration successful", fiber.Map{
		"user": user,
	})
}
