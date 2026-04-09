package account

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"tili/app/internal/profile"
	"tili/app/internal/store"
	"tili/app/pkg/db"
)

type testLicenseDeleter struct {
	err error
}

func (t *testLicenseDeleter) DeleteByAccountID(ctx context.Context, accountID int) error {
	return t.err
}

func setupAccountService(t *testing.T, licErr error) (*Service, sqlmock.Sqlmock, func()) {
	bunDB, mock := setupMockDB(t)

	mockDB := &db.Db{DB: bunDB}
	accRepo := NewRepository(mockDB)
	storeRepo := store.NewRepository(mockDB)
	profileRepo := profile.NewRepository(mockDB)

	svc := NewService(
		accRepo,
		store.NewService(storeRepo),
		profile.NewService(profileRepo),
		&testLicenseDeleter{err: licErr},
	)

	cleanup := func() { _ = bunDB.Close() }
	return svc, mock, cleanup
}

func TestServiceCreate_EmailExists(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"account_id", "email", "password", "name", "stripe_customer_id", "created_at"}).
		AddRow(1, "a@a.com", "hash", "A", "cus_1", time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnRows(rows)

	_, err := svc.Create(context.Background(), RegistrationInput{Name: "A", Email: "a@a.com", Password: "secret123"})

	assert.ErrorIs(t, err, ErrEmailExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceCreate_FindByEmailUnexpectedError(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnError(sql.ErrConnDone)

	_, err := svc.Create(context.Background(), RegistrationInput{Name: "A", Email: "a@a.com", Password: "secret123"})

	assert.ErrorIs(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceLogin_EmailNotFound(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, _, err := svc.Login(context.Background(), LoginInput{Email: "a@a.com", Password: "secret123"})

	assert.EqualError(t, err, "invalid email or password")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceLogin_WrongPassword(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	rows := sqlmock.NewRows([]string{"account_id", "email", "password", "name", "stripe_customer_id", "created_at"}).
		AddRow(1, "a@a.com", string(hash), "A", "cus_1", time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnRows(rows)

	_, _, err = svc.Login(context.Background(), LoginInput{Email: "a@a.com", Password: "wrong"})

	assert.EqualError(t, err, "invalid email or password")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceLogin_Success(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	accRows := sqlmock.NewRows([]string{"account_id", "email", "password", "name", "stripe_customer_id", "created_at"}).
		AddRow(1, "a@a.com", string(hash), "A", "cus_1", time.Now())
	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).WillReturnRows(accRows)

	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "Store", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(storeRows)

	acc, stores, err := svc.Login(context.Background(), LoginInput{Email: "a@a.com", Password: "secret123"})

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Len(t, stores, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceExists_Branches(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc, mock, cleanup := setupAccountService(t, nil)
		defer cleanup()

		mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(account_id = .+\)$`).WillReturnError(sql.ErrNoRows)

		exists, err := svc.Exists(context.Background(), 1)
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("found", func(t *testing.T) {
		svc, mock, cleanup := setupAccountService(t, nil)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"account_id", "email", "password", "name", "stripe_customer_id", "created_at"}).
			AddRow(1, "a@a.com", "hash", "A", "cus_1", time.Now())
		mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(account_id = .+\)$`).WillReturnRows(rows)

		exists, err := svc.Exists(context.Background(), 1)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestServiceFullDelete_FindStoresError(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnError(sql.ErrConnDone)

	err := svc.FullDelete(context.Background(), 1)

	assert.EqualError(t, err, "failed to find stores")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceFullDelete_SuccessWithStore(t *testing.T) {
	svc, mock, cleanup := setupAccountService(t, nil)
	defer cleanup()

	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "Store", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(storeRows)

	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	storeByIDRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).
		AddRow(1, "Store", 1, uuid.New(), time.Now(), "", "")
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnRows(storeByIDRows)
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "account" AS "a" WHERE \(account_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.FullDelete(context.Background(), 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceFullDelete_LicenseDeleteError(t *testing.T) {
	licErr := errors.New("license delete failed")
	svc, mock, cleanup := setupAccountService(t, licErr)
	defer cleanup()

	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"})
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(buyer_id = .+\)$`).WillReturnRows(storeRows)

	err := svc.FullDelete(context.Background(), 1)

	assert.EqualError(t, err, "license delete failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}
