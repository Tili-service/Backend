package sale

import (
	"context"
	"database/sql"
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

func (r *Repository) Create(ctx context.Context, s *Sale) (*Sale, error) {
	_, err := r.db.NewInsert().Model(s).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Sale, error) {
	sale := &Sale{}
	err := r.db.NewSelect().
		Model(sale).
		Relation("PayementMethod").
		Where("s.sale_id = ?", id).
		Where("s.is_deleted = ?", false).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSaleNotFound
	}
	if err != nil {
		return nil, err
	}
	return sale, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]*Sale, error) {
	var sales []*Sale
	err := r.db.NewSelect().
		Model(&sales).
		Relation("PayementMethod").
		Where("s.is_deleted = ?", false).
		OrderExpr("s.time_stamp DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *Repository) Update(ctx context.Context, s *Sale) (*Sale, error) {
	_, err := r.db.NewUpdate().Model(s).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.db.NewUpdate().Model(&Sale{}).Set("is_deleted = ?", true).Where("sale_id = ?", id).Where("is_deleted = ?", false).Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrSaleNotFound
	}
	return nil
}
