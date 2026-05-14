package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ─── User ────────────────────────────────────────────────────────────────────

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id,omitempty"`
	Email     string             `bson:"email"           json:"email"`
	Password  string             `bson:"password"        json:"-"`
	Role      string             `bson:"role"            json:"role"`
	CreatedAt time.Time          `bson:"created_at"      json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"      json:"updated_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ─── Item ────────────────────────────────────────────────────────────────────
// Generic "item" collection. Adapt to your domain (products, articles, etc.)

type Item struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"  json:"id,omitempty"`
	Name        string                 `bson:"name"           json:"name"`
	Description string                 `bson:"description"    json:"description"`
	Tags        []string               `bson:"tags"           json:"tags"`
	Metadata    map[string]interface{} `bson:"metadata"       json:"metadata"`
	CreatedBy   primitive.ObjectID     `bson:"created_by"     json:"created_by"`
	CreatedAt   time.Time              `bson:"created_at"     json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at"     json:"updated_at"`
}

type CreateItemRequest struct {
	Name        string                 `json:"name"        binding:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type UpdateItemRequest struct {
	Name        *string                `json:"name"        binding:"omitempty,min=1,max=255"`
	Description *string                `json:"description"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int64       `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
