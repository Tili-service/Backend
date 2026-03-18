package item

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

func setupItemHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	return NewHandler(svc), mock
}

func TestItemHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestItemHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)
	r.POST("/item", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/item", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)
	r.PUT("/item/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/item/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)
	r.DELETE("/item/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/item/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)
	r.GET("/item/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/item/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_GetByCategorieID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupItemHandler(t)
	r.GET("/item/categorie/:id", h.GetByCategorieID)

	req := httptest.NewRequest(http.MethodGet, "/item/categorie/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestItemHandler_GetByName_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupItemHandler(t)
	r.GET("/item/name/:name", h.GetByName)

	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.name = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/item/name/missing", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupItemHandler(t)
	r.GET("/item", h.GetAll)

	rows := sqlmock.NewRows([]string{"item_id", "name", "price", "tax", "tax_amount", "categorie_id"}).
		AddRow(1, "Laptop", "100", "0.2", "16.67", 1)
	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i"$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/item", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
