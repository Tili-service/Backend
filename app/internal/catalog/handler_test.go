package catalog

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

func setupCatalogHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	return NewHandler(svc), mock
}

func TestCatalogHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestCatalogHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.POST("/store/:store_id/catalog", h.Create)

	storeID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/store/"+storeID+"/catalog", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.GET("/store/:store_id/catalog/:id", h.GetByID)

	storeID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/store/"+storeID+"/catalog/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCatalogHandler(t)
	r.GET("/store/:store_id/catalog/:id", h.GetByID)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\) AND \(c\.store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	storeID := uuid.New().String()
	catalogID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/store/"+storeID+"/catalog/"+catalogID, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCatalogHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.PUT("/store/:store_id/catalog/:id", h.Update)

	storeID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/store/"+storeID+"/catalog/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.DELETE("/store/:store_id/catalog/:id", h.Delete)

	storeID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/store/"+storeID+"/catalog/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCatalogHandler(t)
	r.GET("/store/:store_id/catalog", h.GetAll)

	storeID := uuid.New()
	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).AddRow("C1", "D1", storeID)
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/store/"+storeID.String()+"/catalog", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
