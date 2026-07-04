package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"tili/app/internal/middleware"
	"tili/app/internal/store"
)

var sumupOauthConfig *oauth2.Config

type Handler struct {
	storeService *store.Service
}

func NewHandler(storeService *store.Service) *Handler {
	sumupOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID_SUMUP_OAUTH"),
		ClientSecret: os.Getenv("CLIENT_SECRET_SUMUP_OAUTH"),
		RedirectURL:  "http://localhost:8000/oauth/callback",

		Scopes: []string{"transactions.history", "user.profile_readonly", "readers.read", "readers.write"},

		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.sumup.com/authorize",
			TokenURL: "https://api.sumup.com/token",
		},
	}

	return &Handler{storeService: storeService}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	oauthRoutes := router.Group("/oauth")
	accountRoutes := oauthRoutes.Group("")
	accountRoutes.Use(middleware.AccountAuthMiddleware())
	accountRoutes.GET("/login", h.login)          // GET /oauth/login
	oauthRoutes.GET("/callback", h.OAuthCallback) // GET /oauth/callback
}

func generateStateOauthCookie(c *gin.Context) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Paramètres du cookie : name, value, maxAge (en secondes), path, domain, secure, httpOnly
	// ATTENTION : 'secure' est à 'false' pour localhost. À passer absolument à 'true' en production (HTTPS) !
	c.SetCookie("oauthstate", state, 3600, "/", "localhost", false, true)

	return state
}

func setOAuthContextCookies(c *gin.Context, accountID int, storeID int) {
	c.SetCookie("oauthaccount", strconv.Itoa(accountID), 3600, "/", "localhost", false, true)
	c.SetCookie("oauthstore", strconv.Itoa(storeID), 3600, "/", "localhost", false, true)
}

func clearOAuthCookies(c *gin.Context) {
	c.SetCookie("oauthstate", "", -1, "/", "localhost", false, true)
	c.SetCookie("oauthaccount", "", -1, "/", "localhost", false, true)
	c.SetCookie("oauthstore", "", -1, "/", "localhost", false, true)
}

// login initiates the OAuth flow by redirecting the user to the SumUp authorization page
// @Summary      Initiate SumUp OAuth login
// @Description  Redirects the user to the SumUp authorization page to initiate the OAuth login process.
// @Tags         oauth
// @Accept       json
// @Produce      json
// @Router       /oauth/login [get]
func (h *Handler) login(c *gin.Context) {
	accountID := c.GetInt("accountID")
	storeID, err := strconv.Atoi(c.Query("store_id"))
	if err != nil || storeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	storeEntity, err := h.storeService.FindByID(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if storeEntity.BuyerID != accountID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the owner of this store"})
		return
	}

	setOAuthContextCookies(c, accountID, storeID)
	oauthState := generateStateOauthCookie(c)

	url := sumupOauthConfig.AuthCodeURL(oauthState)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuthCallback handles the callback from SumUp after the user authorizes the application
// @Summary      Handle SumUp OAuth callback
// @Description  Handles the callback from SumUp after the user authorizes the application and exchanges the authorization code for an access token.
// @Tags         oauth
// @Accept       json
// @Produce      json
// @Router       /oauth/callback [get]
func (h *Handler) OAuthCallback(c *gin.Context) {
	oauthState, err := c.Cookie("oauthstate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cookie d'état introuvable ou expiré"})
		return
	}

	if c.Query("state") != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "État OAuth invalide ou expiré"})
		return
	}

	accountCookie, err := c.Cookie("oauthaccount")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id introuvable ou expiré"})
		return
	}

	accountID, err := strconv.Atoi(accountCookie)
	if err != nil || accountID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	storeCookie, err := c.Cookie("oauthstore")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "store_id introuvable ou expiré"})
		return
	}

	storeID, err := strconv.Atoi(storeCookie)
	if err != nil || storeID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := sumupOauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	merchantCode := c.Query("merchant_code")
	if merchantCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_code missing"})
		return
	}

	_, err = h.storeService.LinkSumupCredentials(c.Request.Context(), storeID, accountID, merchantCode, token.AccessToken)
	if err != nil {
		if errors.Is(err, store.ErrStoreOwnershipMismatch) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, store.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearOAuthCookies(c)
	c.Redirect(http.StatusTemporaryRedirect, os.Getenv("APP_URL")+"/admin/shop/"+strconv.Itoa(storeID)+"/services-externes")
}
