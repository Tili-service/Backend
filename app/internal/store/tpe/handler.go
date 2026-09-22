package tpe

import (
	"strconv"
	"tili/app/internal/middleware"
	"tili/app/internal/store"

	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service      *Service
	storeService *store.Service
}

func NewHandler(service *Service, storeService *store.Service) *Handler {
	return &Handler{service: service, storeService: storeService}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	tpeRoutes := router.Group("/store/tpe")
	accountProtected := tpeRoutes.Group("")
	accountProtected.Use(middleware.ProfileAuthMiddleware())
	{
		accountProtected.GET("/", h.getTPE)                         // GET /store/tpe/
		accountProtected.POST("/payment", h.initiatePayment)        // POST /store/tpe/payment
		accountProtected.GET("/payment/status", h.getPaymentStatus) // GET /store/tpe/payment/status
	}
}

// GetMyTPE retrieves all TPE connected to the store
// @Summary      Get store TPE
// @Description  Returns all TPE connected to the store using sumup API.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Success      200  {array}   TPE
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/tpe [get]
func (h *Handler) getTPE(c *gin.Context) {
	storeID, err := uuid.Parse(c.GetString("storeID"))
	if err != nil || storeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid storeID in token"})
		return
	}

	st, err := h.storeService.FindByID(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tpeList, err := h.service.getTPE(c.Request.Context(), st)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tpeList)
}

// SendPayment emit a payment request to the TPE
// @Summary      Send payment request to TPE
// @Description  Sends a payment request to the TPE connected to the store using sumup API.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param        description   body      string  true  "Payment description"
// @Param	     amount    query     float64  true  "Payment amount"
// @Param	     currency  query     string   true  "Payment currency"
// @Param	     tpe_id    query     string   true  "TPE ID"
// @Success      200  {array}   TPE
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/tpe/payment [POST]
func (h *Handler) initiatePayment(c *gin.Context) {
	storeID, err := uuid.Parse(c.GetString("storeID"))
	if err != nil || storeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid storeID in token"})
		return
	}

	st, err := h.storeService.FindByID(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tpeID := c.Query("tpe_id")
	if tpeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tpe_id query parameter"})
		return
	}

	tpeList, err := h.service.getTPE(c.Request.Context(), st)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validTPE := false
	for _, tpe := range tpeList {
		if tpe.ID != nil && *tpe.ID == tpeID {
			validTPE = true
			break
		}
	}

	if !validTPE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tpe_id for this store"})
		return
	}

	amountFloat, err := strconv.ParseFloat(c.Query("amount"), 64)
	if err != nil || amountFloat <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount query parameter"})
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		currency = "EUR"
	}
	amountInCents := int64(amountFloat * 100)

	description := c.Query("description")
	if description == "" {
		description = "Payment"
	}

	paymentInformation := TPEPaymentRequest{
		TotalAmount: TotalAmount{
			Value:     amountInCents,
			Currency:  currency,
			MinorUnit: 2,
		},
		Description: description,
		TPEID:       tpeID,
	}

	checkoutData, err := h.service.initiatePayment(c.Request.Context(), st, paymentInformation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	checkoutID := ""
	status := "PENDING"
	if checkoutData != nil {
		checkoutID = checkoutData.ID
		if checkoutData.Status != "" {
			status = checkoutData.Status
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "payment request sent to TPE successfully",
		"checkout_id": checkoutID,
		"status":      status,
		"tpe_id":      tpeID,
		"amount":      amountFloat,
	})
}

// GetPaymentStatus retrieves the status of a payment checkout on the TPE
// @Summary      Get TPE payment status
// @Description  Retrieves the status of a payment checkout on the TPE using SumUp API.
// @Tags         stores
// @Accept       json
// @Produce      json
// @Security     ProfileToken
// @Param	     tpe_id       query     string  true  "TPE ID"
// @Param	     checkout_id  query     string  true  "Checkout ID"
// @Success      200  {object}  ReaderCheckoutData
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /store/tpe/payment/status [get]
func (h *Handler) getPaymentStatus(c *gin.Context) {
	storeID, err := uuid.Parse(c.GetString("storeID"))
	if err != nil || storeID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid storeID in token"})
		return
	}

	st, err := h.storeService.FindByID(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tpeID := c.Query("tpe_id")
	checkoutID := c.Query("checkout_id")
	if tpeID == "" || checkoutID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tpe_id or checkout_id query parameter"})
		return
	}

	statusData, err := h.service.getPaymentStatus(c.Request.Context(), st, tpeID, checkoutID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statusData)
}
