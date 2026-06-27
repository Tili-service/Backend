package oauth

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

var sumupOauthConfig *oauth2.Config

const oauthStateString = "randomstatestring12347"

type Handler struct{}

func NewHandler() *Handler {
	sumupOauthConfig = &oauth2.Config{
		ClientID:     "cc_classic_mt65Wke1nSvUocIDaQoZFYTGXR1p1",
		ClientSecret: "cc_sk_classic_bRZOn42iUXubavOWFFuWx36eN7dFk60g8J32Ly2DwoH4h1MNmL",
		RedirectURL:  "http://localhost:8000/oauth/callback",

		Scopes: []string{"readers.read", "readers.write", "user.profile_readonly"},

		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.sumup.com/authorize",
			TokenURL: "https://api.sumup.com/token",
		},
	}

	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	accountRoutes := router.Group("/oauth")
	{
		accountRoutes.GET("login", h.login)            // GET /oauth/login
		accountRoutes.GET("callback", h.OAuthCallback) // GET /oauth/callback
	}
}

// login initiates the OAuth flow by redirecting the user to the SumUp authorization page
// @Summary      Initiate SumUp OAuth login
// @Description  Redirects the user to the SumUp authorization page to initiate the OAuth login process.
// @Tags         oauth
// @Accept       json
// @Produce      json
// @Router       /oauth/login [get]
func (h *Handler) login(c *gin.Context) {
	url := sumupOauthConfig.AuthCodeURL(oauthStateString)
	log.Println("👉 DEBUG URL :", url)
	log.Println("👉 DEBUG CLIENT_ID :", os.Getenv("CLIENT_ID_SUMUP_OAUTH"))
	log.Println("👉 DEBUG CLIENT_SECRET :", os.Getenv("CLIENT_SECRET_SUMUP_OAUTH"))
	log.Println("👉 DEBUG URL :", oauthStateString)
	// CORRECTION IMPORTANTE : Utiliser 307 ou 302, jamais 301 pour l'OAuth !
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
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	code := c.Query("code")
	token, err := sumupOauthConfig.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	accessToken := token.AccessToken


	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}
