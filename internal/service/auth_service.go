package service

import (
	"accounting-web/internal/config"
	"accounting-web/internal/models"
	"accounting-web/internal/repository"
	"accounting-web/internal/utils"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(*user, s.cfg.JWTSecret, s.cfg.JWTAccessExpire)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := utils.GenerateRefreshToken(*user, s.cfg.JWTSecret, s.cfg.JWTRefreshExpire)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(tokenString, s.cfg.JWTSecret)
}

func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	existingEmail, _ := s.userRepo.FindByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create new user
	user := &models.User{
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         "user", // Default role
		IsActive:     true,   // Auto-activate new users
	}

	// Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// WebLogin untuk autentikasi web menggunakan session
func (s *AuthService) WebLogin(req models.LoginRequest, c *fiber.Ctx, store *session.Store) (*models.User, error) {
	// Development mode: accept admin credentials
	if req.Username == "admin" && req.Password == "admin" {
		// Create mock user for development
		user := &models.User{
			ID:       1,
			Name:     "Development User",
			Username: "admin",
			Email:    "dev@example.com",
			Role:     "admin",
			IsActive: true,
		}

		// Create session
		sess, err := store.Get(c)
		if err != nil {
			return nil, errors.New("failed to create session")
		}

		// Set session data
		sess.Set("user_id", user.ID)
		sess.Set("username", user.Username)
		sess.Set("role", user.Role)
		sess.Set("expires_at", time.Now().Add(24*time.Hour).Unix())

		// Save session
		if err := sess.Save(); err != nil {
			return nil, errors.New("failed to save session")
		}

		return user, nil
	}

	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// Create session
	sess, err := store.Get(c)
	if err != nil {
		return nil, errors.New("failed to create session")
	}

	// Set session data
	sess.Set("user_id", user.ID)
	sess.Set("username", user.Username)
	sess.Set("role", user.Role)
	sess.Set("expires_at", time.Now().Add(24*time.Hour).Unix())

	// Save session
	if err := sess.Save(); err != nil {
		return nil, errors.New("failed to save session")
	}

	return user, nil
}

// WebLogout untuk menghapus session
func (s *AuthService) WebLogout(c *fiber.Ctx, store *session.Store) error {
	sess, err := store.Get(c)
	if err != nil {
		return err
	}

	// Destroy session
	return sess.Destroy()
}

// GetCurrentUser untuk mendapatkan user yang sedang login
func (s *AuthService) GetCurrentUser(c *fiber.Ctx, store *session.Store) (*models.User, error) {
	sess, err := store.Get(c)
	if err != nil {
		return nil, err
	}

	userID := sess.Get("user_id")
	if userID == nil {
		return nil, errors.New("user not logged in")
	}

	return s.userRepo.FindByID(userID.(int))
}
