package item

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_Create_Validations(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	_, err := svc.Create(context.Background(), Item{})
	assert.EqualError(t, err, "name is required")

	_, err = svc.Create(context.Background(), Item{Name: "A", Price: decimal.NewFromFloat(-1), Tax: decimal.NewFromFloat(0.2), CategorieID: catID})
	assert.EqualError(t, err, "price must be a positive number")

	_, err = svc.Create(context.Background(), Item{Name: "A", Price: decimal.NewFromFloat(10), Tax: decimal.NewFromFloat(2), CategorieID: catID})
	assert.EqualError(t, err, "tax must be a positive number between 0 and 1")

	_, err = svc.Create(context.Background(), Item{Name: "A", Price: decimal.NewFromFloat(10), Tax: decimal.NewFromFloat(0.2), CategorieID: uuid.Nil})
	assert.EqualError(t, err, "categorie_id is required")
}

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	testID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.item_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), testID)

	assert.ErrorIs(t, err, ErrItemNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	testID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.item_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), testID)

	assert.ErrorIs(t, err, ErrItemNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByCategorieID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	itemID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	rows := sqlmock.NewRows([]string{"item_id", "name", "price", "tax", "tax_amount", "categorie_id"}).
		AddRow(itemID, "Laptop", "999.99", "0.20", "166.67", catID)
	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.categorie_id = .+\)$`).WillReturnRows(rows)

	items, err := svc.GetByCategorieID(context.Background(), catID)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Laptop", items[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	itemID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	catID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	rows := sqlmock.NewRows([]string{"item_id", "name", "price", "tax", "tax_amount", "categorie_id"}).
		AddRow(itemID, "Keyboard", "99.99", "0.20", "16.67", catID)
	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i"$`).WillReturnRows(rows)

	items, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Keyboard", items[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByName_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.name = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByName(context.Background(), "missing")

	assert.ErrorIs(t, err, ErrItemNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	testID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.item_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), testID, ItemUpdate{})

	assert.ErrorIs(t, err, ErrItemNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
