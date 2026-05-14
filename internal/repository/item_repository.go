package repository

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/lambare/go-mongo-api/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ItemRepository struct {
	col *mongo.Collection
}

func NewItemRepository(db *mongo.Database) *ItemRepository {
	col := db.Collection("items")

	// Useful indexes
	col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: "text"}, {Key: "description", Value: "text"}}},
		{Keys: bson.D{{Key: "created_by", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})

	return &ItemRepository{col: col}
}

// List returns a paginated list of items, optionally filtered by tags or text search.
func (r *ItemRepository) List(ctx context.Context, page, pageSize int, search string, tags []string) ([]*models.Item, int64, error) {
	filter := bson.M{}

	if search != "" {
		filter["$text"] = bson.M{"$search": search}
	}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$all": tags}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	opts := options.Find().
		SetSkip(skip).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var items []*models.Item
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetByID returns a single item by its ObjectID.
func (r *ItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Item, error) {
	var item models.Item
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &item, err
}

// Create inserts a new item.
func (r *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	item.ID = primitive.NewObjectID()
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	if item.Tags == nil {
		item.Tags = []string{}
	}

	_, err := r.col.InsertOne(ctx, item)
	return err
}

// Update applies partial changes to an existing item.
func (r *ItemRepository) Update(ctx context.Context, id primitive.ObjectID, req *models.UpdateItemRequest) (*models.Item, error) {
	update := bson.M{"updated_at": time.Now()}

	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Description != nil {
		update["description"] = *req.Description
	}
	if req.Tags != nil {
		update["tags"] = req.Tags
	}
	if req.Metadata != nil {
		update["metadata"] = req.Metadata
	}

	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var updated models.Item
	err := r.col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": update}, opts).Decode(&updated)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	return &updated, err
}

// Delete removes an item by ID and returns whether it existed.
func (r *ItemRepository) Delete(ctx context.Context, id primitive.ObjectID) (bool, error) {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

// TotalPages is a helper used by handlers.
func TotalPages(total int64, pageSize int) int64 {
	return int64(math.Ceil(float64(total) / float64(pageSize)))
}
