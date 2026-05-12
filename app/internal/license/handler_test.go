package license

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"

	"tili/app/internal/profile"
	"tili/app/internal/store"
	"tili/app/pkg/db"
)

func setupLicenseHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	svc.SetDependencies(store.NewService(store.NewRepository(&db.Db{DB: bunDB})), profile.NewService(profile.NewRepository(&db.Db{DB: bunDB})))
	return NewHandler(svc), mock
}

func withLicenseAccountContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("accountID", 1)
		c.Set("customerID", "")
		c.Next()
	}
}

func TestLicenseHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestLicenseHandler_CreatePaymentLink_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.POST("/licences/payment", h.CreatePaymentLink)

	req := httptest.NewRequest(http.MethodPost, "/licences/payment", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_HandleStripeWebhook_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.POST("/api/webhooks/stripe", h.HandleStripeWebhook)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewBufferString("{}"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_HandleStripeWebhook_CheckoutCompleted_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.POST("/api/webhooks/stripe", h.HandleStripeWebhook)

	secret := "whsec_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)

	payload := `{"id":"evt_test","object":"event","api_version":"2026-02-25.clover","type":"checkout.session.completed","data":{"object":{"id":"cs_test_create_1","object":"checkout.session","metadata":{"account_id":"1","offer":"mensuel"}}}}`
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	mock.ExpectExec(`^INSERT INTO "licence"`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewBuffer(signed.Payload))
	req.Header.Set("Stripe-Signature", signed.Header)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_HandleStripeWebhook_SubscriptionDeleted_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.POST("/api/webhooks/stripe", h.HandleStripeWebhook)

	secret := "whsec_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, 1, "sub_123")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(transaction = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	payload := `{"id":"evt_test","object":"event","api_version":"2026-02-25.clover","type":"customer.subscription.deleted","data":{"object":{"id":"sub_123","object":"subscription"}}}`
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewBuffer(signed.Payload))
	req.Header.Set("Stripe-Signature", signed.Header)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_HandleStripeWebhook_SubscriptionDeleted_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.POST("/api/webhooks/stripe", h.HandleStripeWebhook)

	secret := "whsec_test"
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)

	payload := `{"id":"evt_test","object":"event","api_version":"2026-02-25.clover","type":"customer.subscription.deleted","data":{"object":{"object":"subscription"}}}`
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   []byte(payload),
		Secret:    secret,
		Timestamp: time.Now(),
		Scheme:    "v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewBuffer(signed.Payload))
	req.Header.Set("Stripe-Signature", signed.Header)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_GetByID_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.GET("/licences/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/licences/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_Delete_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.DELETE("/licences/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/licences/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_Update_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.PUT("/licences/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/licences/not-a-uuid", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_GetByID_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.GET("/licences/:id", h.GetByID)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 2)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/licences/"+licID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_GetLicences_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.GET("/licences", h.GetLicences)

	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l"`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/licences", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.GET("/licences/:id", h.GetByID)

	licID := uuid.New()
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/licences/"+licID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_GetByID_WithStripeInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.GET("/licences/:id", h.GetByID)

	origRetrieveCheckoutSession := retrieveCheckoutSession
	origRetrieveSubscriptionForLicence := retrieveSubscriptionForLicence
	t.Cleanup(func() {
		retrieveCheckoutSession = origRetrieveCheckoutSession
		retrieveSubscriptionForLicence = origRetrieveSubscriptionForLicence
	})

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		return &stripe.CheckoutSession{
			ID:           "cs_test_123",
			Subscription: &stripe.Subscription{ID: "sub_123"},
		}, nil
	}
	retrieveSubscriptionForLicence = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		return &stripe.Subscription{
			ID:                "sub_123",
			Status:            "active",
			CancelAtPeriodEnd: true,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{{
					CurrentPeriodEnd: 1714406400,
					Price: &stripe.Price{
						ID:         "price_123",
						UnitAmount: 1999,
						Currency:   stripe.Currency("eur"),
						Recurring: &stripe.PriceRecurring{
							Interval: stripe.PriceRecurringIntervalMonth,
						},
					},
				}},
			},
		}, nil
	}

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, 1, "cs_test_123")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/licences/"+licID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"stripe"`)
	assert.Contains(t, w.Body.String(), `"subscription_id":"sub_123"`)
	assert.Contains(t, w.Body.String(), `"next_payment_at"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.DELETE("/licences/:id", h.Delete)

	licID := uuid.New()
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodDelete, "/licences/"+licID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_Delete_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.DELETE("/licences/:id", h.Delete)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 999)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodDelete, "/licences/"+licID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_Update_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.PUT("/licences/:id", h.Update)

	licID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/licences/"+licID.String(), bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_Update_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.PUT("/licences/:id", h.Update)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 999)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodPut, "/licences/"+licID.String(), bytes.NewBufferString(`{"transaction":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLicenseHandler_RefundLicense_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.POST("/licences/refund-license", h.RefundLicense)

	req := httptest.NewRequest(http.MethodPost, "/licences/refund-license?licenceId=bad-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLicenseHandler_RefundLicense_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupLicenseHandler(t)
	r.Use(withLicenseAccountContext())
	r.POST("/licences/refund-license", h.RefundLicense)

	licID := uuid.New()
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/licences/refund-license?licenceId="+licID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
