package license

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func setupLicenseHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
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
