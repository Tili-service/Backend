package sale

import (
	"errors"
	"net/http"
	"strconv"
	"tili/app/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.Engine) {
	sales := rg.Group("/sales")
	accountProtected := sales.Group("")
	accountProtected.Use(middleware.AccountAuthMiddleware())
	{
		accountProtected.POST("", h.CreateSale)
		accountProtected.GET("", h.GetAllSales)
		accountProtected.GET("/:id", h.GetSaleByID)
		accountProtected.PUT("/:id", h.UpdateSale)
		accountProtected.DELETE("/:id", h.DeleteSale)
	}
}

// @Summary      Create a new sale
// @Tags         sales
// @Accept       json
// @Produce      json
// @Param        sale  body      CreateSaleInput  true  "Sale payload"
// @Success      201   {object}  Sale
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /sales [post]
func (h *Handler) CreateSale(c *gin.Context) {
	var input CreateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sale, err := h.service.CreateSale(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidSaleTotal) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, sale)
}

// @Summary      List all sales
// @Tags         sales
// @Produce      json
// @Success      200  {array}   Sale
// @Failure      500  {object}  map[string]string
// @Router       /sales [get]
func (h *Handler) GetAllSales(c *gin.Context) {
	sales, err := h.service.GetAllSales(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

// @Summary      Get a sale by ID
// @Tags         sales
// @Produce      json
// @Param        id   path      int  true  "Sale ID"
// @Success      200  {object}  Sale
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sales/{id} [get]
func (h *Handler) GetSaleByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	sale, err := h.service.GetSaleByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrSaleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, sale)
}

// @Summary      Update a sale
// @Tags         sales
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Sale ID"
// @Param        sale  body      UpdateSaleInput  true  "Fields to update"
// @Success      200   {object}  Sale
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /sales/{id} [put]
func (h *Handler) UpdateSale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	var input UpdateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var changedBy *int
	if accountID := c.GetInt("accountID"); accountID > 0 {
		changedBy = &accountID
	}

	sale, err := h.service.UpdateSale(c.Request.Context(), id, input, changedBy)
	if err != nil {
		if errors.Is(err, ErrSaleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrInvalidSaleTotal) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, sale)
}

// @Summary      Delete a sale
// @Tags         sales
// @Param        id   path      int  true  "Sale ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sales/{id} [delete]
func (h *Handler) DeleteSale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	if err := h.service.DeleteSale(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrSaleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}
