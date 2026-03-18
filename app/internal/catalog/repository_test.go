package catalog

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

	rows := sqlmock.NewRows([]string{"name", "description"}).
		AddRow("Winter 2026 Collection", "All items available for the winter 2026 season")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "c"."name", "c"."description" FROM "catalog" AS "c" WHERE (c.catalog_id = 1)`)).
		WillReturnRows(rows)

	ctx := context.Background()
	c, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Winter 2026 Collection", c.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	c := &catalog{
		Name:        "Summer 2026",
		Description: "Summer clothes",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "catalog"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.Create(ctx, c)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"name", "description"}).
		AddRow("Cat 1", "Desc 1").
		AddRow("Cat 2", "Desc 2")

	mock.ExpectQuery(`^SELECT "c"."name", "c"."description" FROM "catalog" AS "c"$`).
		WillReturnRows(rows)

	ctx := context.Background()
	catalogs, err := repo.FindAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, catalogs, 2)
	assert.Equal(t, "Cat 1", catalogs[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"name", "description"}).
		AddRow("Test Cat", "Test Desc")

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c.name = \?|'Test Cat'\)`).
		WillReturnRows(rows)

	ctx := context.Background()
	c, err := repo.FindByName(ctx, "Test Cat")

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Test Cat", c.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	mock.ExpectExec(`^DELETE FROM "catalog" AS "c" WHERE \(catalog_id = 1\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteByID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	mock.ExpectExec(`^DELETE FROM "catalog" AS "c" WHERE \(name = '.+'|\?\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteByName(ctx, "Test Cat")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	// Mock FindByID inside Update
	rows := sqlmock.NewRows([]string{"name", "description"}).
		AddRow("Old Name", "Old Desc")
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = \?|1\)$`).
		WillReturnRows(rows)

	// Mock Update
	mock.ExpectExec(`^UPDATE "catalog" AS "c" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	newName := "New Name"
	input := catalogUpdate{Name: &newName}
	ctx := context.Background()
	c, err := repo.Update(ctx, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "New Name", c.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}
