package categorie

import (
	"errors"
	"net/http"

	"tili/app/internal/middleware"
	"tili/app/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	categorieRoutes := router.Group("/categorie/catalog/:catalog_id")
	{
		protected := categorieRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("", h.GetAll)               // GET /categorie/catalog/:catalog_id
			protected.GET("/type/:type", h.GetByType) // GET /categorie/catalog/:catalog_id/type/:type
			protected.GET("/:id", h.GetByID)          // GET /categorie/catalog/:catalog_id/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)                    // POST /categorie/catalog/:catalog_id
				managerRoutes.PUT("/:id", h.Update)                 // PUT /categorie/catalog/:catalog_id/:id
				managerRoutes.DELETE("/type/:type", h.DeleteByType) // DELETE /categorie/catalog/:catalog_id/type/:type
				managerRoutes.DELETE("/:id", h.DeleteByID)          // DELETE /categorie/catalog/:catalog_id/:id
			}
		}
	}
}

func parseCatalogID(c *gin.Context) (uuid.UUID, bool) {
	catalogID, err := uuid.Parse(c.Param("catalog_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid catalog_id"})
		return uuid.Nil, false
	}
	return catalogID, true
}

// Create adds a new categorie
// @Summary      Create a categorie
// @Description  Creates a new categorie in a catalog. Requires Manager level access.
// @Tags         categorie
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int       true "Catalog ID"    example(1)
// @Param        body       body      Categorie true "Categorie payload"
// @Success      201  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id} [post]
func (h *Handler) Create(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	var input Categorie
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categorie, err := h.service.Create(c.Request.Context(), catalogID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, categorie)
}

// GetAll retrieves all categories for a catalog
// @Summary      List categories
// @Description  Retrieves all categories belonging to a catalog
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int  true "Catalog ID"  example(1)
// @Success      200  {array}   Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id} [get]
func (h *Handler) GetAll(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	categories, err := h.service.FindAll(c.Request.Context(), catalogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetByID retrieves a categorie by its ID within a catalog
// @Summary      Retrieve a categorie
// @Description  Retrieves the details of a categorie using its ID within a catalog
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int  true "Catalog ID"    example(1)
// @Param        id         path      int  true "Categorie ID"  example(1)
// @Success      200  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id}/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	categorie, err := h.service.FindByID(c.Request.Context(), id, catalogID)
	if err != nil {
		if errors.Is(err, ErrCategorieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "categorie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categorie)
}

// GetByType retrieves a categorie by its type within a catalog
// @Summary      Retrieve a categorie by type
// @Description  Retrieves the details of a categorie using its type within a catalog
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int    true "Catalog ID"       example(1)
// @Param        type       path      string true "Categorie type"   example(Electronics)
// @Success      200  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id}/type/{type} [get]
func (h *Handler) GetByType(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	typ := c.Param("type")
	categorie, err := h.service.FindByType(c.Request.Context(), typ, catalogID)
	if err != nil {
		if errors.Is(err, ErrCategorieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "categorie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categorie)
}

// Update modifies an existing categorie within a catalog
// @Summary      Update a categorie
// @Description  Modifies the information of an existing categorie. Requires Manager level access.
// @Tags         categorie
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int       true "Catalog ID"    example(1)
// @Param        id         path      int       true "Categorie ID"  example(1)
// @Param        body       body      Categorie true "Categorie update payload"
// @Success      200  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id}/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input Categorie
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categorie, err := h.service.Update(c.Request.Context(), id, catalogID, input)
	if err != nil {
		if errors.Is(err, ErrCategorieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "categorie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categorie)
}

// DeleteByID removes a categorie by its ID within a catalog
// @Summary      Delete a categorie
// @Description  Deletes a categorie from a catalog via its ID. Requires Manager level access.
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int  true "Catalog ID"    example(1)
// @Param        id         path      int  true "Categorie ID"  example(1)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id}/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.service.DeleteByID(c.Request.Context(), id, catalogID)
	if err != nil {
		if errors.Is(err, ErrCategorieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "categorie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteByType removes categories by their type within a catalog
// @Summary      Delete a categorie by type
// @Description  Deletes a categorie from a catalog via its type. Requires Manager level access.
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        catalog_id path      int    true "Catalog ID"      example(1)
// @Param        type       path      string true "Categorie type"  example(Electronics)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie/catalog/{catalog_id}/type/{type} [delete]
func (h *Handler) DeleteByType(c *gin.Context) {
	catalogID, ok := parseCatalogID(c)
	if !ok {
		return
	}
	typ := c.Param("type")
	err := h.service.DeleteByType(c.Request.Context(), typ, catalogID)
	if err != nil {
		if errors.Is(err, ErrCategorieNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "categorie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
