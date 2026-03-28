package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teerakarna/service-demo/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

type itemRequest struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

func (h *Handler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Readyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) ListItems(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.List())
}

func (h *Handler) GetItem(c *gin.Context) {
	item, ok := h.store.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := h.store.Create(req.Name, req.Description)
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateItem(c *gin.Context) {
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, ok := h.store.Update(c.Param("id"), req.Name, req.Description)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	if !h.store.Delete(c.Param("id")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
