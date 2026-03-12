package license

import (
	"net/http"

	"tili/app/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	accountRoutes := router.Group("/licences")
	accountRoutes.Use(middleware.AccountAuthMiddleware())
	{
		accountRoutes.GET("", h.GetLicences)
		accountRoutes.POST("", h.CreateLicence)
	}
}

// GetLicences retrieves all licences for the current account
// @Summary      Get my licences
// @Description  Returns all licences belonging to the currently authenticated account.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Success      200  {array}   Licence
// @Failure      500  {object}  map[string]interface{}
// @Router       /licences [get]
func (h *Handler) GetLicences(c *gin.Context) {
	accountID := c.GetInt("accountID")
	licences, err := h.service.GetByAccountID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, licences)
}

// CreateLicence creates a new licence for the current account
// @Summary      Buy a licence
// @Description  Creates a new licence for the authenticated account. The licence_id can then be used when creating a store.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        body body      CreateLicenceInput true "Licence creation payload"
// @Success      201  {object}  Licence
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /licences [post]
func (h *Handler) CreateLicence(c *gin.Context) {
	accountID := c.GetInt("accountID")

	var input CreateLicenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lic, err := h.service.Create(c.Request.Context(), accountID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lic)
}
