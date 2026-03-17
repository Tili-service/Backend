package item

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
	itemRoutes := router.Group("/item")
	{
		protected := itemRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("", h.GetAll)                         // GET /item
			protected.GET("/name/:name", h.GetByName)           // GET /item/name/:name — must be before /:id
			protected.GET("/:id", h.GetByID)                    // GET /item/:id
			protected.GET("/categorie/:id", h.GetByCategorieID) // GET /item/categorie/:id

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)       // POST /item
				managerRoutes.PUT("/:id", h.Update)    // PUT /item/:id
				managerRoutes.DELETE("/:id", h.Delete) // DELETE /item/:id
			}
		}
	}
}

// Create adds a new item
// @Summary      Create an item
// @Description  Creates a new item in the system. Requires Manager level access.
// @Tags         item
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        body body      Item true "Item payload"
// @Success      201  {object}  Item
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item [post]
func (h *Handler) Create(c *gin.Context) {
	var input Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Update modifies an existing item
// @Summary      Update an item
// @Description  Modifies the information of an existing item via its ID. Requires Manager level access.
// @Tags         item
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int        true "Item ID"     example(1)
// @Param        body body      ItemUpdate true "Item update payload"
// @Success      200  {object}  Item
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input ItemUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete removes an item
// @Summary      Delete an item
// @Description  Deletes an item from the system via its ID. Requires Manager level access.
// @Tags         item
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true  "Item ID"  example(1)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetAll retrieves the list of all items
// @Summary      List items
// @Description  Retrieves the complete list of items
// @Tags         item
// @Produce      json
// @Security     ProfileToken
// @Success      200  {array}   Item
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item [get]
func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID retrieves an item by its ID
// @Summary      Retrieve an item
// @Description  Retrieves the details of an item using its ID
// @Tags         item
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true  "Item ID"  example(1)
// @Success      200  {object}  Item
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// GetByName retrieves an item by its name
// @Summary      Retrieve an item by name
// @Description  Retrieves the details of an item using its name
// @Tags         item
// @Produce      json
// @Security     ProfileToken
// @Param        name path      string  true  "Item name"  example(Laptop Pro 15)
// @Success      200  {object}  Item
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item/name/{name} [get]
func (h *Handler) GetByName(c *gin.Context) {
	name := c.Param("name")
	item, err := h.service.GetByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// GetByCategorieID retrieves items by their category ID
// @Summary      Retrieve items by category
// @Description  Retrieves the list of items belonging to a specific category using the category ID
// @Tags         item
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true  "Category ID"  example(1)
// @Success      200  {array}   Item
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /item/categorie/{id} [get]
func (h *Handler) GetByCategorieID(c *gin.Context) {
	categorieID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}
	items, err := h.service.GetByCategorieID(c.Request.Context(), categorieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []Item{}
	}
	c.JSON(http.StatusOK, items)
}
