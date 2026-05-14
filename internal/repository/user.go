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
	// Unique index on username (sparse: OIDC users may not have one)
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	// Unique compound index on OIDC identities (provider + sub)
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "identities.provider", Value: 1},
			{Key: "identities.sub", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetSparse(true),
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

// FindOrCreateByOIDC finds an existing user by (provider, sub) or creates one.
// email may be empty (Apple does not return it on subsequent logins).
func (r *UserRepository) FindOrCreateByOIDC(ctx context.Context, provider, sub, email string) (*model.User, error) {
	filter := bson.M{
		"identities": bson.M{
			"$elemMatch": bson.M{"provider": provider, "sub": sub},
		},
	}

	var user model.User
	if err := r.col.FindOne(ctx, filter).Decode(&user); err == nil {
		return &user, nil
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	newUser := model.User{
		ID:    primitive.NewObjectID(),
		Email: email,
		Identities: []model.OIDCIdentity{
			{Provider: provider, Sub: sub},
		},
	}
	if _, err := r.col.InsertOne(ctx, newUser); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Race condition: user was created by a concurrent request
			if fetchErr := r.col.FindOne(ctx, filter).Decode(&user); fetchErr != nil {
				return nil, fetchErr
			}
			return &user, nil
		}
		return nil, err
	}
	return &newUser, nil
}
