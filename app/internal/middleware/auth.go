package middleware

import (
	"net/http"

	"tili/app/internal/token"

	"github.com/gin-gonic/gin"
)

func AccountAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		claims, err := token.ValidateAccountToken(authorizationHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid account token"})
			c.Abort()
			return
		}

		c.Set("accountID", claims.AccountID)
		c.Set("customerID", claims.CustomerID)
		c.Set("name", claims.Name)
		c.Set("email", claims.Email)
		c.Next()
	}
}

func ProfileAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		claims, err := token.ValidateProfileToken(authorizationHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid profile token"})
			c.Abort()
			return
		}

		c.Set("profileID", claims.ProfileID)
		c.Set("name", claims.Name)
		c.Set("accessLevel", claims.LevelAccess)
		c.Set("storeID", claims.StoreID)
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
