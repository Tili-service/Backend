package catalog

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_Create_ValidationError(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	desc := "desc"
	_, err := svc.Create(context.Background(), catalogUpdate{Description: &desc})

	assert.EqualError(t, err, "name is required")
}

func TestService_Update_ValidationError(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), 1, catalogUpdate{})

	assert.EqualError(t, err, "at least one field is required")
}

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), 123)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"name", "description"}).
		AddRow("Cat 1", "Desc 1")
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c"$`).WillReturnRows(rows)

	list, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Cat 1", list[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), 10)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
