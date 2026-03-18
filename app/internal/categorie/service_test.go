package categorie

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

	_, err := svc.Create(context.Background(), Categorie{})

	assert.EqualError(t, err, "type is required")
}

func TestService_FindByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.FindByID(context.Background(), 1)

	assert.ErrorIs(t, err, ErrCategorieNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.DeleteByID(context.Background(), 1)

	assert.ErrorIs(t, err, ErrCategorieNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	rows := sqlmock.NewRows([]string{"categorie_id", "type"}).AddRow(1, "Books")
	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat"$`).WillReturnRows(rows)

	list, err := svc.FindAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Books", list[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.categorie_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(context.Background(), 1, Categorie{Type: "Books"})

	assert.ErrorIs(t, err, ErrCategorieNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_FindByType_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.type = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.FindByType(context.Background(), "missing")

	assert.ErrorIs(t, err, ErrCategorieNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByType_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "categorie" AS "cat" WHERE \(cat\.type = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.DeleteByType(context.Background(), "missing")

	assert.ErrorIs(t, err, ErrCategorieNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
