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

func (r *Repository) Create(ctx context.Context, s *Store) (*Store, error) {
	_, err := r.db.NewInsert().Model(s).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Store, error) {
	store := &Store{}
	err := r.db.NewSelect().Model(store).Where("store_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]*Store, error) {
	var stores []*Store
	err := r.db.NewSelect().Model(&stores).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *Repository) FindByBuyerID(ctx context.Context, buyerID int) ([]Store, error) {
	var stores []Store
	err := r.db.NewSelect().Model(&stores).Where("buyer_id = ?", buyerID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return stores, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Store{}).Where("store_id = ?", id).Exec(ctx)
	return err
}
