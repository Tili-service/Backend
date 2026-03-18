package account

import (
	"context"
	"regexp"
	"testing"
	"time"

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

	expectedAccount := &Account{
		AccountID:        1,
		Email:            "test@example.com",
		Name:             "Test User",
		Password:         "hashedpassword",
		StripeCustomerID: "cus_123",
		CreatedAt:        time.Now(),
	}

	rows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(expectedAccount.AccountID, expectedAccount.Email, expectedAccount.Name, expectedAccount.Password, expectedAccount.StripeCustomerID, expectedAccount.CreatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a"."account_id", "a"."email", "a"."password", "a"."name", "a"."stripe_customer_id", "a"."created_at" FROM "account" AS "a" WHERE (account_id = 1)`)).
		WillReturnRows(rows)

	ctx := context.Background()
	acc, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, "test@example.com", acc.Email)
	assert.Equal(t, "Test User", acc.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByEmail(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	expectedAccount := &Account{
		AccountID:        1,
		Email:            "test@example.com",
		Name:             "Test User",
		Password:         "hashedpassword",
		StripeCustomerID: "cus_123",
		CreatedAt:        time.Now(),
	}

	rows := sqlmock.NewRows([]string{"account_id", "email", "name", "password", "stripe_customer_id", "created_at"}).
		AddRow(expectedAccount.AccountID, expectedAccount.Email, expectedAccount.Name, expectedAccount.Password, expectedAccount.StripeCustomerID, expectedAccount.CreatedAt)

	mock.ExpectQuery(`^SELECT .* FROM "account" AS "a" WHERE \(email = .+\)$`).
		WillReturnRows(rows)

	ctx := context.Background()
	acc, err := repo.FindByEmail(ctx, "test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, "test@example.com", acc.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	acc := &Account{
		Email:            "new@example.com",
		Name:             "New User",
		Password:         "hashedpassword",
		StripeCustomerID: "cus_456",
		CreatedAt:        time.Now(),
	}

	mock.ExpectQuery(`^INSERT INTO "account"`).
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(2))

	ctx := context.Background()
	err := repo.Create(ctx, acc)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	mock.ExpectExec(`^DELETE FROM "account" AS "a" WHERE \(account_id = 1\)$`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.Delete(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := &Repository{db: bunDB}

	acc := &Account{
		AccountID:        1,
		Email:            "updated@example.com",
		Name:             "Updated User",
		Password:         "hashed",
		StripeCustomerID: "cus_789",
		CreatedAt:        time.Now(),
	}

	mock.ExpectExec(`^UPDATE "account" AS "a" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	updatedAcc, err := repo.Update(ctx, acc)

	assert.NoError(t, err)
	assert.NotNil(t, updatedAcc)
	assert.Equal(t, "updated@example.com", updatedAcc.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}
