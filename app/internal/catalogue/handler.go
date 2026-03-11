package catalogue

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
	catalogueRoutes := router.Group("/catalogue")
	{
		protected := catalogueRoutes.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", h.GetAll)        // GET /catalogue
			protected.GET("/:id", h.GetByID)   // GET /catalogue/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)       // POST /catalogue
				managerRoutes.PUT("/:id", h.Update)    // PUT /catalogue/:id
				managerRoutes.DELETE("/:id", h.Delete) // DELETE /catalogue/:id
			}
		}
	}
}

// Create adds a new catalogue
// @Summary      Create a catalogue
// @Description  Creates a new catalogue in the system
// @Tags         catalogue
// @Accept       json
// @Produce      json
// @Param        body body      CatalogueUpdate true "Catalogue payload"
// @Success      201  {object}  Catalogue
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalogue [post]
func (h *Handler) Create(c *gin.Context) {
	var input CatalogueUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catalogue, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, catalogue)
}

// GetAll retrieves the list of all catalogues
// @Summary      List catalogues
// @Description  Retrieves the complete list of catalogues
// @Tags         catalogue
// @Produce      json
// @Success      200  {array}   Catalogue
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalogue [get]
func (h *Handler) GetAll(c *gin.Context) {
	catalogues, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalogues)
}

// GetByID retrieves a catalogue by its ID
// @Summary      Retrieve a catalogue
// @Description  Retrieves the details of a catalogue using its ID
// @Tags         catalogue
// @Produce      json
// @Param        id   path      int  true  "Catalogue ID"
// @Success      200  {object}  Catalogue
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /catalogue/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	catalogue, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "catalogue not found"})
		return
	}
	c.JSON(http.StatusOK, catalogue)
}

// Update modifies an existing catalogue
// @Summary      Update a catalogue
// @Description  Modifies the information of an existing catalogue via its ID
// @Tags         catalogue
// @Accept       json
// @Produce      json
// @Param        id   path      int             true "Catalogue ID"
// @Param        body body      CatalogueUpdate true "Catalogue update payload"
// @Success      200  {object}  Catalogue
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalogue/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input CatalogueUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	catalogue, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalogue)
}

// Delete removes a catalogue
// @Summary      Delete a catalogue
// @Description  Deletes a catalogue from the system via its ID
// @Tags         catalogue
// @Produce      json
// @Param        id   path      int  true  "Catalogue ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /catalogue/{id} [delete]
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
