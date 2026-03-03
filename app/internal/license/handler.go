package license

import (
	"net/http"
	"fmt"
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
		registrationRoutes.POST("", h.Create) // POST /registration/account
		registrationRoutes.DELETE("", h.Delete) // DELETE /registration/account
	}
}

// Create a new user a licence and a shop
// @Summary      Create a new user a licence and a shop
// @Description  This route creates a new user, a licence, and a shop in admin mode, linking them together. It accepts a JSON payload with the necessary information for each entity and returns the created user object upon success.
// @Tags         licenses
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

// Delete a user a licence and a shop
// @Summary      Delete a user a licence and a shop using the account ID
// @Description  This route deletes an existing user, licence, and shop in admin mode, unlinking them together. It accepts an account ID and returns a success message upon deletion.
// @Tags         licenses
// @Accept       json
// @Produce      json
// @Param        body body      AccountDeleting true "Account deletion payload"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /registration/account [delete]
func (h *Handler) Delete(c *gin.Context) {
	var input AccountDeleting
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exist, err := h.service.Exists(c.Request.Context(), input.AccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	err = h.service.FullDeleteAccount(c.Request.Context(), input.AccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

