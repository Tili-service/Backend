package salehistory

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

// Insert writes a history row using the provided executor.
// Pass a bun.Tx to participate in an outer transaction, or the *bun.DB for a standalone insert.
func (r *Repository) Insert(ctx context.Context, exec bun.IDB, h *SaleHistory) error {
	_, err := exec.NewInsert().Model(h).Exec(ctx)
	return err
}

func (r *Repository) ListBySaleID(ctx context.Context, saleID int) ([]*SaleHistory, error) {
	var history []*SaleHistory
	err := r.db.NewSelect().
		Model(&history).
		Where("sh.sale_id = ?", saleID).
		OrderExpr("sh.changed_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return history, nil
}
