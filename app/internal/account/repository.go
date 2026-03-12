package account

import (
	"context"

	"tili/app/pkg/db"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func (r *Repository) Create(ctx context.Context, a *Account) error {
	_, err := r.db.NewInsert().Model(a).Exec(ctx)
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Account, error) {
	a := &Account{}
	err := r.db.NewSelect().Model(a).Where("account_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	a := &Account{}
	err := r.db.NewSelect().Model(a).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Account{}).Where("account_id = ?", id).Exec(ctx)
	return err
}
