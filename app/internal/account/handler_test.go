package account

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"golang.org/x/crypto/bcrypt"

	"tili/app/internal/profile"
	"tili/app/internal/store"
	"tili/app/pkg/db"
)

func setupTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	bunDB := bun.NewDB(sqldb, pgdialect.New())
	return bunDB, mock
}

type MockLicenseDeleter struct{}

func (m *MockLicenseDeleter) DeleteByAccountID(ctx context.Context, accountID int) error {
	return nil
}

func setupTestEnv(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, *Service) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	bunDB, mock := setupTestDB(t)

	mockDbObj := &db.Db{DB: bunDB}
	repo := NewRepository(mockDbObj)

	// Since store and profile services expect a similar db setup, we use the easiest dummy references.
	// Depending on your requirements, you might want full interfaces. For now we use actual structs with mock DB.
	storeRepo := store.NewRepository(mockDbObj)
	storeService := store.NewService(storeRepo)
	profileRepo := profile.NewRepository(mockDbObj)
	profileService := profile.NewService(profileRepo)

	licenceService := &MockLicenseDeleter{}

	service := NewService(repo, storeService, profileService, licenceService)
	handler := NewHandler(service)

	// Middleware stub for GetAccount
	protected := router.Group("/account")
	protected.Use(func(c *gin.Context) {
		c.Set("accountID", 1) // Force mock account ID
		c.Next()
	})
	protected.GET("", handler.GetAccount)
	protected.PUT("", handler.Update)

	return router, mock, service
}

func TestHandler_GetAccount(t *testing.T) {
	router, mock, _ := setupTestEnv(t)

	// Expected row return for Account 1 from GetAccount
	rows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(1, "test@example.com", "Test User", "pwd", "cus_123", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a"."account_id", "a"."email", "a"."password", "a"."name", "a"."stripe_customer_id", "a"."created_at" FROM "account" AS "a" WHERE (account_id = 1)`)).
		WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/account", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test@example.com")
	assert.Contains(t, w.Body.String(), "Test User")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Update(t *testing.T) {
	router, mock, _ := setupTestEnv(t)

	rows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(1, "old@example.com", "Old Name", "pwd", "cus_123", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a"."account_id", "a"."email", "a"."password", "a"."name", "a"."stripe_customer_id", "a"."created_at" FROM "account" AS "a" WHERE (account_id = 1)`)).WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "account" AS "a"`)).WillReturnResult(sqlmock.NewResult(1, 1))

	updateData := gin.H{"name": "New Name", "email": "new@example.com"}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/account", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "New Name")
	assert.Contains(t, w.Body.String(), "new@example.com")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Create_BadJSON(t *testing.T) {
	router, _, service := setupTestEnv(t)
	handler := NewHandler(service)
	router.POST("/account/create-test", handler.Create)

	req, _ := http.NewRequest("POST", "/account/create-test", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Login_BadJSON(t *testing.T) {
	router, _, service := setupTestEnv(t)
	handler := NewHandler(service)
	router.POST("/account/login-test", handler.Login)

	req, _ := http.NewRequest("POST", "/account/login-test", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	router, mock, service := setupTestEnv(t)
	handler := NewHandler(service)

	protected := router.Group("/account")
	protected.Use(func(c *gin.Context) {
		c.Set("accountID", 1)
		c.Next()
	})
	protected.DELETE("", handler.Delete)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a"."account_id", "a"."email", "a"."password", "a"."name", "a"."stripe_customer_id", "a"."created_at" FROM "account" AS "a" WHERE (account_id = 1)`)).WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("DELETE", "/account", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Create_ConflictEmailExists(t *testing.T) {
	router, mock, service := setupTestEnv(t)
	handler := NewHandler(service)
	router.POST("/account/create-conflict", handler.Create)

	rows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(1, "already@example.com", "Already", "pwd", "cus_123", time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest("POST", "/account/create-conflict", bytes.NewBufferString(`{"name":"A","email":"already@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	router, mock, service := setupTestEnv(t)
	handler := NewHandler(service)
	router.POST("/account/login-invalid", handler.Login)

	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("POST", "/account/login-invalid", bytes.NewBufferString(`{"email":"x@y.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Login_Success(t *testing.T) {
	router, mock, service := setupTestEnv(t)
	handler := NewHandler(service)
	router.POST("/account/login-success", handler.Login)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	accRows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(1, "test@example.com", "User", string(hash), "cus_123", time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnRows(accRows)

	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "Store", 1, "00000000-0000-0000-0000-000000000000", time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(storeRows)

	req := httptest.NewRequest("POST", "/account/login-success", bytes.NewBufferString(`{"email":"test@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_Update_NotFound(t *testing.T) {
	router, mock, _ := setupTestEnv(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a"."account_id", "a"."email", "a"."password", "a"."name", "a"."stripe_customer_id", "a"."created_at" FROM "account" AS "a" WHERE (account_id = 1)`)).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("PUT", "/account", bytes.NewBufferString(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
