package account

import (
	"errors"
	"net/http"

	"tili/app/internal/middleware"
	"tili/app/internal/token"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	accountRoutes := router.Group("/account")
	{
		accountRoutes.POST("", h.Create)      // POST /account
		accountRoutes.POST("/login", h.Login) // POST /account/login

		protected := accountRoutes.Group("")
		protected.Use(middleware.AccountAuthMiddleware())
		{
			protected.GET("", h.GetAccount) // GET /account
			protected.DELETE("", h.Delete)  // DELETE /account
			protected.PUT("", h.Update)     // PUT /account
		}
	}
}

// Create registers a new account
// @Summary      Register a new account
// @Description  Creates a new account. Returns the created account.
// @Tags         account
// @Accept       json
// @Produce      json
// @Param        body body      RegistrationInput true "Account registration payload"
// @Success      201  {object}  Account
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /account [post]
func (h *Handler) Create(c *gin.Context) {
	var input RegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, acc)
}

// Login authenticates an account and returns a token with store list
// @Summary      Account login
// @Description  Authenticates with email/password and returns an account JWT token along with the list of owned stores.
// @Tags         account
// @Accept       json
// @Produce      json
// @Param        body body      LoginInput true "Login payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /account/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, stores, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accountToken, err := token.CreateAccountToken(acc.AccountID, acc.Name, acc.Email, acc.StripeCustomerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   accountToken,
		"account": acc,
		"stores":  stores,
	})
}

// GetAccount retrieves the current account information
// @Summary      Get current account
// @Description  Retrieves the account information of the currently authenticated user.
// @Tags         account
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Success      200  {object}  Account
// @Failure      500  {object}  map[string]interface{}
// @Router       /account [get]
func (h *Handler) GetAccount(c *gin.Context) {
	accountID := c.GetInt("accountID")
	acc, err := h.service.GetByID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve account"})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// Delete removes the current account and all related data
// @Summary      Delete account
// @Description  Deletes the currently authenticated account along with all licences, stores, and profiles.
// @Tags         account
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Success      204
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /account [delete]
func (h *Handler) Delete(c *gin.Context) {
	accountID := c.GetInt("accountID")

	exist, err := h.service.Exists(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	if err := h.service.FullDelete(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Update modifies the current account information
// @Summary      Update account
// @Description  Updates the name and/or email of the currently authenticated account.
// @Tags         account
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        body body      UpdateAccountInput true "Account update payload"
// @Success      200  {object}  Account
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /account [put]
func (h *Handler) Update(c *gin.Context) {
	accountID := c.GetInt("accountID")
	var input UpdateAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.service.Update(c.Request.Context(), accountID, input)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}
