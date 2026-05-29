package http

import (
	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/handler"
	"github.com/predicta/predicta/internal/delivery/http/middleware"
	"github.com/predicta/predicta/internal/domain/port"
)

func NewRouter(
	h *handler.Handler,
	auth *handler.AuthHandler,
	jiraSetup *handler.JiraIntegration,
	openAPIPath string,
	tokens port.TokenIssuer,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if openAPIPath != "" {
		registerSwagger(r, openAPIPath)
	}

	if auth != nil {
		auth.RegisterPublic(r)
	}

	protected := r.Group("/")
	if tokens != nil {
		protected.Use(middleware.JWTAuth(tokens))
	}
	if auth != nil {
		auth.RegisterProtected(protected)
	}
	h.Register(protected)
	if jiraSetup != nil {
		jiraSetup.Register(protected)
	}

	return r
}
