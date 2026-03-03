package store

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

func (r *Repository) Create(ctx context.Context, u *Store) (*Store, error) {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Store, error) {
	store := &Store{}
	err := r.db.NewSelect().Model(store).Where("store_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (r *Repository) FindByAccountID(ctx context.Context, accountID int64) (*Store, error) {
	store := &Store{}
	err := r.db.NewSelect().Model(store).Where("account_id = ?", accountID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model(&Store{}).Where("store_id = ?", id).Exec(ctx)
	return err
}
