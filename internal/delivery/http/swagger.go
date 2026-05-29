package http

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

func registerSwagger(r *gin.Engine, openAPIPath string) {
	// Без no-cache браузер часто показывает старый openapi.yaml после деплоя.
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File(openAPIPath)
	})

	ui, err := fs.Sub(swaggerUI, "swagger-ui")
	if err != nil {
		return
	}
	r.StaticFS("/swagger", http.FS(ui))

	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/")
	})
}
