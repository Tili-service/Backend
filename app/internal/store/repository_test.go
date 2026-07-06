package store

import (
	"context"
	"regexp"
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

func TestRepository_FindByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	mockDbObj := &db.Db{DB: bunDB}
	repo := NewRepository(mockDbObj)

	mockUUID := uuid.New()
	storeID := uuid.New()
	buyerID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation"}).
		AddRow(storeID, "My Store", buyerID, mockUUID, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "s"."store_id", "s"."name", "s"."buyer_id", "s"."licence_id", "s"."date_creation", "s"."numero_tva", "s"."siret" FROM "store" AS "s" WHERE (store_id = `)).
		WillReturnRows(rows)

	ctx := context.Background()
	s, err := repo.FindByID(ctx, storeID)

	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, "My Store", s.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectQuery(`^INSERT INTO "store"`).WillReturnRows(sqlmock.NewRows([]string{"store_id"}).AddRow(1))

	s := &Store{Name: "New Store", BuyerID: uuid.New()}
	ctx := context.Background()
	created, err := repo.Create(ctx, s)

	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"store_id", "name"}).AddRow(uuid.New(), "Store 1").AddRow(uuid.New(), "Store 2")
	mock.ExpectQuery(`^SELECT "s"\."store_id", "s"\."name", "s"\."buyer_id", "s"\."licence_id", "s"\."date_creation", "s"\."numero_tva", "s"\."siret" FROM "store" AS "s"$`).WillReturnRows(rows)

	ctx := context.Background()
	stores, err := repo.FindAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, stores, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByBuyerID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"store_id", "name"}).AddRow(uuid.New(), "Buyer Store")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	buyerID := uuid.New()
	stores, err := repo.FindByBuyerID(ctx, buyerID)

	assert.NoError(t, err)
	assert.Len(t, stores, 1)
	assert.Equal(t, "Buyer Store", stores[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByLicenceID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	licenceID := uuid.New()
	rows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation"}).AddRow(uuid.New(), "Licence Store", uuid.New(), licenceID, time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	st, err := repo.FindByLicenceID(ctx, licenceID)

	assert.NoError(t, err)
	if assert.NotNil(t, st) {
		assert.Equal(t, "Licence Store", st.Name)
		assert.Equal(t, licenceID, st.LicenceID)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	storeID := uuid.New()
	err := repo.Delete(ctx, storeID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "store" AS "s" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	storeID := uuid.New()
	s := &Store{StoreID: storeID, Name: "Updated Store", BuyerID: uuid.New()}
	ctx := context.Background()
	updated, err := repo.Update(ctx, s)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Store", updated.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
