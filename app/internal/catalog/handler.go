package catalog

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
	catalogRoutes := router.Group("/catalog")
	{
		protected := catalogRoutes.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", h.GetAll)      // GET /catalog
			protected.GET("/:id", h.GetByID) // GET /catalog/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)       // POST /catalog
				managerRoutes.PUT("/:id", h.Update)    // PUT /catalog/:id
				managerRoutes.DELETE("/:id", h.Delete) // DELETE /catalog/:id
			}
		}
	}
}

// Create adds a new catalog
// @Summary      Create a catalog
// @Description  Creates a new catalog in the system
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        body body      catalogUpdate true "catalog payload"
// @Success      201  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalog [post]
func (h *Handler) Create(c *gin.Context) {
	var input catalogUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catalog, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, catalog)
}

// GetAll retrieves the list of all catalogs
// @Summary      List catalogs
// @Description  Retrieves the complete list of catalogs
// @Tags         catalog
// @Produce      json
// @Success      200  {array}   catalog
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalog [get]
func (h *Handler) GetAll(c *gin.Context) {
	catalogs, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalogs)
}

// GetByID retrieves a catalog by its ID
// @Summary      Retrieve a catalog
// @Description  Retrieves the details of a catalog using its ID
// @Tags         catalog
// @Produce      json
// @Param        id   path      int  true  "catalog ID"
// @Success      200  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /catalog/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	catalog, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "catalog not found"})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

// Update modifies an existing catalog
// @Summary      Update a catalog
// @Description  Modifies the information of an existing catalog via its ID
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        id   path      int             true "catalog ID"
// @Param        body body      catalogUpdate true "catalog update payload"
// @Success      200  {object}  catalog
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalog/{id} [put]
func (h *Handler) Update(c *gin.Context) {
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
	catalog, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

// Delete removes a catalog
// @Summary      Delete a catalog
// @Description  Deletes a catalog from the system via its ID
// @Tags         catalog
// @Produce      json
// @Param        id   path      int  true  "catalog ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalog/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
