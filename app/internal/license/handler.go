package license

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"tili/app/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
)

type Handler struct {
	service               *Service
	mu                    sync.Mutex
	processedTransactions map[string]struct{}
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	accountRoutes := router.Group("/licences")
	accountRoutes.Use(middleware.AccountAuthMiddleware())
	{
		accountRoutes.GET("", h.GetLicences)
		accountRoutes.POST("payment", h.CreatePaymentLink)
	}
	router.POST("/api/webhooks/stripe", h.HandleStripeWebhook)
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

type PaymentLinkResponse struct {
	URL string `json:"url"`
}

// CreatePaymentLink creates a new payment link to buy a licence for the current account
// @Summary      Create a payment link
// @Description  Creates a new payment link for the authenticated account.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Security     AccountToken
// @Param        body body      CreatePaymentLinkInput true "Payment link creation payload"
// @Success      201  {object}  PaymentLinkResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /licences/payment [post]
func (h *Handler) CreatePaymentLink(c *gin.Context) {
	accountID := c.GetInt("accountID")
	customerID := c.GetString("customerID")

	var input CreatePaymentLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.service.CreatePaymentLink(c.Request.Context(), accountID, customerID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, PaymentLinkResponse{URL: url})
}

// HandleStripeWebhook processes incoming Stripe webhook events, specifically handling completed checkout sessions to create licences.
// @Summary      Handle Stripe webhook
// @Description  Endpoint to receive and process Stripe webhook events, creating licences upon successful checkout sessions.
// @Tags         licence
// @Accept       json
// @Produce      json
// @Param        body body      string true "Raw Stripe webhook payload"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      503  {object}  map[string]interface{}
// @Router       /api/webhooks/stripe [post]
func (h *Handler) HandleStripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Erreur de lecture"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature invalide"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing JSON"})
			return
		}

		h.mu.Lock()
		if h.processedTransactions == nil {
			h.processedTransactions = make(map[string]struct{})
		}
		if _, exists := h.processedTransactions[session.ID]; exists {
			h.mu.Unlock()
			c.Status(http.StatusOK)
			return
		}
		h.processedTransactions[session.ID] = struct{}{}
		h.mu.Unlock()

		accountIDStr, ok := session.Metadata["account_id"]
		if !ok || accountIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Metadata account_id manquante ou invalide"})
			return
		}
		accountID, _ := strconv.Atoi(accountIDStr)
		accountID, errConv := strconv.Atoi(accountIDStr)
		if errConv != nil || accountID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Metadata account_id invalide"})
			return
		}
		offer, ok := session.Metadata["offer"]
		if !ok || offer == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Metadata offer manquante ou invalide"})
			return
		}

		var durationDays int
		switch offer {
		case "mensuel":
			durationDays = 30
		case "semestriel":
			durationDays = 182
		case "annuel":
			durationDays = 365
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Offre invalide"})
			return
		}

		input := CreateLicenceInput{
			DurationDays: durationDays,
			Transaction:  session.ID,
		}

		_, err = h.service.Create(c.Request.Context(), accountID, input)
		if err != nil {
			fmt.Printf("Erreur création licence: %w\n", err)
		} else {
			fmt.Printf("✅ Licence créée avec succès pour le compte %d\n", accountID)
		}
	}
	c.Status(http.StatusOK)
}
