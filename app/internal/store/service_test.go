package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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

	uuid1 := uuid.New()
	_, err := svc.FindByID(context.Background(), uuid1)

	assert.ErrorIs(t, err, ErrStoreNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByID_InternalError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("db fail"))

	uuid1 := uuid.New()
	_, err := svc.FindByID(context.Background(), uuid1)

	assert.EqualError(t, err, "db fail")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(storeID, "S1", buyerID)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	uuid1 := uuid.New()
	store, err := svc.FindByID(context.Background(), uuid1)

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

	uuid1 := uuid.New()
	err := svc.Delete(context.Background(), uuid1)

	assert.ErrorIs(t, err, ErrStoreNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_InternalFindError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("select failed"))

	uuid1 := uuid.New()
	err := svc.Delete(context.Background(), uuid1)

	assert.EqualError(t, err, "select failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_DeleteError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(storeID, "S1", buyerID)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnError(errors.New("delete failed"))

	uuid1 := uuid.New()
	err := svc.Delete(context.Background(), uuid1)

	assert.EqualError(t, err, "delete failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(storeID, "S1", buyerID)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	uuid1 := uuid.New()
	err := svc.Delete(context.Background(), uuid1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Create_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	mock.ExpectQuery(`^INSERT INTO "store"`).WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow(storeID))

	licenceID := uuid.New()
	buyerID := uuid.New()
	store, err := svc.Create(context.Background(), CreateStoreInput{Name: "Store", LicenceID: licenceID}, buyerID)

	assert.NoError(t, err)
	assert.NotNil(t, store)
	assert.Equal(t, buyerID, store.BuyerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByBuyerID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id"}).AddRow(storeID, "Buyer Store", buyerID)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(rows)

	stores, err := svc.FindByBuyerID(context.Background(), buyerID)

	assert.NoError(t, err)
	assert.Len(t, stores, 1)
	assert.Equal(t, "Buyer Store", stores[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByLicenceID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	licenceID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation"}).AddRow(storeID, "Licence Store", buyerID, licenceID, time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	store, err := svc.FindByLicenceID(context.Background(), licenceID)

	assert.NoError(t, err)
	if assert.NotNil(t, store) {
		assert.Equal(t, licenceID, store.LicenceID)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	uuid1 := uuid.New()
	err := svc.DeleteByID(context.Background(), uuid1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	uuid1 := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name"}).AddRow(uuid1, "Store A")
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
	uuid1 := uuid.New()
	_, err := svc.Update(context.Background(), uuid1, UpdateStoreInput{Name: &name})

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
	uuid1 := uuid.New()
	_, err := svc.Update(context.Background(), uuid1, UpdateStoreInput{Name: &name})

	assert.EqualError(t, err, "select failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_RepoUpdateError(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "numero_tva", "siret"}).AddRow(storeID, "Store A", buyerID, "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnError(errors.New("update failed"))

	name := "updated"
	uuid1 := uuid.New()
	_, err := svc.Update(context.Background(), uuid1, UpdateStoreInput{Name: &name})

	assert.EqualError(t, err, "update failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "numero_tva", "siret"}).AddRow(storeID, "Store A", buyerID, "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	name := "updated"
	numeroTVA := "FR123"
	siret := "SIRET123"
	uuid1 := uuid.New()
	store, err := svc.Update(context.Background(), uuid1, UpdateStoreInput{Name: &name, NumeroTVA: &numeroTVA, Siret: &siret})

	assert.NoError(t, err)
	if assert.NotNil(t, store) {
		assert.Equal(t, "updated", store.Name)
		assert.Equal(t, "FR123", store.NumeroTVA)
		assert.Equal(t, "SIRET123", store.Siret)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
