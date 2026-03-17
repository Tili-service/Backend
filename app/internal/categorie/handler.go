package categorie

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
	categorieRoutes := router.Group("/categorie")
	{
		protected := categorieRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("", h.GetAll)               // GET /categorie
			protected.GET("/type/:type", h.GetByType) // GET /categorie/type/:type — must be before /:id
			protected.GET("/:id", h.GetByID)          // GET /categorie/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)                    // POST /categorie
				managerRoutes.PUT("/:id", h.Update)                 // PUT /categorie/:id
				managerRoutes.DELETE("/type/:type", h.DeleteByType) // DELETE /categorie/type/:type — must be before /:id
				managerRoutes.DELETE("/:id", h.DeleteByID)          // DELETE /categorie/:id
			}
		}
	}
}

// Create adds a new categorie
// @Summary      Create a categorie
// @Description  Creates a new categorie in the system. Requires Manager level access.
// @Tags         categorie
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        body body      Categorie true "Categorie payload"
// @Success      201  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie [post]
func (h *Handler) Create(c *gin.Context) {
	var input Categorie
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categorie, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, categorie)
}

// GetAll retrieves the list of all categories
// @Summary      List categories
// @Description  Retrieves the complete list of categories
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Success      200  {array}   Categorie
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /categorie [get]
func (h *Handler) GetAll(c *gin.Context) {
	categories, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetByID retrieves a categorie by its ID
// @Summary      Retrieve a categorie
// @Description  Retrieves the details of a categorie using its ID
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true  "Categorie ID"  example(1)
// @Success      200  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categorie/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	categorie, err := h.service.FindByID(c.Request.Context(), id)
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

// GetByType retrieves a categorie by its type
// @Summary      Retrieve a categorie by type
// @Description  Retrieves the details of a categorie using its type
// @Tags         categorie
// @Produce      json
// @Security     ProfileToken
// @Param        type path      string  true  "Categorie type"  example(Electronics)
// @Success      200  {object}  Categorie
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /categorie/type/{type} [get]
func (h *Handler) GetByType(c *gin.Context) {
	typ := c.Param("type")
	categorie, err := h.service.FindByType(c.Request.Context(), typ)
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

// Update modifies an existing categorie
// @Summary      Update a categorie
// @Description  Modifies the information of an existing categorie via its ID. Requires Manager level access.
// @Tags         categorie
// @Accept       json
// @Produce      json
// @Param        id   path      int       true "Categorie ID"  example(1)
// @Param        body body      Categorie true "Categorie update payload"
// @Success      200  {object}  Categorie
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     ProfileToken
// @Router       /categorie/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input Categorie
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categorie, err := h.service.Update(c.Request.Context(), id, input)
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

// DeleteByID removes a categorie by its ID
// @Summary      Delete a categorie
// @Description  Deletes a categorie from the system via its ID. Requires Manager level access.
// @Tags         categorie
// @Produce      json
// @Param        id   path      int  true  "Categorie ID"  example(1)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     ProfileToken
// @Router       /categorie/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = h.service.DeleteByID(c.Request.Context(), id)
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

// DeleteByType removes categories by their type
// @Summary      Delete a categorie by type
// @Description  Deletes a categorie from the system via its type. Requires Manager level access.
// @Tags         categorie
// @Produce      json
// @Param        type path      string  true  "Categorie type"  example(Electronics)
// @Success      204  {object}  nil
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     ProfileToken
// @Router       /categorie/type/{type} [delete]
func (h *Handler) DeleteByType(c *gin.Context) {
	typ := c.Param("type")
	err := h.service.DeleteByType(c.Request.Context(), typ)
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
