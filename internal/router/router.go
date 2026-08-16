package router

import (
	"github.com/gin-gonic/gin"

	"bookmark-api/internal/handler"
	"bookmark-api/internal/middleware"
)

func New(h *handler.Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), middleware.Recovery())

	r.GET("/healthz", h.Health)

	v1 := r.Group("/api/v1")
	{
		bookmarks := v1.Group("/bookmarks")
		{
			bookmarks.POST("", h.Create)
			bookmarks.GET("", h.List)
			bookmarks.GET("/:id", h.Get)
			bookmarks.PUT("/:id", h.Update)
			bookmarks.DELETE("/:id", h.Delete)
			bookmarks.POST("/:id/click", h.IncrementClickCount)
		}
	}

	return r
}
