package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_FindByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.FindByID(context.Background(), 1)

	assert.ErrorIs(t, err, ErrStoreNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByID_InternalError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("db fail"))

	_, err := svc.FindByID(context.Background(), 1)

	assert.EqualError(t, err, "db fail")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(1, "S1", 12)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	store, err := svc.FindByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.Equal(t, "S1", store.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), 1)

	assert.ErrorIs(t, err, ErrStoreNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_InternalFindError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("select failed"))

	err := svc.Delete(context.Background(), 1)

	assert.EqualError(t, err, "select failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_DeleteError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(1, "S1", 12)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("delete failed"))

	err := svc.Delete(context.Background(), 1)

	assert.EqualError(t, err, "delete failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(1, "S1", 12)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Create_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^INSERT INTO "store"`).WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow(1))

	licenceID := uuid.New()
	store, err := svc.Create(context.Background(), CreateStoreInput{Name: "Store", LicenceID: licenceID}, 12)

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.Equal(t, 12, store.BuyerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByBuyerID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(1, "Buyer Store", 12)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(rows)

	stores, err := svc.FindByBuyerID(context.Background(), 12)

	assert.NoError(t, err)
	assert.Len(t, stores, 1)
	assert.Equal(t, "Buyer Store", stores[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name"}).AddRow(1, "Store A")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s"$`).WillReturnRows(rows)

	stores, err := svc.FindAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, stores, 1)
	assert.Equal(t, "Store A", stores[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindAll_Error(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s"$`).WillReturnError(errors.New("query fail"))

	stores, err := svc.FindAll(context.Background())

	assert.Nil(t, stores)
	assert.EqualError(t, err, "query fail")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	name := "updated"
	_, err := svc.Update(context.Background(), 1, UpdateStoreInput{Name: &name})

	assert.ErrorIs(t, err, ErrStoreNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_InternalFindError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("select failed"))

	name := "updated"
	_, err := svc.Update(context.Background(), 1, UpdateStoreInput{Name: &name})

	assert.EqualError(t, err, "select failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_RepoUpdateError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "numero_tva", "siret"}).AddRow(1, "Store A", 12, "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnError(errors.New("update failed"))

	name := "updated"
	_, err := svc.Update(context.Background(), 1, UpdateStoreInput{Name: &name})

	assert.EqualError(t, err, "update failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "numero_tva", "siret"}).AddRow(1, "Store A", 12, "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	name := "updated"
	numeroTVA := "FR123"
	siret := "SIRET123"
	store, err := svc.Update(context.Background(), 1, UpdateStoreInput{Name: &name, NumeroTVA: &numeroTVA, Siret: &siret})

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.Equal(t, "updated", store.Name)
	assert.Equal(t, "FR123", store.NumeroTVA)
	assert.Equal(t, "SIRET123", store.Siret)
	assert.NoError(t, mock.ExpectationsWereMet())
}
