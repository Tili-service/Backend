package categorie

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

func setupCategorieHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	return NewHandler(svc), mock
}

func TestHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCategorieHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCategorieHandler(t)
	r.POST("/catalog/:catalog_id/categorie", h.Create)

	catalogID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/catalog/"+catalogID+"/categorie", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCategorieHandler(t)
	r.GET("/catalog/:catalog_id/categorie/:id", h.GetByID)

	catalogID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/catalog/"+catalogID+"/categorie/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCategorieHandler(t)
	r.PUT("/catalog/:catalog_id/categorie/:id", h.Update)

	catalogID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPut, "/catalog/"+catalogID+"/categorie/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_DeleteByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupCategorieHandler(t)
	r.DELETE("/catalog/:catalog_id/categorie/:id", h.DeleteByID)

	catalogID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/catalog/"+catalogID+"/categorie/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetByType_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCategorieHandler(t)
	r.GET("/catalog/:catalog_id/categorie/type/:type", h.GetByType)

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/catalog/"+catalogID.String()+"/categorie/type/missing", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_DeleteByType_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCategorieHandler(t)
	r.DELETE("/catalog/:catalog_id/categorie/type/:type", h.DeleteByType)

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodDelete, "/catalog/"+catalogID.String()+"/categorie/type/missing", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupCategorieHandler(t)
	r.GET("/catalog/:catalog_id/categorie", h.GetAll)

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(catID, "Books", catalogID)
	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.catalog_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/catalog/"+catalogID.String()+"/categorie", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
