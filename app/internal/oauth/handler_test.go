package oauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"golang.org/x/oauth2"

	"tili/app/internal/store"
	"tili/app/pkg/db"
)

func setupTestRouter(storeSvc *store.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewHandler(storeSvc)
	handler.RegisterRoutes(r)
	return r
}

func setupMockDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	assert.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	return bunDB, mock
}

func TestRefreshToken_MissingStoreID(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()
	storeRepo := store.NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(storeRepo)

	router := setupTestRouter(storeSvc)

	req, _ := http.NewRequest(http.MethodPost, "/oauth/refresh", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestRefreshToken_StoreNotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	storeID := uuid.New()
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sqlmock.ErrCancelled)

	storeRepo := store.NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(storeRepo)

	router := setupTestRouter(storeSvc)

	req, _ := http.NewRequest(http.MethodPost, "/oauth/refresh?store_id="+storeID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestRefreshToken_NoRefreshToken(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	storeID := uuid.New()
	buyerID := uuid.New()

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "sumup_merchant_code", "sumup_access_token", "sumup_refresh_token", "siret"}).
		AddRow(storeID, "Store Without Token", buyerID, uuid.New(), time.Now(), "", "", "", "", "")

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	storeRepo := store.NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(storeRepo)

	router := setupTestRouter(storeSvc)

	req, _ := http.NewRequest(http.MethodPost, "/oauth/refresh?store_id="+storeID.String(), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var res map[string]string
	_ = json.Unmarshal(resp.Body.Bytes(), &res)
	assert.Equal(t, "no refresh token available for this store", res["error"])
}

func TestRefreshToken_Success(t *testing.T) {
	// Mock SumUp Token server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		_ = r.ParseForm()
		assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
		assert.Equal(t, "my_refresh_token", r.Form.Get("refresh_token"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new_access_token_123",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "new_refresh_token_456",
		})
	}))
	defer server.Close()

	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	storeID := uuid.New()
	buyerID := uuid.New()

	rows1 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "sumup_merchant_code", "sumup_access_token", "sumup_refresh_token", "siret"}).
		AddRow(storeID, "Store With Token", buyerID, uuid.New(), time.Now(), "", "M123", "old_access_token", "my_refresh_token", "")
	rows2 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "sumup_merchant_code", "sumup_access_token", "sumup_refresh_token", "siret"}).
		AddRow(storeID, "Store With Token", buyerID, uuid.New(), time.Now(), "", "M123", "old_access_token", "my_refresh_token", "")

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows1)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows2)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	storeRepo := store.NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(storeRepo)

	router := setupTestRouter(storeSvc)

	// Override sumupOauthConfig endpoint URL to point to mock server
	sumupOauthConfig = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL,
		},
	}

	body, _ := json.Marshal(map[string]string{"store_id": storeID.String()})
	req, _ := http.NewRequest(http.MethodPost, "/oauth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if !assert.Equal(t, http.StatusOK, resp.Code) {
		t.Logf("Response body: %s", resp.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(resp.Body.Bytes(), &res)
	assert.Equal(t, "Token refreshed successfully", res["message"])
	assert.Equal(t, "new_access_token_123", res["access_token"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
