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
				managerRoutes.POST("", h.Create)    // POST /profile
				managerRoutes.PUT("/:id", h.Update) // PUT /profile/:id
				managerRoutes.GET("/allProfilesByStoreId/:id", h.GetProfilesByStoreId) // GET /profile/allProfilesByStoreId/:id
				managerRoutes.PUT("/updateProfile/:id/:storeId", h.UpdateProfileByIdAndStoreId) // PUT /profile/updateProfile/:id/:storeId
				managerRoutes.PUT("/deactivateProfile/:id/:storeId", h.DeactivateProfile) // PUT /profile/deactivateProfile/:id/:storeId
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

// Update modifies an existing profile
// @Summary      Update a profile
// @Description  Updates profile fields (name, PIN, level access, active status). Requires manager+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int                true  "Profile ID"
// @Param        body body      updateProfileInput true  "Profile update payload"
// @Success      200  {object}  Profile
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /profile/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// GetProfilesByStoreId retrieves all profiles for a given store ID
// @Summary      Get profiles by store ID
// @Description  Retrieves a list of profiles belonging to the specified store. Requires manager+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true  "Store ID"
// @Success      200  {array}   Profile
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /profile/allProfilesByStoreId/{id} [get]
func (h *Handler) GetProfilesByStoreId(c *gin.Context) {
	storeId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store ID"})
		return
	}

	profiles, err := h.service.GetProfilesByStoreId(c.Request.Context(), storeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// UpdateProfileByIdAndStoreId modifies an existing profile using its ID and store ID
// @Summary      Update a profile by ID and store ID
// @Description  Updates profile fields (name, PIN, level access, active status) for a specific store. Requires manager+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id      path      int                true  "Profile ID"
// @Param        storeId path      int                true  "Store ID"
// @Param        body    body      updateProfileInput true  "Profile update payload"
// @Success      200     {object}  Profile
// @Failure      400     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /profile/updateProfile/{id}/{storeId} [put]
func (h *Handler) UpdateProfileByIdAndStoreId(c *gin.Context) {
	idProfile, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile ID"})
		return
	}
	storeId, err := strconv.Atoi(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store ID"})
		return
	}
	var input updateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := h.service.UpdateProfileByIdAndStoreId(c.Request.Context(), idProfile, storeId, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// DeactivateProfile deactivates a profile by its ID and store ID
// @Summary      Deactivate a profile
// @Description  Sets the is_active status of a profile to false. Requires manager+ access.
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id      path      int  true  "Profile ID"
// @Param        storeId path      int  true  "Store ID"
// @Success      200     {object}  Profile
// @Failure      400     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /profile/deactivateProfile/{id}/{storeId} [put]
func (h *Handler) DeactivateProfile(c *gin.Context) {
	idProfile, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile ID"})
		return
	}
	storeId, err := strconv.Atoi(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store ID"})
		return
	}

	input := updateProfileInput{
		IsActive: func(b bool) *bool { return &b }(false),
	}

	profile, err := h.service.UpdateProfileByIdAndStoreId(c.Request.Context(), idProfile, storeId, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}