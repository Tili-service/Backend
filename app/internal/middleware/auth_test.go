package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/internal/token"
)

func TestAccountAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing header", func(t *testing.T) {
		r := gin.New()
		r.Use(AccountAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		r := gin.New()
		r.Use(AccountAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		accountID := uuid.New()
		tok, err := token.CreateAccountToken(accountID, "A", "a@a.com", "cus_1")
		assert.NoError(t, err)

		r := gin.New()
		r.Use(AccountAuthMiddleware())
		r.GET("/", func(c *gin.Context) {
			assert.Equal(t, accountID.String(), c.GetString("accountID"))
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProfileAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing header", func(t *testing.T) {
		r := gin.New()
		r.Use(ProfileAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		r := gin.New()
		r.Use(ProfileAuthMiddleware())
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		profileID := uuid.New()
		storeID := uuid.New()
		tok, err := token.CreateProfileToken(profileID, "P", int(token.Manager), storeID)
		assert.NoError(t, err)

		r := gin.New()
		r.Use(ProfileAuthMiddleware())
		r.GET("/", func(c *gin.Context) {
			assert.Equal(t, profileID.String(), c.GetString("profileID"))
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestLevelAccessRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("access level missing", func(t *testing.T) {
		r := gin.New()
		r.Use(LevelAccessRequired(token.Manager))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("accessLevel", int(token.UserLevel))
			c.Next()
		})
		r.Use(LevelAccessRequired(token.Manager))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("allowed", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("accessLevel", int(token.Manager))
			c.Next()
		})
		r.Use(LevelAccessRequired(token.Manager))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
