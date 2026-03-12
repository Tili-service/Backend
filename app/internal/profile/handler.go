package profile

import (
	"net/http"
	"strconv"

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
	profileRoutes := router.Group("/profile")
	{
		profileRoutes.POST("/login/pin", h.loginWithPin) // POST /profile/login/pin

		protected := profileRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("/me", h.me) // GET /profile/me

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create) // POST /profile
			}

			adminRoutes := protected.Group("")
			adminRoutes.Use(middleware.LevelAccessRequired(token.Admin))
			{
				adminRoutes.DELETE("/:id", h.Delete) // DELETE /profile/:id
			}
		}
	}
}

// loginWithPin authenticates a profile by PIN code within a store
// @Summary      Login with PIN
// @Description  Authenticates a profile using a PIN code and a store ID.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        body body      PinLoginInput true "PIN login payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /profile/login/pin [post]
func (h *Handler) loginWithPin(c *gin.Context) {
	var input PinLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.service.LoginWithPin(c.Request.Context(), input.StoreID, input.Pin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid pin"})
		return
	}

	profileToken, err := token.CreateProfileToken(p.ProfileID, p.Name, p.LevelAccess, input.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": profileToken, "profile": p})
}

// me retrieves the current authenticated profile
// @Summary      Get current profile
// @Description  Retrieves the profile information of the currently authenticated user.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Success      200  {object}  Profile
// @Failure      500  {object}  map[string]interface{}
// @Router       /profile/me [get]
func (h *Handler) me(c *gin.Context) {
	profileID := c.GetInt("profileID")
	p, err := h.service.GetByID(c.Request.Context(), profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve profile"})
		return
	}
	c.JSON(http.StatusOK, p)
}

// Create adds a new profile to the current store
// @Summary      Create a profile
// @Description  Creates a new profile in the current store. Requires manager+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        body body      CreateProfileInput true "Profile creation payload"
// @Success      201  {object}  ProfileWithPin
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /profile [post]
func (h *Handler) Create(c *gin.Context) {
	storeID := c.GetInt("storeID")
	var input CreateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.StoreID = storeID

	p, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Delete removes a profile
// @Summary      Delete a profile
// @Description  Deletes a profile by its ID. Requires admin+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id path int true "Profile ID"
// @Success      204
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /profile/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
