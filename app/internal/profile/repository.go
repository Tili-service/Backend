package profile

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

func (r *Repository) Create(ctx context.Context, p *Profile) error {
	_, err := r.db.NewInsert().Model(p).Exec(ctx)
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Profile, error) {
	p := new(Profile)
	err := r.db.NewSelect().Model(p).Where("p.profile_id = ?", id).Scan(ctx)
	return p, err
}

func (r *Repository) FindByStoreAndPin(ctx context.Context, storeID int, pin string) (*Profile, error) {
	p := new(Profile)
	err := r.db.NewSelect().Model(p).
		Where("p.store_id = ?", storeID).
		Where("p.pin = ?", pin).
		Where("p.is_active = ?", true).
		Scan(ctx)
	return p, err
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Profile{}).Where("profile_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) DeleteByStoreID(ctx context.Context, storeID int) error {
	_, err := r.db.NewDelete().Model(&Profile{}).Where("store_id = ?", storeID).Exec(ctx)
	return err
}

func (r *Repository) PinExistsInStore(ctx context.Context, storeID int, pin string) (bool, error) {
	exists, err := r.db.NewSelect().Model(&Profile{}).
		Where("store_id = ?", storeID).
		Where("pin = ?", pin).
		Exists(ctx)
	return exists, err
}
