package license

import (
	"net/http"

	"tili/app/internal/middleware"

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
	accountRoutes := router.Group("/licences")
	accountRoutes.Use(middleware.AccountAuthMiddleware())
	{
		accountRoutes.GET("", h.GetLicences)
		accountRoutes.POST("", h.CreateLicence)
		accountRoutes.GET("/:id", h.GetByID)   // GET /licences/:id
		accountRoutes.DELETE("/:id", h.Delete) // DELETE /licences/:id
		accountRoutes.PUT("/:id", h.Update)    // PUT /licences/:id
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

// GetByID retrieves a licence by its ID
// @Summary      Get a licence by ID
// @Description  Returns the details of a specific licence. Only the owner can access it.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        id  path      string  true  "Licence UUID"
// @Success      200  {object}  Licence
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /licences/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	accountID := c.GetInt("accountID")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid licence ID"})
		return
	}

	lic, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "licence not found"})
		return
	}
	if lic.AccountID != accountID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the owner of this licence"})
		return
	}
	c.JSON(http.StatusOK, lic)
}

// Delete removes a licence
// @Summary      Delete a licence
// @Description  Deletes a licence by its ID. Only the owner can delete it.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        id  path      string  true  "Licence UUID"
// @Success      204
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /licences/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	accountID := c.GetInt("accountID")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid licence ID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), accountID, id); err != nil {
		switch err.Error() {
		case "licence not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "forbidden":
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not the owner of this licence"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// Update modifies an existing licence
// @Summary      Update a licence
// @Description  Updates the transaction reference or active status of a licence. Only the owner can update it.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        id   path      string             true  "Licence UUID"
// @Param        body body      UpdateLicenceInput true  "Licence update payload"
// @Success      200  {object}  Licence
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /licences/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	accountID := c.GetInt("accountID")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid licence ID"})
		return
	}

	var input UpdateLicenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lic, err := h.service.Update(c.Request.Context(), accountID, id, input)
	if err != nil {
		switch err.Error() {
		case "licence not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "forbidden":
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not the owner of this licence"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, lic)
}
