package repository

import (
	"context"
	"errors"

	"github.com/ronanlambare/backend-mongo-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	col := db.Collection("users")
	// Ensure unique index on username
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	})
	return &UserRepository{col: col}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.col.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := model.User{
		ID:       primitive.NewObjectID(),
		Username: username,
		Password: string(hash),
	}
	if _, err := r.col.InsertOne(ctx, user); err != nil {
		// Duplicate key → username already taken
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("username already exists")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
