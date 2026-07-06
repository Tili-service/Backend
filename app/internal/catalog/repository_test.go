package catalog

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
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

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).
		AddRow("Winter 2026 Collection", "All items available for the winter 2026 season", uuid.New())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "c"."name", "c"."description", "c"."store_id" FROM "catalog" AS "c" WHERE (c.catalog_id = `)).
		WillReturnRows(rows)

	ctx := context.Background()
	storeID := uuid.New()
	catalogID := uuid.New()
	c, err := repo.FindByID(ctx, catalogID, storeID)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Winter 2026 Collection", c.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).
		AddRow("Cat 1", "Desc 1", 1).
		AddRow("Cat 2", "Desc 2", 1)

	mock.ExpectQuery(`^SELECT "c"\."name", "c"\."description", "c"\."store_id" FROM "catalog" AS "c" WHERE \(c\.store_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	storeID := uuid.New()
	catalogs, err := repo.FindAll(ctx, storeID)

	assert.NoError(t, err)
	assert.Len(t, catalogs, 2)
	assert.Equal(t, "Cat 1", catalogs[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).
		AddRow("Test Cat", "Test Desc", 1)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.name = .+\) AND \(c\.store_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	storeID := uuid.New()
	c, err := repo.FindByName(ctx, "Test Cat", storeID)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Test Cat", c.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	mock.ExpectExec(`^DELETE FROM "catalog" AS "c" WHERE \(catalog_id = .+\) AND \(store_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	storeID := uuid.New()
	catalogID := uuid.New()
	err := repo.DeleteByID(ctx, catalogID, storeID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	mock.ExpectExec(`^DELETE FROM "catalog" AS "c" WHERE \(name = .+\) AND \(store_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	storeID := uuid.New()
	err := repo.DeleteByName(ctx, "Test Cat", storeID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	// Mock FindByID inside Update
	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).
		AddRow("Old Name", "Old Desc", 1)
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\) AND \(c\.store_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Update
	mock.ExpectExec(`^UPDATE "catalog" AS "c" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	newName := "New Name"
	input := catalogUpdate{Name: &newName}
	ctx := context.Background()
	storeID := uuid.New()
	catalogID := uuid.New()
	c, err := repo.Update(ctx, catalogID, storeID, input)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "New Name", c.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
