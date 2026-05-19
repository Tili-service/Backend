package salehistory

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

func (h *Handler) RegisterRoutes(rg *gin.Engine) {
	sales := rg.Group("/sales")
	protected := sales.Group("")
	protected.Use(middleware.ProfileAuthMiddleware())
	{
		managerRoutes := protected.Group("")
		managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
		{
			managerRoutes.GET("/:id/history", h.ListBySaleID)
		}
	}
}

// @Summary      List the change history of a sale
// @Description  Returns the audit trail of changes for a sale, ordered most-recent-first. Each row captures the state of the sale after the change was applied, alongside a diff of what changed (fields "action", and any modified field with from/to values). Requires manager access.
// @Tags         sales
// @Produce      json
// @Param        id   path      int  true  "Sale ID"  example(1)
// @Success      200  {array}   SaleHistory
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /sales/{id}/history [get]
func (h *Handler) ListBySaleID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sale id"})
		return
	}

	history, err := h.service.ListBySaleID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, history)
}
