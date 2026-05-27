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
		total = total.Add(line.UnitPrice.Mul(qty))
	}
	return total.Round(2)
}

func snapshotLines(lines []SaleLine) []salehistory.SaleLineSnapshot {
	out := make([]salehistory.SaleLineSnapshot, len(lines))
	for i, line := range lines {
		out[i] = salehistory.SaleLineSnapshot{
			ItemID:    line.ItemID,
			Name:      line.Name,
			Quantity:  line.Quantity,
			UnitPrice: line.UnitPrice,
			TaxRate:   line.TaxRate,
		}
	}
	return out
}

// historyFromSale builds a history row reflecting the sale's state after the change.
func historyFromSale(s *Sale, changedByProfileID *int, changes map[string]any) *salehistory.SaleHistory {
	return &salehistory.SaleHistory{
		SaleID:             s.SaleID,
		ChangedByProfileID: changedByProfileID,
		Lines:              snapshotLines(s.Lines),
		Price:              s.Price,
		PayementMethodID:   s.PayementMethodID,
		TimeStamp:          s.TimeStamp,
		IsDeleted:          s.IsDeleted,
		Changes:            changes,
	}
}

func (s *Service) CreateSale(ctx context.Context, input CreateSaleInput) (*Sale, error) {
	total := computeTotal(input.Lines)
	if !total.IsPositive() {
		return nil, ErrInvalidSaleTotal
	}

	sale := &Sale{
		Lines:            input.Lines,
		Price:            total,
		TimeStamp:        time.Now(),
		PayementMethodID: input.PayementMethodID,
	}

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(sale).Exec(ctx); err != nil {
			return err
		}
		hist := historyFromSale(sale, nil, map[string]any{"action": "created"})
		return s.historyRepo.Insert(ctx, tx, hist)
	})
	if err != nil {
		return nil, err
	}
	return sale, nil
}

func (s *Service) GetSaleByID(ctx context.Context, id int) (*Sale, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetAllSales(ctx context.Context) ([]*Sale, error) {
	return s.repo.FindAll(ctx)
}

// diffLines returns a diff between old and new lines keyed by item_id.
// Shape: {"modified":[{"item_id":X,"quantity":{"from":a,"to":b}}], "added":[...], "removed":[...]}
func diffLines(oldLines, newLines []SaleLine) map[string]any {
	oldByItem := make(map[int]SaleLine, len(oldLines))
	for _, l := range oldLines {
		oldByItem[l.ItemID] = l
	}
	newByItem := make(map[int]SaleLine, len(newLines))
	for _, l := range newLines {
		newByItem[l.ItemID] = l
	}

	var modified []map[string]any
	var added []map[string]any
	var removed []map[string]any

	for itemID, newLine := range newByItem {
		if oldLine, ok := oldByItem[itemID]; ok {
			if oldLine.Quantity != newLine.Quantity {
				modified = append(modified, map[string]any{
					"item_id":  itemID,
					"name":     newLine.Name,
					"quantity": map[string]any{"from": oldLine.Quantity, "to": newLine.Quantity},
				})
			}
		} else {
			added = append(added, map[string]any{
				"item_id":  itemID,
				"name":     newLine.Name,
				"quantity": newLine.Quantity,
			})
		}
	}
	for itemID, oldLine := range oldByItem {
		if _, ok := newByItem[itemID]; !ok {
			removed = append(removed, map[string]any{
				"item_id":  itemID,
				"name":     oldLine.Name,
				"quantity": oldLine.Quantity,
			})
		}
	}

	if modified == nil && added == nil && removed == nil {
		return nil
	}
	out := map[string]any{}
	if modified != nil {
		out["modified"] = modified
	}
	if added != nil {
		out["added"] = added
	}
	if removed != nil {
		out["removed"] = removed
	}
	return out
}

// UpdateSale applies a partial update to a sale and records the post-update state
// alongside a diff of what changed in sale_history, within the same transaction.
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

		oldLines := existing.Lines
		oldPrice := existing.Price
		oldPayementMethodID := existing.PayementMethodID

		if input.Lines != nil {
			newLineMap := make(map[int]UpdateSaleLine)
			for _, l := range *input.Lines {
				newLineMap[l.ItemID] = l
			}

			var updatedLines []SaleLine
			for _, line := range existing.Lines {
				if update, ok := newLineMap[line.ItemID]; ok {
					delete(newLineMap, line.ItemID)
					if *update.Quantity == 0 {
						continue // quantity 0 = remove this item
					}
					line.Quantity = *update.Quantity
				}
				updatedLines = append(updatedLines, line)
			}

			// remaining entries in newLineMap are items not yet in the sale — add them
			for _, newItem := range newLineMap {
				if *newItem.Quantity == 0 {
					return ErrNewLineZeroQuantity
				}
				if newItem.Name == "" || newItem.UnitPrice == nil {
					return ErrNewLineIncomplete
				}
				line := SaleLine{
					ItemID:    newItem.ItemID,
					Name:      newItem.Name,
					Quantity:  *newItem.Quantity,
					UnitPrice: *newItem.UnitPrice,
				}
				if newItem.TaxRate != nil {
					line.TaxRate = *newItem.TaxRate
				}
				updatedLines = append(updatedLines, line)
			}

			if len(updatedLines) == 0 {
				return ErrSaleWouldBeEmpty
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

		changes := map[string]any{"action": "updated"}
		if lineDiff := diffLines(oldLines, existing.Lines); lineDiff != nil {
			changes["lines"] = lineDiff
		}
		if !oldPrice.Equal(existing.Price) {
			changes["price"] = map[string]any{"from": oldPrice, "to": existing.Price}
		}
		if oldPayementMethodID != existing.PayementMethodID {
			changes["payement_method_id"] = map[string]any{"from": oldPayementMethodID, "to": existing.PayementMethodID}
		}

		hist := historyFromSale(existing, changedByProfileID, changes)
		if err := s.historyRepo.Insert(ctx, tx, hist); err != nil {
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

		existing.IsDeleted = true
		changes := map[string]any{
			"action":     "deleted",
			"is_deleted": map[string]any{"from": false, "to": true},
		}
		hist := historyFromSale(existing, changedByProfileID, changes)
		return s.historyRepo.Insert(ctx, tx, hist)
	})
}
