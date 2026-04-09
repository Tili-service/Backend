package payementmethod

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

func setupPayementMethodHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	return NewHandler(svc), mock
}

func TestPayementMethodHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupPayementMethodHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestPayementMethodHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupPayementMethodHandler(t)
	r.POST("/payementmethod", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/payementmethod", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPayementMethodHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupPayementMethodHandler(t)
	r.PUT("/payementmethod/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/payementmethod/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPayementMethodHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupPayementMethodHandler(t)
	r.DELETE("/payementmethod/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/payementmethod/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPayementMethodHandler_GetByName_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupPayementMethodHandler(t)
	r.GET("/payementmethod/name/:name", h.GetByName)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/payementmethod/name/missing", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPayementMethodHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupPayementMethodHandler(t)
	r.GET("/payementmethod", h.GetAll)

	rows := sqlmock.NewRows([]string{"payement_method_id", "name"}).AddRow(1, "Card")
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm"$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/payementmethod", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
