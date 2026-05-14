package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type OIDCIdentity struct {
	Provider string `bson:"provider" json:"-"`
	Sub      string `bson:"sub"      json:"-"`
}

type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	Username   string             `bson:"username,omitempty"   json:"username,omitempty"`
	Password   string             `bson:"password,omitempty"   json:"-"`
	Email      string             `bson:"email,omitempty"      json:"email,omitempty"`
	Identities []OIDCIdentity     `bson:"identities,omitempty" json:"-"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type OIDCLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
