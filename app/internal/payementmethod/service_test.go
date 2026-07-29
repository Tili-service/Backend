package payementmethod

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

	_, err := svc.Create(context.Background(), PayementMethod{})

	assert.EqualError(t, err, "name is required")
}

func TestService_Update_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	pmID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	_, err := svc.Update(context.Background(), pmID, PayementMethod{Name: "Card"})

	assert.ErrorIs(t, err, ErrPayementMethodNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.Delete(context.Background(), "Card")

	assert.ErrorIs(t, err, ErrPayementMethodNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetAll_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	pmID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	rows := sqlmock.NewRows([]string{"payement_method_id", "name", "is_active"}).AddRow(pmID, "Card", true)
	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(is_active = .*\)$`).WillReturnRows(rows)

	list, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Card", list[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	pmID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	err := svc.DeleteByID(context.Background(), pmID)

	assert.ErrorIs(t, err, ErrPayementMethodNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Reactivate_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(payement_method_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	pmID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	err := svc.Reactivate(context.Background(), pmID)

	assert.ErrorIs(t, err, ErrPayementMethodNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByName_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "payementmethod" AS "pm" WHERE \(name = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByName(context.Background(), "missing")

	assert.ErrorIs(t, err, ErrPayementMethodNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
