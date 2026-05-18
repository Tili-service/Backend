package item

import (
	"context"

	"tili/app/pkg/db"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func calcTaxAmount(price, tax decimal.Decimal) decimal.Decimal {
	one := decimal.NewFromInt(1)
	return price.Mul(tax).Div(one.Add(tax)).Round(2)
}

func (r *Repository) Create(ctx context.Context, i *Item) error {
	i.Tax_amount = calcTaxAmount(i.Price, i.Tax)
	_, err := r.db.NewInsert().Model(i).Exec(ctx)
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]Item, error) {
	var items []Item
	err := r.db.NewSelect().Model(&items).Scan(ctx)
	return items, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Item, error) {
	i := new(Item)
	err := r.db.NewSelect().Model(i).Where("i.item_id = ?", id).Scan(ctx)
	return i, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*Item, error) {
	i := new(Item)
	err := r.db.NewSelect().Model(i).Where("i.name = ?", name).Scan(ctx)
	return i, err
}

func (r *Repository) FindByCategorieID(ctx context.Context, categorieID int) ([]Item, error) {
	var items []Item
	err := r.db.NewSelect().Model(&items).Where("i.categorie_id = ?", categorieID).Scan(ctx)
	return items, err
}

func (r *Repository) FindByPrice(ctx context.Context, price float64) ([]Item, error) {
	var items []Item
	err := r.db.NewSelect().Model(&items).Where("i.price = ?", price).Scan(ctx)
	return items, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Item{}).Where("item_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model(&Item{}).Where("name = ?", name).Exec(ctx)
	return err
}

func (r *Repository) Update(ctx context.Context, id int, input ItemUpdate) (*Item, error) {
	item := &Item{}
	err := r.db.NewSelect().Model(item).Where("i.item_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	if input.Name != nil && *input.Name != "" {
		item.Name = *input.Name
	}
	if input.Price != nil && input.Price.IsPositive() {
		item.Price = *input.Price
	}
	if input.Tax != nil && input.Tax.IsPositive() {
		item.Tax = *input.Tax
	}
	if input.CategorieID != nil && *input.CategorieID > 0 {
		item.CategorieID = *input.CategorieID
	}
	item.Tax_amount = calcTaxAmount(item.Price, item.Tax)
	_, err = r.db.NewUpdate().Model(item).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return item, nil
}
