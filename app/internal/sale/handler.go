package sale

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

func (h *Handler) RegisterRoutes(rg *gin.Engine) {
	sales := rg.Group("/sales")
	protected := sales.Group("")
	protected.Use(middleware.ProfileAuthMiddleware())
	{
		protected.POST("", h.CreateSale)
		protected.GET("", h.GetAllSales)
		protected.GET("/:id", h.GetSaleByID)
		managerRoutes := protected.Group("")
		managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
		{
			managerRoutes.PUT("/:id", h.UpdateSale)
			managerRoutes.DELETE("/:id", h.DeleteSale)
		}
	}
}

// @Summary      Create a new sale
// @Description  Creates a sale with one or more line items. The total is computed from unit prices × quantities. Returns 400 if the computed total is not positive.
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

	var changedByProf *int
	if profileID := c.GetInt("profileID"); profileID > 0 {
		changedByProf = &profileID
	}

	sale, err := h.service.CreateSale(c.Request.Context(), input, changedByProf)
	if err != nil {
		if errors.Is(err, ErrInvalidSaleTotal) || errors.Is(err, ErrInvalidPaymentsTotal) || errors.Is(err, ErrInvalidPaymentAmount) || errors.Is(err, ErrPayementMethodInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, sale)
}

// @Summary      List all sales
// @Description  Returns all non-deleted sales ordered by most recent first.
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
// @Description  Returns a single non-deleted sale by its ID.
// @Tags         sales
// @Produce      json
// @Param        id   path      int  true  "Sale ID"  example(1)
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
// @Description  Partially updates a sale. Only the fields provided are changed. For lines: existing item IDs update their quantity; new item IDs are added (name and unit_price required); items not mentioned are kept unchanged. The total is recomputed. The change is recorded in sale_history. Requires manager access.
// @Tags         sales
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Sale ID"         example(1)
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

	var changedByProf *int
	if profileID := c.GetInt("profileID"); profileID > 0 {
		changedByProf = &profileID
	}

	sale, err := h.service.UpdateSale(c.Request.Context(), id, input, changedByProf)
	if err != nil {
		if errors.Is(err, ErrSaleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNewLineIncomplete) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrNewLineZeroQuantity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrSaleWouldBeEmpty) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Description  Soft-deletes a sale by setting is_deleted=true. The deletion is recorded in sale_history. Requires manager access.
// @Tags         sales
// @Param        id   path      int  true  "Sale ID"  example(1)
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

	var changedByProf *int
	if profileID := c.GetInt("profileID"); profileID > 0 {
		changedByProf = &profileID
	}

	if err := h.service.DeleteSale(c.Request.Context(), id, changedByProf); err != nil {
		if errors.Is(err, ErrSaleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}
