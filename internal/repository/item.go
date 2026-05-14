package repository

import (
	"context"
	"time"

	"github.com/ronanlambare/backend-mongo-api/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ItemRepository struct {
	col *mongo.Collection
}

func NewItemRepository(db *mongo.Database) *ItemRepository {
	return &ItemRepository{col: db.Collection("items")}
}

func (r *ItemRepository) FindAll(ctx context.Context) ([]model.Item, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []model.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ItemRepository) FindByID(ctx context.Context, id string) (*model.Item, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var item model.Item
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepository) Create(ctx context.Context, req model.CreateItemRequest) (*model.Item, error) {
	now := time.Now()
	item := model.Item{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.col.InsertOne(ctx, item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepository) Update(ctx context.Context, id string, req model.UpdateItemRequest) (*model.Item, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	update := bson.M{"$set": bson.M{
		"name":       req.Name,
		"updated_at": time.Now(),
	}}
	if _, err := r.col.UpdateOne(ctx, bson.M{"_id": oid}, update); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *ItemRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
