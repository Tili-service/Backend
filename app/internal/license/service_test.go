package license

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tili/app/pkg/db"
)

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), uuid.New())

	assert.ErrorIs(t, err, ErrLicenceNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_Forbidden(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 99)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	err := svc.Delete(context.Background(), 1, licID)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_Forbidden(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 99)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	_, err := svc.Update(context.Background(), 1, licID, UpdateLicenceInput{})

	assert.ErrorIs(t, err, ErrForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreatePaymentLink_InvalidOffer(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	_, err := svc.CreatePaymentLink(context.Background(), 1, "", CreatePaymentLinkInput{Offer: "weekly"})

	assert.EqualError(t, err, "offre invalide: weekly")
}

func TestService_CreatePaymentLink_MissingConfig(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	t.Setenv("STRIPE_PRICE_MENSUEL", "")
	t.Setenv("APP_URL", "https://example.com")

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	_, err := svc.CreatePaymentLink(context.Background(), 1, "", CreatePaymentLinkInput{Offer: "mensuel"})

	assert.EqualError(t, err, "config manquante pour l'offre: mensuel")
}

func TestService_GetByAccountID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	licID := uuid.New()
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, 1)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l"`).WillReturnRows(rows)

	list, err := svc.GetByAccountID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, 1, list[0].AccountID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByAccountID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(account_id = .+\)`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.DeleteByAccountID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Create_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo)

	mock.ExpectExec(`^INSERT INTO "licence"`).WillReturnResult(sqlmock.NewResult(1, 1))

	lic, err := svc.Create(context.Background(), 1, CreateLicenceInput{DurationDays: 30, Transaction: "txn_123"})

	assert.NoError(t, err)
	if assert.NotNil(t, lic) {
		assert.Equal(t, 1, lic.AccountID)
		assert.True(t, lic.IsActive)
		assert.WithinDuration(t, time.Now().Add(30*24*time.Hour), lic.Expiration, time.Second)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
