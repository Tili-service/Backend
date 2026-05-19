package sale

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"tili/app/internal/salehistory"
	"tili/app/pkg/db"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

var ErrInvalidSaleTotal = errors.New("sale total must be positive")

type Service struct {
	db          *bun.DB
	repo        *Repository
	historyRepo *salehistory.Repository
}

func NewService(d *db.Db, repo *Repository, historyRepo *salehistory.Repository) *Service {
	return &Service{db: d.DB, repo: repo, historyRepo: historyRepo}
}

func computeTotal(lines []SaleLine) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		qty := decimal.NewFromInt(int64(line.Quantity))
		fmt.Println("Line:", line.Name, "Qty:", qty, "UnitPrice:", line.UnitPrice)
		total = total.Add(line.UnitPrice.Mul(qty))
	}
	return total.Round(2)
}

func (s *Service) CreateSale(ctx context.Context, input CreateSaleInput) (*Sale, error) {
	total := computeTotal(input.Lines)
	fmt.Println("Computed total:", total)
	if !total.IsPositive() {
		return nil, ErrInvalidSaleTotal
	}

	sale := &Sale{
		Lines:            input.Lines,
		Price:            total,
		TimeStamp:        time.Now(),
		PayementMethodID: input.PayementMethodID,
	}
	return s.repo.Create(ctx, sale)
}

func (s *Service) GetSaleByID(ctx context.Context, id int) (*Sale, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetAllSales(ctx context.Context) ([]*Sale, error) {
	return s.repo.FindAll(ctx)
}

// UpdateSale applies a partial update to a sale, recording the prior state in
// sale_history within the same transaction so the original sale is never lost.
// changedByProfileID is optional and identifies the profile that authored the change.
func (s *Service) UpdateSale(ctx context.Context, id int, input UpdateSaleInput, changedByProfileID *int) (*Sale, error) {
	var result *Sale

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		existing := &Sale{}
		err := tx.NewSelect().
			Model(existing).
			Where("s.sale_id = ?", id).
			Scan(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSaleNotFound
		}
		if err != nil {
			return err
		}

		previousLines := make([]salehistory.SaleLineSnapshot, len(existing.Lines))
		for i, line := range existing.Lines {
			previousLines[i] = salehistory.SaleLineSnapshot{
				ItemID:    line.ItemID,
				Name:      line.Name,
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
				TaxRate:   line.TaxRate,
			}
		}
		hist := &salehistory.SaleHistory{
			SaleID:                   existing.SaleID,
			ChangedByProfileID:       changedByProfileID,
			PreviousLines:            previousLines,
			PreviousPrice:            existing.Price,
			PreviousPayementMethodID: existing.PayementMethodID,
			PreviousTimeStamp:        existing.TimeStamp,
		}
		if err := s.historyRepo.Insert(ctx, tx, hist); err != nil {
			return err
		}

		if input.Lines != nil {
			newLineMap := make(map[int]int)
			for _, l := range *input.Lines {
				newLineMap[l.ItemID] = l.Quantity
			}

			var updatedLines []SaleLine
			for _, oldLine := range existing.Lines {
				if qty, ok := newLineMap[oldLine.ItemID]; ok {
					oldLine.Quantity = qty
					updatedLines = append(updatedLines, oldLine)
				}
			}

			total := computeTotal(updatedLines)
			if !total.IsPositive() {
				return ErrInvalidSaleTotal
			}
			existing.Lines = updatedLines
			existing.Price = total
		}
		if input.PayementMethodID != nil {
			existing.PayementMethodID = *input.PayementMethodID
		}

		if _, err := tx.NewUpdate().Model(existing).WherePK().Exec(ctx); err != nil {
			return err
		}

		result = existing
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) DeleteSale(ctx context.Context, id int, changedByProfileID *int) error {
	return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		existing := &Sale{}
		err := tx.NewSelect().
			Model(existing).
			Where("s.sale_id = ?", id).
			Where("s.is_deleted = ?", false).
			Scan(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSaleNotFound
		}
		if err != nil {
			return err
		}

		previousLines := make([]salehistory.SaleLineSnapshot, len(existing.Lines))
		for i, line := range existing.Lines {
			previousLines[i] = salehistory.SaleLineSnapshot{
				ItemID:    line.ItemID,
				Name:      line.Name,
				Quantity:  line.Quantity,
				UnitPrice: line.UnitPrice,
				TaxRate:   line.TaxRate,
			}
		}
		hist := &salehistory.SaleHistory{
			SaleID:                   existing.SaleID,
			ChangedByProfileID:       changedByProfileID,
			PreviousLines:            previousLines,
			PreviousPrice:            existing.Price,
			PreviousPayementMethodID: existing.PayementMethodID,
			PreviousTimeStamp:        existing.TimeStamp,
		}
		if err := s.historyRepo.Insert(ctx, tx, hist); err != nil {
			return err
		}

		res, err := tx.NewUpdate().Model(&Sale{}).Set("is_deleted = ?", true).Where("sale_id = ?", id).Where("is_deleted = ?", false).Exec(ctx)
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
	})
}
