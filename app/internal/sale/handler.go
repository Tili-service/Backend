package sale

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"tili/app/internal/middleware"
	"tili/app/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const dateOnlyLayout = "2006-01-02"

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
		protected.GET("/kpi", h.GetSalesKPI)
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

	var changedByProf *uuid.UUID
	if profileID, err := uuid.Parse(c.GetString("profileID")); err == nil && profileID != uuid.Nil {
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
	id, err := uuid.Parse(c.Param("id"))
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

// @Summary      Sales KPI report
// @Description  Returns revenue KPIs bucketed by day, week, or month: total revenue, revenue and tax amount per tax bracket (5%, 10%, 20%, other), and revenue by product. Line amounts are treated as tax-inclusive. `from`/`to` are optional dates (YYYY-MM-DD); `to` is inclusive.
// @Tags         sales
// @Produce      json
// @Param        granularity  query     string  false  "daily, weekly, or monthly"  default(daily)
// @Param        from         query     string  false  "Start date (YYYY-MM-DD), inclusive"
// @Param        to           query     string  false  "End date (YYYY-MM-DD), inclusive"
// @Success      200  {object}  KPIReport
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sales/kpi [get]
func (h *Handler) GetSalesKPI(c *gin.Context) {
	granularity := Granularity(strings.ToLower(c.DefaultQuery("granularity", string(GranularityDaily))))

	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(dateOnlyLayout, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, expected YYYY-MM-DD"})
			return
		}
		from = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(dateOnlyLayout, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, expected YYYY-MM-DD"})
			return
		}
		t = t.AddDate(0, 0, 1) // to is inclusive of the given day
		to = &t
	}

	report, err := h.service.GetSalesKPI(c.Request.Context(), granularity, from, to)
	if err != nil {
		if errors.Is(err, ErrInvalidGranularity) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, report)
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	var input UpdateSaleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var changedByProf *uuid.UUID
	if profileID, err := uuid.Parse(c.GetString("profileID")); err == nil && profileID != uuid.Nil {
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	var changedByProf *uuid.UUID
	if profileID, err := uuid.Parse(c.GetString("profileID")); err == nil && profileID != uuid.Nil {
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
