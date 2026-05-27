package http

import (
	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/adapter/http/handler"
)

func NewRouter(h *handler.Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	h.Register(r)
	return r
}
