package categorie

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

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow(1, "Electronics", 1)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "cat"."categorie_id", "cat"."type", "cat"."catalog_id" FROM "categorie" AS "cat" WHERE (cat.categorie_id = 1) AND (cat.catalog_id = 1)`)).
		WillReturnRows(rows)

	ctx := context.Background()
	cat, err := repo.FindByID(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, cat)
	assert.Equal(t, "Electronics", cat.Type)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	cat := &Categorie{
		Type:      "Books",
		CatalogID: 1,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categorie"`)).
		WillReturnRows(sqlmock.NewRows([]string{"categorie_id"}).AddRow(1))

	ctx := context.Background()
	err := repo.Create(ctx, cat)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow(1, "Electronics", 1).
		AddRow(2, "Books", 1)

	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	categories, err := repo.FindAll(ctx, 1)

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Electronics", categories[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByType(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow(1, "Electronics", 1)

	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	c, err := repo.FindByType(ctx, "Electronics", 1)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Electronics", c.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(1, "Old Type", 1)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Update
	mock.ExpectExec(`^UPDATE "categorie" AS "cat" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	c := &Categorie{Type: "New Type"}
	ctx := context.Background()
	updatedCat, err := repo.Update(ctx, 1, 1, c)

	assert.NoError(t, err)
	assert.NotNil(t, updatedCat)
	assert.Equal(t, "New Type", updatedCat.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteById(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(1, "Electronics", 1)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Delete
	mock.ExpectExec(`^DELETE FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteById(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByType(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(1, "Electronics", 1)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Delete
	mock.ExpectExec(`^DELETE FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteByType(ctx, "Electronics", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
