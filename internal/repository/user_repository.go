package repository

import (
	"context"
	"errors"
	"time"

	"github.com/lambare/go-mongo-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	col := db.Collection("users")

	// Unique index on email
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &UserRepository{col: col}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.col.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("email already registered")
	}
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &user, err
}
