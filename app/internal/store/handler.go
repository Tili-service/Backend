package store

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	storeRoutes := router.Group("/store")
	{
		storeRoutes.GET("/account/:accountID", h.GetByAccountID) // GET /store/account/:accountID
		storeRoutes.PUT("/:id", h.Update)                        // PUT /store/:id
	}
}

// Get store by account ID
// @Summary      Get store by account ID
// @Description  This route retrieves the store information associated with a specific account ID. It accepts an account ID as a path parameter and returns the corresponding store details upon success.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Param        accountID path      int true "Account ID"
// @Success      200  {object}  store.Store
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /store/account/{accountID} [get]
func (h *Handler) GetByAccountID(c *gin.Context) {
	accountID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	store, err := h.service.FindByAccountID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// Update an existing store
// @Summary      Update a store
// @Description  Updates the details of an existing store. The store ID is taken from the URL path and cannot be overridden. The request body must contain the new store name. Returns the updated `store.Store`.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true "Store ID"
// @Param        input body      store.UpdateStoreInput  true "Store update payload (storeName only)"
// @Success      200   {object}  store.Store
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /store/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	var input UpdateStoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, store)
}
