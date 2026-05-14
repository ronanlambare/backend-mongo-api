package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ronanlambare/backend-mongo-api/internal/model"
	"github.com/ronanlambare/backend-mongo-api/internal/repository"
)

type AuthHandler struct {
	repo          *repository.UserRepository
	jwtSecret     string
	oidcVerifiers map[string]*gooidc.IDTokenVerifier // issuer → verifier
}

// NewAuthHandler initialises the handler and pre-fetches OIDC discovery documents
// for every configured provider (issuer → clientID).
// Pass an empty map when no OIDC providers are needed.
func NewAuthHandler(repo *repository.UserRepository, jwtSecret string, providers map[string]string) (*AuthHandler, error) {
	verifiers := make(map[string]*gooidc.IDTokenVerifier, len(providers))
	for issuer, clientID := range providers {
		p, err := gooidc.NewProvider(context.Background(), issuer)
		if err != nil {
			return nil, fmt.Errorf("OIDC provider %q: %w", issuer, err)
		}
		verifiers[issuer] = p.Verifier(&gooidc.Config{ClientID: clientID})
	}
	return &AuthHandler{
		repo:          repo,
		jwtSecret:     jwtSecret,
		oidcVerifiers: verifiers,
	}, nil
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

// OIDCLogin godoc
// @Summary     Sign in / sign up via any configured OIDC provider
// @Description Accepts an id_token issued by Google, Apple, or any other configured OIDC provider.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body model.OIDCLoginRequest true "id_token from the OIDC provider"
// @Success     200  {object} model.LoginResponse
// @Failure     400  {object} model.ErrorResponse
// @Failure     401  {object} model.ErrorResponse
// @Failure     500  {object} model.ErrorResponse
// @Router      /auth/oidc [post]
func (h *AuthHandler) OIDCLogin(c *gin.Context) {
	var req model.OIDCLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	// Read the issuer from the token WITHOUT verifying the signature yet.
	// The real verification happens through the provider's JWKS below.
	unverified, _, err := new(jwt.Parser).ParseUnverified(req.IDToken, jwt.MapClaims{})
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id_token format"})
		return
	}
	claims, ok := unverified.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id_token claims"})
		return
	}
	issuer, _ := claims["iss"].(string)
	if issuer == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "missing iss claim in id_token"})
		return
	}

	verifier, allowed := h.oidcVerifiers[issuer]
	if !allowed {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unknown identity provider"})
		return
	}

	ctx := c.Request.Context()
	idToken, err := verifier.Verify(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid or expired id_token"})
		return
	}

	var idClaims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&idClaims); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to read token claims"})
		return
	}

	user, err := h.repo.FindOrCreateByOIDC(ctx, issuer, idToken.Subject, idClaims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}

	token, err := h.generateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

func (h *AuthHandler) generateToken(userID, subject string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"usr": subject,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
}
