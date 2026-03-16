package payementmethod

import (
	"context"
	"errors"

	"tili/app/pkg/db"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func (r *Repository) Create(ctx context.Context, pm *PayementMethod) error {
	existing := new(PayementMethod)
	err := r.db.NewSelect().Model(existing).Where("name = ?", pm.Name).Scan(ctx)
	if err == nil {
		return errors.New("payement method already exists")
	}
	_, err = r.db.NewInsert().Model(pm).Exec(ctx)
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]PayementMethod, error) {
	var payementMethods []PayementMethod
	err := r.db.NewSelect().Model(&payementMethods).Scan(ctx)
	return payementMethods, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*PayementMethod, error) {
	pm := new(PayementMethod)
	err := r.db.NewSelect().Model(pm).Where("name = ?", name).Scan(ctx)
	return pm, err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model((*PayementMethod)(nil)).Where("name = ?", name).Exec(ctx)
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*PayementMethod, error) {
	pm := new(PayementMethod)
	err := r.db.NewSelect().Model(pm).Where("payement_method_id = ?", id).Scan(ctx)
	return pm, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model((*PayementMethod)(nil)).Where("payement_method_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) Update(ctx context.Context, pm *PayementMethod) error {
	existingPM := new(PayementMethod)
	err := r.db.NewSelect().Model(existingPM).Where("name = ?", pm.Name).Scan(ctx)
	if err == nil && existingPM.PayementMethodID != pm.PayementMethodID {
		return errors.New("payement method name already in use")
	}
	_, err = r.db.NewUpdate().Model(pm).Where("payement_method_id = ?", pm.PayementMethodID).Exec(ctx)
	return err
}
