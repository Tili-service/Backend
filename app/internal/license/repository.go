package license

import (
	"context"

	"tili/app/pkg/db"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func (r *Repository) CreateLicence(ctx context.Context, l *Licence) error {
	_, err := r.db.NewInsert().Model(l).Exec(ctx)
	return err
}

func (r *Repository) FindLicencesByAccountID(ctx context.Context, accountID int) ([]Licence, error) {
	var licences []Licence
	err := r.db.NewSelect().
		Model(&licences).
		Relation("Store").
		Where("l.account_id = ?", accountID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return licences, nil
}

func (r *Repository) DeleteLicencesByAccountID(ctx context.Context, accountID int) error {
	_, err := r.db.NewDelete().Model(&Licence{}).Where("account_id = ?", accountID).Exec(ctx)
	return err
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Licence, error) {
	l := &Licence{}
	err := r.db.NewSelect().Model(l).Where("licence_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model(&Licence{}).Where("licence_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) Update(ctx context.Context, l *Licence) (*Licence, error) {
	_, err := r.db.NewUpdate().Model(l).Where("licence_id = ?", l.LicenceID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return l, nil
}
