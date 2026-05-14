package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name"          json:"name"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}

type CreateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

// ErrorResponse is the standard error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}
