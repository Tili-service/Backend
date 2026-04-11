package sale

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

func (r *Repository) Create(ctx context.Context, s *Sales) (*Sales, error) {
	_, err := r.db.NewInsert().Model(s).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Sales, error) {
	sale := &Sales{}
	err := r.db.NewSelect().
		Model(sale).
		Relation("PayementMethod").
		Where("v.sale_id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sale, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]*Sales, error) {
	var sales []*Sales
	err := r.db.NewSelect().
		Model(&sales).
		Relation("PayementMethod").
		OrderExpr("v.time_stamp DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *Repository) Update(ctx context.Context, s *Sales) (*Sales, error) {
	_, err := r.db.NewUpdate().Model(s).Where("sale_id = ?", s.Sale_ID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Sales{}).Where("sale_id = ?", id).Exec(ctx)
	return err
}
