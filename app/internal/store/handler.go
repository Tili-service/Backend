package store

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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
		storeRoutes.GET("/account/:accountID", h.GetByAccountID) // GET /store/account/:accountID get store by account ID
		storeRoutes.PUT("/:id", h.Update)
	}
}

// Get store by account ID
// @Summary      Get store by account ID
// @Description  This route retrieves the store information associated with a specific account ID. It accepts an account ID as a path parameter and returns the corresponding store details upon success.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Param        accountID path      int64 true "Account ID"
// @Success      200  {object}  store.Store
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /store/account/{accountID} [get]
func (h *Handler) GetByAccountID(c *gin.Context) {
	accountIDParam := c.Param("accountID")
	accountID, err := strconv.ParseInt(accountIDParam, 10, 64)
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
// @Description  Updates the details of an existing store. The request body must contain an `UpdateStoreInput` object with the store ID and new values. Returns the updated `store.Store`.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Param        input body      store.UpdateStoreInput true "Store update payload"
// @Success      200  {object}  store.Store
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var input UpdateStoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	store, err := h.service.Update(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, store)
}
