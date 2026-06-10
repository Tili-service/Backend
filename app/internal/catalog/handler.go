package catalog

import (
	"errors"
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
	catalogRoutes := router.Group("/store/:store_id/catalog")
	{
		protected := catalogRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("", h.GetAll)      // GET /store/:store_id/catalog
			protected.GET("/:id", h.GetByID) // GET /store/:store_id/catalog/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)       // POST /store/:store_id/catalog
				managerRoutes.PUT("/:id", h.Update)    // PUT /store/:store_id/catalog/:id
				managerRoutes.DELETE("/:id", h.Delete) // DELETE /store/:store_id/catalog/:id
			}
		}
	}
}

func parseStoreID(c *gin.Context) (int, bool) {
	storeID, err := strconv.Atoi(c.Param("store_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return 0, false
	}
	return storeID, true
}

// Create adds a new catalog for a store
// @Summary      Create a catalog
// @Description  Creates a new catalog in the system for a given store
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        store_id path      int           true "Store ID"  example(1)
// @Param        body     body      catalogUpdate true "catalog payload"
// @Success      201  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{store_id}/catalog [post]
func (h *Handler) Create(c *gin.Context) {
	storeID, ok := parseStoreID(c)
	if !ok {
		return
	}
	var input catalogUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catalog, err := h.service.Create(c.Request.Context(), storeID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, catalog)
}

// GetAll retrieves all catalogs for a store
// @Summary      List catalogs
// @Description  Retrieves the complete list of catalogs for a store
// @Tags         catalog
// @Produce      json
// @Security     ProfileToken
// @Param        store_id path      int  true "Store ID"  example(1)
// @Success      200  {array}   catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{store_id}/catalog [get]
func (h *Handler) GetAll(c *gin.Context) {
	storeID, ok := parseStoreID(c)
	if !ok {
		return
	}
	catalogs, err := h.service.GetAll(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalogs)
}

// GetByID retrieves a catalog by its ID within a store
// @Summary      Retrieve a catalog
// @Description  Retrieves the details of a catalog using its ID within a store
// @Tags         catalog
// @Produce      json
// @Security     ProfileToken
// @Param        store_id path      int  true "Store ID"   example(1)
// @Param        id       path      int  true "catalog ID" example(1)
// @Success      200  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /store/{store_id}/catalog/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	storeID, ok := parseStoreID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	catalog, err := h.service.GetByID(c.Request.Context(), id, storeID)
	if err != nil {
		if errors.Is(err, ErrCatalogNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "catalog not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

// Update modifies an existing catalog within a store
// @Summary      Update a catalog
// @Description  Modifies the information of an existing catalog via its ID within a store
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        store_id path      int           true "Store ID"   example(1)
// @Param        id       path      int           true "catalog ID" example(1)
// @Param        body     body      catalogUpdate true "catalog update payload"
// @Success      200  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{store_id}/catalog/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	storeID, ok := parseStoreID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input catalogUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catalog, err := h.service.Update(c.Request.Context(), id, storeID, input)
	if err != nil {
		if errors.Is(err, ErrCatalogNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "catalog not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

// Delete removes a catalog from a store
// @Summary      Delete a catalog
// @Description  Deletes a catalog from the system via its ID within a store
// @Tags         catalog
// @Produce      json
// @Security     ProfileToken
// @Param        store_id path      int  true "Store ID"   example(1)
// @Param        id       path      int  true "catalog ID" example(1)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{store_id}/catalog/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	storeID, ok := parseStoreID(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.service.Delete(c.Request.Context(), id, storeID)
	if err != nil {
		if errors.Is(err, ErrCatalogNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "catalog not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
