package http

import (
	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/handler"
)

func NewRouter(h *handler.Handler, jiraSetup *handler.JiraIntegration) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	h.Register(r)
	if jiraSetup != nil {
		jiraSetup.Register(r)
	}
	return r
}
