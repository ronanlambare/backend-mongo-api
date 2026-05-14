package handler

import (
	"errors"
	"net/http"

	"github.com/ronanlambare/backend-mongo-api/internal/model"
	"github.com/ronanlambare/backend-mongo-api/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ItemHandler struct {
	repo *repository.ItemRepository
}

func NewItemHandler(repo *repository.ItemRepository) *ItemHandler {
	return &ItemHandler{repo: repo}
}

// List godoc
// @Summary     List all items
// @Tags        items
// @Produce     json
// @Security    BearerAuth
// @Success     200  {array}   model.Item
// @Failure     500  {object}  model.ErrorResponse
// @Router      /items [get]
func (h *ItemHandler) List(c *gin.Context) {
	items, err := h.repo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary     Get an item by ID
// @Tags        items
// @Produce     json
// @Security    BearerAuth
// @Param       id   path      string  true  "Item ID (MongoDB ObjectID)"
// @Success     200  {object}  model.Item
// @Failure     400  {object}  model.ErrorResponse
// @Failure     404  {object}  model.ErrorResponse
// @Router      /items/{id} [get]
func (h *ItemHandler) GetByID(c *gin.Context) {
	item, err := h.repo.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Create godoc
// @Summary     Create a new item
// @Tags        items
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body      model.CreateItemRequest true "Item payload"
// @Success     201  {object}  model.Item
// @Failure     400  {object}  model.ErrorResponse
// @Failure     500  {object}  model.ErrorResponse
// @Router      /items [post]
func (h *ItemHandler) Create(c *gin.Context) {
	var req model.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	item, err := h.repo.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Update godoc
// @Summary     Update an item
// @Tags        items
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path      string                 true "Item ID (MongoDB ObjectID)"
// @Param       body body      model.UpdateItemRequest true "Item payload"
// @Success     200  {object}  model.Item
// @Failure     400  {object}  model.ErrorResponse
// @Failure     404  {object}  model.ErrorResponse
// @Failure     500  {object}  model.ErrorResponse
// @Router      /items/{id} [put]
func (h *ItemHandler) Update(c *gin.Context) {
	var req model.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	item, err := h.repo.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary     Delete an item
// @Tags        items
// @Produce     json
// @Security    BearerAuth
// @Param       id  path string true "Item ID (MongoDB ObjectID)"
// @Success     204 "No Content"
// @Failure     400 {object} model.ErrorResponse
// @Failure     404 {object} model.ErrorResponse
// @Router      /items/{id} [delete]
func (h *ItemHandler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

