package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lambare/go-mongo-api/internal/models"
	"github.com/lambare/go-mongo-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemHandler struct {
	repo *repository.ItemRepository
}

func NewItemHandler(repo *repository.ItemRepository) *ItemHandler {
	return &ItemHandler{repo: repo}
}

// List godoc
// GET /api/v1/items?page=1&page_size=20&search=foo&tags=a,b
func (h *ItemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	var tags []string
	if raw := c.Query("tags"); raw != "" {
		tags = strings.Split(raw, ",")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.repo.List(c.Request.Context(), page, pageSize, search, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch items"})
		return
	}

	// Return empty array instead of null
	if items == nil {
		items = []*models.Item{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: repository.TotalPages(total, pageSize),
	})
}

// GetByID godoc
// GET /api/v1/items/:id
func (h *ItemHandler) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch item"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Create godoc
// POST /api/v1/items
func (h *ItemHandler) Create(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := primitive.ObjectIDFromHex(userIDStr.(string))

	item := &models.Item{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedBy:   userID,
	}

	if err := h.repo.Create(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// Update godoc
// PUT /api/v1/items/:id
func (h *ItemHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item"})
		return
	}
	if updated == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// Delete godoc
// DELETE /api/v1/items/:id
func (h *ItemHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	deleted, err := h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
