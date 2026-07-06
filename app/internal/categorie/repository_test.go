package categorie

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

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow(catID, "Electronics", catalogID)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "cat"."categorie_id", "cat"."type", "cat"."catalog_id" FROM "categorie" AS "cat" WHERE (cat.categorie_id = '00000000-0000-0000-0000-000000000001') AND (cat.catalog_id = '00000000-0000-0000-0000-000000000002')`)).
		WillReturnRows(rows)

	ctx := context.Background()
	cat, err := repo.FindByID(ctx, catID, catalogID)

	assert.NoError(t, err)
	assert.NotNil(t, cat)
	assert.Equal(t, "Electronics", cat.Type)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	cat := &Categorie{
		Type:      "Books",
		CatalogID: catalogID,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categorie"`)).
		WillReturnRows(sqlmock.NewRows([]string{"categorie_id"}).AddRow("00000000-0000-0000-0000-000000000001"))

	ctx := context.Background()
	err := repo.Create(ctx, cat)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow("00000000-0000-0000-0000-000000000001", "Electronics", catalogID).
		AddRow("00000000-0000-0000-0000-000000000003", "Books", catalogID)

	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	categories, err := repo.FindAll(ctx, catalogID)

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Electronics", categories[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByType(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).
		AddRow("00000000-0000-0000-0000-000000000001", "Electronics", catalogID)

	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	c, err := repo.FindByType(ctx, "Electronics", catalogID)

	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "Electronics", c.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(catID, "Old Type", catalogID)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Update
	mock.ExpectExec(`^UPDATE "categorie" AS "cat" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	c := &Categorie{Type: "New Type"}
	ctx := context.Background()
	updatedCat, err := repo.Update(ctx, catID, catalogID, c)

	assert.NoError(t, err)
	assert.NotNil(t, updatedCat)
	assert.Equal(t, "New Type", updatedCat.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteById(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(catID, "Electronics", catalogID)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Delete
	mock.ExpectExec(`^DELETE FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteById(ctx, catID, catalogID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByType(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	catID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	catalogID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// Mock Find
	rows := sqlmock.NewRows([]string{"categorie_id", "type", "catalog_id"}).AddRow(catID, "Electronics", catalogID)
	mock.ExpectQuery(`^SELECT "cat"\."categorie_id", "cat"\."type", "cat"\."catalog_id" FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnRows(rows)

	// Mock Delete
	mock.ExpectExec(`^DELETE FROM "categorie" AS "cat" WHERE \(cat\.type = .+\) AND \(cat\.catalog_id = .+\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeleteByType(ctx, "Electronics", catalogID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
