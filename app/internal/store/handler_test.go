package store

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/internal/profile"
	"tili/app/pkg/db"
)

func setupStoreHandler(t *testing.T) *Handler {
	bunDB, _ := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	storeRepo := NewRepository(&db.Db{DB: bunDB})
	storeSvc := NewService(storeRepo)

	profileRepo := profile.NewRepository(&db.Db{DB: bunDB})
	profileSvc := profile.NewService(profileRepo)

	return NewHandler(storeSvc, profileSvc)
}

func withAccountContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("accountID", 1)
		c.Set("name", "owner")
		c.Next()
	}
}

func TestStoreHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestStoreHandler_CreateStore_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.POST("/store", h.CreateStore)

	req := httptest.NewRequest(http.MethodPost, "/store", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreHandler_DeleteStore_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	req := httptest.NewRequest(http.MethodDelete, "/store/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.GET("/store/:id", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/store/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	req := httptest.NewRequest(http.MethodPut, "/store/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStoreHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.GET("/store", h.GetAll)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name"}).AddRow(1, "S1")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s"$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/store", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_GetAll_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.GET("/store", h.GetAll)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s"$`).WillReturnError(errors.New("db down"))

	req := httptest.NewRequest(http.MethodGet, "/store", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_GetMyStores_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.GET("/store/me", h.GetMyStores)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(1, "Mine", 1)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/store/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_GetMyStores_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.GET("/store/me", h.GetMyStores)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnError(errors.New("db down"))

	req := httptest.NewRequest(http.MethodGet, "/store/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 2, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_InternalErrorOnFind(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("db fail"))

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_ProfileDeleteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))
	h.profileService = profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("profile delete failed"))

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_ServiceDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))
	h.profileService = profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_DeleteStore_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.DELETE("/store/:id", h.DeleteStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))
	h.profileService = profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))

	rows1 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	rows2 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows1)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows2)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.GET("/store/:id", h.GetByID)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_GetByID_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.GET("/store/:id", h.GetByID)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 2, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/store/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 2, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_InternalFindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("db fail"))

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_UpdateError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("select fail"))

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_UpdateStoreById_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.PUT("/store/:id", h.updateStoreById)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))

	rows1 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	rows2 := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "S1", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows1)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows2)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPut, "/store/1", bytes.NewBufferString(`{"name":"new"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStoreHandler_CreateStore_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := setupStoreHandler(t)
	r.Use(withAccountContext())
	r.POST("/store", h.CreateStore)

	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })
	h.service = NewService(NewRepository(&db.Db{DB: bunDB}))
	h.profileService = profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))

	licenceID := uuid.New()
	mock.ExpectQuery(`^INSERT INTO "store"`).WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow(1))
	mock.ExpectQuery(`^SELECT EXISTS \(SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\) AND \(pin = .+\)\)$`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`^INSERT INTO "profile"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(1))

	req := httptest.NewRequest(http.MethodPost, "/store", bytes.NewBufferString(`{"name":"New Store","licence_id":"`+licenceID.String()+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
