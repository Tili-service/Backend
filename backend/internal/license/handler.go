package license

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	registrationRoutes := router.Group("/registration/account")
	{
		registrationRoutes.POST("", h.Create) // POST /registration/account create licence, shop, user en mode admin lier a tout ça <3
	}
}

// Create a new user a licence and a shop
// @Summary      Create a new user a licence and a shop
// @Description  This route creates a new user, a licence, and a shop in admin mode, linking them together. It accepts a JSON payload with the necessary information for each entity and returns the created user object upon success.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body      AccountRegistrationInput true "Account registration payload"
// @Success      201  {object}  account.Account
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /registration/account [post]
func (h *Handler) Create(c *gin.Context) {
	var input AccountRegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}
