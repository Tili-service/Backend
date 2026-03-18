package profile

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_NewService(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	assert.NotNil(t, svc)
}

func TestGeneratePIN_Length(t *testing.T) {
	pin, err := generatePIN(6)

	assert.NoError(t, err)
	assert.Len(t, pin, 6)
}

func TestService_Create_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	// First PIN check returns false, so generated PIN is accepted.
	rowsExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`^SELECT EXISTS \(SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\) AND \(pin = .+\)\)$`).WillReturnRows(rowsExists)
	mock.ExpectQuery(`^INSERT INTO "profile"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(1))

	out, err := svc.Create(context.Background(), CreateProfileInput{StoreID: 10, Name: "Alice"})

	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.Equal(t, 10, out.StoreID)
		assert.Equal(t, "Alice", out.Name)
		assert.Equal(t, 4, out.LevelAccess)
		assert.True(t, out.IsActive)
		assert.Len(t, out.Pin, 6)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Create_PinCheckError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT EXISTS \(SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\) AND \(pin = .+\)\)$`).WillReturnError(assert.AnError)

	_, err := svc.Create(context.Background(), CreateProfileInput{StoreID: 10, Name: "Alice"})

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Alice", "123456", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)

	p, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	if assert.NotNil(t, p) {
		assert.Equal(t, "Alice", p.Name)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), 1)

	assert.ErrorIs(t, err, ErrProfileNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByStoreID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.DeleteByStoreID(context.Background(), 10)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LoginWithPin_Invalid(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.store_id = .+\) AND \(p\.pin = .+\) AND \(p\.is_active = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.LoginWithPin(context.Background(), 10, "999999")

	assert.EqualError(t, err, "invalid pin")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), 1, updateProfileInput{})

	assert.ErrorIs(t, err, ErrProfileNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetProfilesByStoreId_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).
		AddRow(1, 10, "A", "111111", 4, true).
		AddRow(2, 10, "B", "222222", 2, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	list, err := svc.GetProfilesByStoreId(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateProfileByIdAndStoreId_StoreMismatch(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 999, "A", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)

	_, err := svc.UpdateProfileByIdAndStoreId(context.Background(), 1, 10, updateProfileInput{})

	assert.EqualError(t, err, "profile does not belong to the specified store")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeactivateProfile_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.DeactivateProfile(context.Background(), 1, 10)

	assert.EqualError(t, err, "profile not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	name := "New"
	level := 2
	active := false
	out, err := svc.Update(context.Background(), 1, updateProfileInput{Name: &name, LevelAccess: &level, IsActive: &active})

	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.Equal(t, "New", out.Name)
		assert.Equal(t, 2, out.LevelAccess)
		assert.False(t, out.IsActive)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_UpdateError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnError(sql.ErrConnDone)

	_, err := svc.Update(context.Background(), 1, updateProfileInput{})

	assert.EqualError(t, err, "failed to update profile")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateProfileByIdAndStoreId_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	name := "New"
	out, err := svc.UpdateProfileByIdAndStoreId(context.Background(), 1, 10, updateProfileInput{Name: &name})

	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.Equal(t, "New", out.Name)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_UpdateProfileByIdAndStoreId_UpdateError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnError(sql.ErrConnDone)

	_, err := svc.UpdateProfileByIdAndStoreId(context.Background(), 1, 10, updateProfileInput{})

	assert.EqualError(t, err, "failed to update profile")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeactivateProfile_StoreMismatch(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 99, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)

	_, err := svc.DeactivateProfile(context.Background(), 1, 10)

	assert.EqualError(t, err, "profile does not belong to the specified store")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeactivateProfile_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := svc.DeactivateProfile(context.Background(), 1, 10)

	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.False(t, out.IsActive)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeactivateProfile_UpdateError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "pin", "level_access", "is_active"}).AddRow(1, 10, "Old", "111111", 4, true)
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.profile_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnError(sql.ErrConnDone)

	_, err := svc.DeactivateProfile(context.Background(), 1, 10)

	assert.EqualError(t, err, "failed to deactivate profile")
	assert.NoError(t, mock.ExpectationsWereMet())
}
