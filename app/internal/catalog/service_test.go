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
	uuid1 := uuid.New()
	_, err := svc.Create(context.Background(), uuid1, catalogUpdate{Description: &desc})

	assert.EqualError(t, err, "name is required")
}

func TestService_Update_ValidationError(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)
	uuid1 := uuid.New()

	_, err := svc.Update(context.Background(), uuid1, uuid1, catalogUpdate{})

	assert.EqualError(t, err, "at least one field is required")
}

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.catalog_id = .+\) AND \(c\.store_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	uuid1 := uuid.New()
	_, err := svc.GetByID(context.Background(), uuid1, uuid1)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	storeID := uuid.New()
	catalogID := uuid.New()

	rows := sqlmock.NewRows([]string{"catalog_id", "name", "description", "store_id"}).
		AddRow(catalogID, "Cat 1", "Desc 1", storeID)
	mock.ExpectQuery(`^SELECT .* FROM "catalog" AS "c" WHERE \(c\.store_id = .+\)$`).WillReturnRows(rows)

	list, err := svc.GetAll(context.Background(), storeID)

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

	uuid1 := uuid.New()
	err := svc.Delete(context.Background(), uuid1, uuid1)

	assert.ErrorIs(t, err, ErrCatalogNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
