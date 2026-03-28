package api

import (
	"github.com/gin-gonic/gin"
	"github.com/teerakarna/service-demo/internal/store"
)

func NewRouter(s store.Store) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	h := NewHandler(s)

	r.GET("/healthz", h.Healthz)
	r.GET("/readyz", h.Readyz)

	v1 := r.Group("/api/v1")
	{
		items := v1.Group("/items")
		items.GET("", h.ListItems)
		items.POST("", h.CreateItem)
		items.GET("/:id", h.GetItem)
		items.PUT("/:id", h.UpdateItem)
		items.DELETE("/:id", h.DeleteItem)
	}

	return r
}
