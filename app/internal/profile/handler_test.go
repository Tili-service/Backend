package profile

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

func setupProfileHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	bunDB, mock := setupMockDB(t)
	t.Cleanup(func() { _ = bunDB.Close() })

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	return NewHandler(svc), mock
}

func withProfileContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("profileID", 1)
		c.Set("storeID", 10)
		c.Next()
	}
}

func TestProfileHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)

	h.RegisterRoutes(r)

	assert.NotEmpty(t, r.Routes())
}

func TestProfileHandler_LoginWithPin_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.POST("/profile/login/pin", h.loginWithPin)

	req := httptest.NewRequest(http.MethodPost, "/profile/login/pin", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_Me_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.Use(withProfileContext())
	r.GET("/profile/me", h.me)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnError(sql.ErrConnDone)

	req := httptest.NewRequest(http.MethodGet, "/profile/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Create_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.Use(withProfileContext())
	r.POST("/profile", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.DELETE("/profile/:id", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/profile/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.PUT("/profile/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/profile/abc", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_GetProfilesByStoreId_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.GET("/profile/allProfilesByStoreId/:id", h.GetProfilesByStoreId)

	req := httptest.NewRequest(http.MethodGet, "/profile/allProfilesByStoreId/abc", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_UpdateProfileByIdAndStoreId_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.PUT("/profile/updateProfile/:id/:storeId", h.UpdateProfileByIdAndStoreId)

	req := httptest.NewRequest(http.MethodPut, "/profile/updateProfile/abc/1", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	req2 := httptest.NewRequest(http.MethodPut, "/profile/updateProfile/1/abc", bytes.NewBufferString("{}"))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestProfileHandler_DeactivateProfile_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.PUT("/profile/deactivateProfile/:id/:storeId", h.DeactivateProfile)

	req := httptest.NewRequest(http.MethodPut, "/profile/deactivateProfile/abc/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	req2 := httptest.NewRequest(http.MethodPut, "/profile/deactivateProfile/1/abc", nil)
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestProfileHandler_LoginWithPin_InvalidPin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.POST("/profile/login/pin", h.loginWithPin)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.store_id = .+\) AND \(p\.pin = .+\) AND \(p\.is_active = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/profile/login/pin", bytes.NewBufferString(`{"store_id":10,"pin":"000000"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_LoginWithPin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.POST("/profile/login/pin", h.loginWithPin)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Alice", "123456", 3, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.store_id = .+\) AND \(p\.pin = .+\) AND \(p\.is_active = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodPost, "/profile/login/pin", bytes.NewBufferString(`{"store_id":10,"pin":"123456"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.Use(withProfileContext())
	r.POST("/profile", h.Create)

	mock.ExpectQuery(`^SELECT EXISTS \(SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\) AND \(pin = .+\)\)$`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`^INSERT INTO "profile"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(1))

	req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(`{"name":"Bob","level_access":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Delete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.DELETE("/profile/:id", h.Delete)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodDelete, "/profile/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Update_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.PUT("/profile/:id", h.Update)

	req := httptest.NewRequest(http.MethodPut, "/profile/1", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_GetProfilesByStoreId_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.GET("/profile/allProfilesByStoreId/:id", h.GetProfilesByStoreId)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/profile/allProfilesByStoreId/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_UpdateProfileByIdAndStoreId_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, _ := setupProfileHandler(t)
	r.PUT("/profile/updateProfile/:id/:storeId", h.UpdateProfileByIdAndStoreId)

	req := httptest.NewRequest(http.MethodPut, "/profile/updateProfile/1/10", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProfileHandler_DeactivateProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.PUT("/profile/deactivateProfile/:id/:storeId", h.DeactivateProfile)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPut, "/profile/deactivateProfile/1/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.DELETE("/profile/:id", h.Delete)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(profile_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodDelete, "/profile/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.PUT("/profile/:id", h.Update)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPut, "/profile/1", bytes.NewBufferString(`{"name":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileHandler_UpdateProfileByIdAndStoreId_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h, mock := setupProfileHandler(t)
	r.PUT("/profile/updateProfile/:id/:storeId", h.UpdateProfileByIdAndStoreId)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPut, "/profile/updateProfile/1/10", bytes.NewBufferString(`{"name":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
