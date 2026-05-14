package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/lambare/go-mongo-api/internal/auth"
	"github.com/lambare/go-mongo-api/internal/config"
	"github.com/lambare/go-mongo-api/internal/models"
	"github.com/lambare/go-mongo-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users *repository.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

// Register godoc
// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hash),
		Role:     "user",
	}

	if err := h.users.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	tokens, err := jwtpkg.GenerateTokenPair(
		user.ID.Hex(), user.Email, user.Role,
		h.cfg.JWTSecret, h.cfg.JWTExpiry, h.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.cfg.JWTExpiry.Seconds()),
		User:         user,
	})
}

// Login godoc
// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Constant-time response to avoid timing attacks
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokens, err := jwtpkg.GenerateTokenPair(
		user.ID.Hex(), user.Email, user.Role,
		h.cfg.JWTSecret, h.cfg.JWTExpiry, h.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(h.cfg.JWTExpiry.Seconds()),
		User:         user,
	})
}

// Refresh godoc
// POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := jwtpkg.ValidateClaims(req.RefreshToken, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	if claims.Type != jwtpkg.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	tokens, err := jwtpkg.GenerateTokenPair(
		claims.UserID, claims.Email, claims.Role,
		h.cfg.JWTSecret, h.cfg.JWTExpiry, h.cfg.JWTRefreshExpiry,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.cfg.JWTExpiry.Seconds()),
	})
}
