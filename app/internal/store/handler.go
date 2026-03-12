package store

import (
	"net/http"
	"strconv"

	"tili/app/internal/middleware"
	"tili/app/internal/profile"
	"tili/app/internal/token"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *Service
	profileService *profile.Service
}

func NewHandler(service *Service, profileService *profile.Service) *Handler {
	return &Handler{service: service, profileService: profileService}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	storeRoutes := router.Group("/store")
	accountProtected := storeRoutes.Group("")
	accountProtected.Use(middleware.AccountAuthMiddleware())
	{
		accountProtected.POST("", h.CreateStore)       // POST /store
		accountProtected.GET("/me", h.GetMyStores)   // GET /store/me
		accountProtected.DELETE("/:id", h.DeleteStore) // DELETE /store/:id
		storeRoutes.GET("/", h.GetAll)                           // GET /store

	}
}

// GetMyStores retrieves all stores owned by the authenticated account
// @Summary      Get my stores
// @Description  Returns all stores owned by the currently authenticated account.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Success      200  {array}   Store
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/me [get]
func (h *Handler) GetMyStores(c *gin.Context) {
	accountID := c.GetInt64("accountID")
	stores, err := h.service.FindByBuyerID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stores)
}

// CreateStore creates a new store for the authenticated account
// @Summary      Create a store
// @Description  Creates a new store linked to the authenticated account. Also creates the first admin profile with an auto-generated PIN.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        body body      CreateStoreInput true "Store creation payload"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store [post]
func (h *Handler) CreateStore(c *gin.Context) {
	accountID := c.GetInt64("accountID")
	accountName := c.GetString("name")

	var input CreateStoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.BuyerID = accountID

	st, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminProfile, err := h.profileService.Create(c.Request.Context(), profile.CreateProfileInput{
		StoreID:     st.StoreID,
		Name:        accountName,
		LevelAccess: int(token.Admin),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store created but failed to create admin profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"store":   st,
		"profile": adminProfile,
	})
}

// Get all stores
// @Summary      Get all stores
// @Description  This route retrieves a list of all stores available in the system. It does not require any parameters and returns an array of store objects upon success.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Success      200  {array}   store.Store
// @Failure      500  {object}  map[string]interface{}
// @Router       /store [get]

func (h *Handler) GetAll(c *gin.Context) {
	stores, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stores)
}

// Get all stores
// @Summary      Get all stores
// @Description  This route retrieves a list of all stores available in the system. It does not require any parameters and returns an array of store objects upon success.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Success      200  {array}   store.Store
// @Failure      500  {object}  map[string]interface{}
// @Router       /store [get]

func (h *Handler) GetAll(c *gin.Context) {
	stores, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stores)
}

// DeleteStore deletes a store owned by the authenticated account
// @Summary      Delete a store
// @Description  Deletes a store and all its profiles. Only the owner can delete.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        id path int64 true "Store ID"
// @Success      204
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{id} [delete]
func (h *Handler) DeleteStore(c *gin.Context) {
	accountID := c.GetInt64("accountID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store ID"})
		return
	}

	st, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	if st.BuyerID != accountID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the owner of this store"})
		return
	}

	if err := h.profileService.DeleteByStoreID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
