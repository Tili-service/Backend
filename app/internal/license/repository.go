package license

import (
	"context"

	"tili/app/pkg/db"
	"tili/app/internal/store"
	"tili/app/internal/account"
	"tili/app/internal/user"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func (r *Repository) CreateAccount(ctx context.Context, u *account.Account) error {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	return err
}

func (r *Repository) CreateUserAdmin(ctx context.Context, u *user.User) error {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	return err
}

func (r *Repository) CreateStore(ctx context.Context, u *store.Store) error {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	return err
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model(&account.Account{}).Where("account_id = ?", id).Exec(ctx)
	return err
}
