package catalog

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_Create_ValidationError(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	desc := "desc"
	uuid, _ := uuid.NewUUID()
	_, err := svc.Create(context.Background(), uuid, catalogUpdate{Description: &desc})

	assert.EqualError(t, err, "name is required")
}

func TestService_Update_ValidationError(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	uuid, _ := uuid.NewUUID()

	_, err := svc.Update(context.Background(), uuid, uuid, catalogUpdate{})

	assert.EqualError(t, err, "at least one field is required")
}

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\) AND \(c\.store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	uuid, _ := uuid.NewUUID()
	_, err := svc.GetByID(context.Background(), uuid, uuid)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"name", "description", "store_id"}).
		AddRow("Cat 1", "Desc 1", 1)
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.store_id = .+\)$`).WillReturnRows(rows)

	uuid, _ := uuid.NewUUID()
	list, err := svc.GetAll(context.Background(), uuid)

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

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\) AND \(c\.store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	uuid, _ := uuid.NewUUID()
	err := svc.Delete(context.Background(), uuid, uuid)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
