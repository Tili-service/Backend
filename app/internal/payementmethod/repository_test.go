package payementmethod

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	// simulate existing check failing (not found)
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`^INSERT INTO "payementmethod"`).
		WillReturnRows(sqlmock.NewRows([]string{"payement_method_id"}).AddRow(1))

	pm := &PayementMethod{Name: "Card"}
	ctx := context.Background()
	err := repo.Create(ctx, pm)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAll(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"payement_method_id", "name", "is_active"}).
		AddRow(1, "Card", true).
		AddRow(2, "Cash", true)
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(is_active = .*\)$`).WillReturnRows(rows)

	ctx := context.Background()
	pms, err := repo.FindAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, pms, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"payement_method_id", "name"}).AddRow(1, "Card")
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	pm, err := repo.FindByName(ctx, "Card")

	assert.NoError(t, err)
	assert.NotNil(t, pm)
	assert.Equal(t, "Card", pm.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeactivateByName(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "payementmethod" AS "pm" SET .* WHERE \(name = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeactivateByName(ctx, "Card")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"payement_method_id", "name"}).AddRow(1, "Card")
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	pm, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, pm)
	assert.Equal(t, "Card", pm.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeactivateByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "payementmethod" AS "pm" SET .* WHERE \(payement_method_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.DeactivateByID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_ReactivateByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "payementmethod" AS "pm" SET .* WHERE \(payement_method_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.ReactivateByID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindActiveByID_Found(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"payement_method_id", "name", "is_active"}).AddRow(1, "Card", true)
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .* AND is_active = .*\)$`).WillReturnRows(rows)

	ctx := context.Background()
	err := repo.FindActiveByID(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindActiveByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .* AND is_active = .*\)$`).WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	err := repo.FindActiveByID(ctx, 99)

	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	// existing check
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`^UPDATE "payementmethod" AS "pm" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	pm := &PayementMethod{PayementMethodID: 1, Name: "Updated Card"}
	ctx := context.Background()
	err := repo.Update(ctx, pm)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
