package catalog

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
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
	r.POST("/catalog", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/catalog", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.GET("/catalog/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/catalog/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCatalogHandler(t)
	r.GET("/catalog/:id", h.GetByID)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/catalog/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCatalogHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.PUT("/catalog/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/catalog/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCatalogHandler(t)
	r.DELETE("/catalog/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/catalog/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCatalogHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCatalogHandler(t)
	r.GET("/catalog", h.GetAll)

	rows := sqlmock.NewRows([]string{"name", "description"}).AddRow("C1", "D1")
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c"$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
