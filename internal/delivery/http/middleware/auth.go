package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/domain/port"
)

const ManagerIDKey = "manager_id"

func JWTAuth(tokens port.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "authorization required"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" || token == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid authorization header"})
			return
		}

		managerID, err := tokens.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		c.Set(ManagerIDKey, managerID)
		c.Next()
	}
}

func ManagerID(c *gin.Context) string {
	v, _ := c.Get(ManagerIDKey)
	id, _ := v.(string)
	return id
}
