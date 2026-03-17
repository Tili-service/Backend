package payementmethod

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
	pmRoutes := router.Group("/payementmethod")
	{
		protected := pmRoutes.Group("")
		protected.Use(middleware.ProfileAuthMiddleware())
		{
			protected.GET("", h.GetAll)               // GET /payementmethod
			protected.GET("/name/:name", h.GetByName) // GET /payementmethod/name/:name

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.POST("", h.Create)       // POST /payementmethod
				managerRoutes.PUT("/:id", h.Update)    // PUT /payementmethod/:id
				managerRoutes.DELETE("/:id", h.Delete) // DELETE /payementmethod/:id
			}
		}
	}
}

// Create adds a new payement method
// @Summary      Create a payement method
// @Description  Creates a new payement method in the system. Requires Manager level access.
// @Tags         payementmethod
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        body body      PayementMethodInput true "Payement method payload"
// @Success      201  {object}  PayementMethod
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /payementmethod [post]
func (h *Handler) Create(c *gin.Context) {
	var input PayementMethodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pm, err := h.service.Create(c.Request.Context(), PayementMethod{Name: input.Name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pm)
}

// Update modifies an existing payement method
// @Summary      Update a payement method
// @Description  Updates an existing payement method in the system. Requires Manager level access.
// @Tags         payementmethod
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int                  true "Payement method ID"   example(1)
// @Param        body body      PayementMethodInput  true "Payement method payload"
// @Success      200  {object}  PayementMethod
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /payementmethod/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input PayementMethodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pm, err := h.service.Update(c.Request.Context(), id, PayementMethod{Name: input.Name})
	if err != nil {
		if errors.Is(err, ErrPayementMethodNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payement method not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pm)
}

// Delete removes a payement method by ID
// @Summary      Delete a payement method
// @Description  Deletes a payement method from the system by its ID. Requires Manager level access.
// @Tags         payementmethod
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        id   path      int  true "Payement method ID"   example(1)
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /payementmethod/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteByID(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrPayementMethodNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payement method not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetAll retrieves all payement methods
// @Summary      Get all payement methods
// @Description  Retrieves a list of all payement methods in the system. Requires authentication.
// @Tags         payementmethod
// @Produce      json
// @Security     ProfileToken
// @Success      200  {array}   PayementMethod
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /payementmethod [get]
func (h *Handler) GetAll(c *gin.Context) {
	payementMethods, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payementMethods)
}

// GetByName retrieves a payement method by its name
// @Summary      Get a payement method by name
// @Description  Retrieves a payement method from the system by its name. Requires authentication.
// @Tags         payementmethod
// @Produce      json
// @Security     ProfileToken
// @Param        name   path      string true "Payement method name"   example(Credit Card)
// @Success      200  {object}  PayementMethod
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /payementmethod/name/{name} [get]
func (h *Handler) GetByName(c *gin.Context) {
	name := c.Param("name")
	pm, err := h.service.GetByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, ErrPayementMethodNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payement method not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pm)
}
