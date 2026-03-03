package license

import (
	"context"

	"tili/app/internal/account"
	"tili/app/pkg/db"

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

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model(&account.Account{}).Where("account_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) FindByID(ctx context.Context, accountID int64) (*account.Account, error) {
	account := &account.Account{}
	err := r.db.NewSelect().Model(account).Where("account_id = ?", accountID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return account, nil
}
