package item

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
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

	rows := sqlmock.NewRows([]string{"item_id", "name", "price"}).
		AddRow(1, "Laptop Pro 15", "999.99")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "i"."item_id", "i"."name", "i"."price", "i"."tax", "i"."tax_amount", "i"."categorie_id" FROM "item" AS "i" WHERE (i.item_id = 1)`)).
		WillReturnRows(rows)

	ctx := context.Background()
	item, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "Laptop Pro 15", item.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	mockDbObj := &db.Db{DB: bunDB}
	repo := NewRepository(mockDbObj)

	mock.ExpectQuery(`^INSERT INTO "item"`).WillReturnRows(sqlmock.NewRows([]string{"item_id"}).AddRow(1))

	item := &Item{Name: "New Item", Price: decimal.NewFromFloat(100.0), Tax: decimal.NewFromFloat(0.20)}
	ctx := context.Background()
	err := repo.Create(ctx, item)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	mockDbObj := &db.Db{DB: bunDB}
	repo := NewRepository(mockDbObj)

	rows := sqlmock.NewRows([]string{"item_id", "name"}).AddRow(1, "Item 1").AddRow(2, "Item 2")
	mock.ExpectQuery(`^SELECT "i"\."item_id", "i"\."name", "i"\."price", "i"\."tax", "i"\."tax_amount", "i"\."categorie_id" FROM "item" AS "i"$`).WillReturnRows(rows)

	ctx := context.Background()
	items, err := repo.FindAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"item_id", "name"}).AddRow(1, "Laptop")
	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.name = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	item, err := repo.FindByName(ctx, "Laptop")

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "Laptop", item.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "item" AS "i" WHERE \(item_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteByID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	// find by id mock
	rows := sqlmock.NewRows([]string{"item_id", "name", "price", "tax"}).AddRow(1, "Laptop", "100.0", "0.2")
	mock.ExpectQuery(`^SELECT .* FROM "item" AS "i" WHERE \(i\.item_id = .+\)$`).WillReturnRows(rows)

	// update mock
	mock.ExpectExec(`^UPDATE "item" AS "i" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	newName := "Laptop 2"
	updateData := ItemUpdate{Name: &newName}

	ctx := context.Background()
	updatedItem, err := repo.Update(ctx, 1, updateData)

	assert.NoError(t, err)
	assert.NotNil(t, updatedItem)
	assert.Equal(t, "Laptop 2", updatedItem.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
