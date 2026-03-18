package license

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"tili/app/pkg/db"
)

func setupMockDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	assert.NoError(t, err)

	bunDB := bun.NewDB(sqldb, pgdialect.New())
	return bunDB, mock
}

func TestRepository_CreateLicence(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mockUUID := uuid.New()
	mock.ExpectQuery(`^INSERT INTO "licence"`).WillReturnRows(sqlmock.NewRows([]string{"licence_id"}).AddRow(mockUUID))

	l := &Licence{AccountID: 1, Expiration: time.Now(), IsActive: true}
	ctx := context.Background()
	err := repo.CreateLicence(ctx, l)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindLicencesByAccountID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mockUUID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(mockUUID, 1)

	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l"`).WillReturnRows(rows)

	ctx := context.Background()
	licences, err := repo.FindLicencesByAccountID(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, licences, 1)
	assert.Equal(t, 1, licences[0].AccountID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteLicencesByAccountID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(account_id = .+\)`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteLicencesByAccountID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mockUUID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(mockUUID, 1)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	l, err := repo.FindByID(ctx, mockUUID)

	assert.NoError(t, err)
	assert.NotNil(t, l)
	assert.Equal(t, 1, l.AccountID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	mockUUID := uuid.New()
	ctx := context.Background()
	err := repo.Delete(ctx, mockUUID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "licence" AS "l" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	mockUUID := uuid.New()
	l := &Licence{LicenceID: mockUUID, AccountID: 1}
	ctx := context.Background()
	updated, err := repo.Update(ctx, l)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}
