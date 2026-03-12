package license

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

func (r *Repository) CreateLicence(ctx context.Context, l *Licence) error {
	_, err := r.db.NewInsert().Model(l).Exec(ctx)
	return err
}

func (r *Repository) FindLicencesByAccountID(ctx context.Context, accountID int64) ([]Licence, error) {
	var licences []Licence
	err := r.db.NewSelect().Model(&licences).Where("account_id = ?", accountID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return licences, nil
}

func (r *Repository) DeleteLicencesByAccountID(ctx context.Context, accountID int64) error {
	_, err := r.db.NewDelete().Model(&Licence{}).Where("account_id = ?", accountID).Exec(ctx)
	return err
}
