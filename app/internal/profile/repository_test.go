package profile

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
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

func TestRepository_FindByID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	mockDbObj := &db.Db{DB: bunDB}
	repo := NewRepository(mockDbObj)

	profID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")

	rows := sqlmock.NewRows([]string{"profile_id", "store_id", "name", "level_access", "is_active"}).
		AddRow(profID, storeID, "Admin", 1, true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "p"."profile_id", "p"."store_id", "p"."name", "p"."pin", "p"."level_access", "p"."is_active" FROM "profile" AS "p" WHERE (p.profile_id = '00000000-0000-0000-0000-000000000001')`)).
		WillReturnRows(rows)

	ctx := context.Background()
	p, err := repo.FindByID(ctx, profID)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "Admin", p.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Create(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	profID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")

	mock.ExpectQuery(`^INSERT INTO "profile"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(profID))

	p := &Profile{StoreID: storeID, Name: "User", Pin: "1234", LevelAccess: 1, IsActive: true}
	ctx := context.Background()
	err := repo.Create(ctx, p)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindByStoreAndPin(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	profID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")

	rows := sqlmock.NewRows([]string{"profile_id", "name"}).AddRow(profID, "User")
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(p\.store_id = .+\) AND \(p\.pin = .+\) AND \(p\.is_active = .+\)$`).WillReturnRows(rows)

	ctx := context.Background()
	p, err := repo.FindByStoreAndPin(ctx, storeID, "1234")

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "User", p.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(profile_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	profID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	ctx := context.Background()
	err := repo.Delete(ctx, profID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteByStoreID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	ctx := context.Background()
	err := repo.DeleteByStoreID(ctx, storeID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_PinExistsInStore(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`^SELECT EXISTS \(SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\) AND \(pin = .+\)\)$`).WillReturnRows(rows)

	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	ctx := context.Background()
	exists, err := repo.PinExistsInStore(ctx, storeID, "1234")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	mock.ExpectExec(`^UPDATE "profile" AS "p" SET`).WillReturnResult(sqlmock.NewResult(1, 1))

	profID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	p := &Profile{ProfileID: profID, StoreID: storeID, Name: "Updated", Pin: "4321", LevelAccess: 2, IsActive: false}
	ctx := context.Background()
	err := repo.Update(ctx, p)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_FindAllProfilesByStoreID(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()
	repo := NewRepository(&db.Db{DB: bunDB})

	profID1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	profID2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	rows := sqlmock.NewRows([]string{"profile_id", "name"}).AddRow(profID1, "User 1").AddRow(profID2, "User 2")
	mock.ExpectQuery(`^SELECT .* FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnRows(rows)

	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	ctx := context.Background()
	profiles, err := repo.FindAllProfilesByStoreID(ctx, storeID)

	assert.NoError(t, err)
	assert.Len(t, profiles, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
