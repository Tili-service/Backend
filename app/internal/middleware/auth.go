package middleware

import (
	"net/http"

	"tili/app/internal/token"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenToSend, err := token.Validate(authorizationHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", tokenToSend.UserID)
		c.Set("name", tokenToSend.Name)
		c.Set("email", tokenToSend.Email)
		c.Set("accessLevel", tokenToSend.AccessLevel)
		c.Next()
	}
}

func LevelAccessRequired(level token.AccessLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessLevel := c.GetInt("accessLevel")

		if accessLevel == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "access level not found in token"})
			c.Abort()
			return
		}
		if accessLevel > int(level) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
