package handler

import (
	"net/http"
	"time"

	"github.com/ronanlambare/backend-mongo-api/internal/model"
	"github.com/ronanlambare/backend-mongo-api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(repo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{repo: repo, jwtSecret: jwtSecret}
}

// Register godoc
// @Summary     Register a new user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body model.RegisterRequest true "Credentials"
// @Success     201  {object} model.LoginResponse
// @Failure     400  {object} model.ErrorResponse
// @Failure     409  {object} model.ErrorResponse "Username already exists"
// @Failure     500  {object} model.ErrorResponse
// @Router      /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.repo.Create(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, model.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	token, err := h.generateToken(user.ID.Hex(), user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to generate token"})
		return
	}
	c.JSON(http.StatusCreated, model.LoginResponse{Token: token})
}

// Login godoc
// @Summary     Login and get JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body model.LoginRequest true "Credentials"
// @Success     200  {object} model.LoginResponse
// @Failure     400  {object} model.ErrorResponse
// @Failure     401  {object} model.ErrorResponse
// @Failure     500  {object} model.ErrorResponse
// @Router      /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.repo.FindByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	if !h.repo.CheckPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid credentials"})
		return
	}

	token, err := h.generateToken(user.ID.Hex(), user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

func (h *AuthHandler) generateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
}
