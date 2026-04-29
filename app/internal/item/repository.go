package item

import (
	"context"
	"fmt"
	"strconv"

	"tili/app/pkg/cache"
	"tili/app/pkg/db"

	"github.com/redis/go-redis/v9"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Repository struct {
	db    *bun.DB
	cache *redis.Client
}

func NewRepository(d *db.Db, cacheClients ...*redis.Client) *Repository {
	var cacheClient *redis.Client
	if len(cacheClients) > 0 {
		cacheClient = cacheClients[0]
	}
	return &Repository{db: d.DB, cache: cacheClient}
}

func calcTaxAmount(price, tax decimal.Decimal) decimal.Decimal {
	one := decimal.NewFromInt(1)
	return price.Mul(tax).Div(one.Add(tax)).Round(2)
}

func (r *Repository) Create(ctx context.Context, i *Item) error {
	i.Tax_amount = calcTaxAmount(i.Price, i.Tax)
	_, err := r.db.NewInsert().Model(i).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "item:")
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]Item, error) {
	const key = "item:all"
	var items []Item
	if hit, err := cache.Get(ctx, r.cache, key, &items); err == nil && hit {
		return items, nil
	}
	err := r.db.NewSelect().Model(&items).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &items, cache.DefaultTTL)
	return items, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Item, error) {
	key := fmt.Sprintf("item:id:%d", id)
	i := new(Item)
	if hit, err := cache.Get(ctx, r.cache, key, i); err == nil && hit {
		return i, nil
	}
	err := r.db.NewSelect().Model(i).Where("i.item_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, i, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("item:name:%s", i.Name), i, cache.DefaultTTL)
	return i, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*Item, error) {
	key := fmt.Sprintf("item:name:%s", name)
	i := new(Item)
	if hit, err := cache.Get(ctx, r.cache, key, i); err == nil && hit {
		return i, nil
	}
	err := r.db.NewSelect().Model(i).Where("i.name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, i, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("item:id:%d", i.ItemId), i, cache.DefaultTTL)
	return i, err
}

func (r *Repository) FindByCategorieID(ctx context.Context, categorieID int) ([]Item, error) {
	key := fmt.Sprintf("item:categorie:%d", categorieID)
	var items []Item
	if hit, err := cache.Get(ctx, r.cache, key, &items); err == nil && hit {
		return items, nil
	}
	err := r.db.NewSelect().Model(&items).Where("i.categorie_id = ?", categorieID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &items, cache.DefaultTTL)
	return items, err
}

func (r *Repository) FindByPrice(ctx context.Context, price float64) ([]Item, error) {
	key := "item:price:" + strconv.FormatFloat(price, 'f', 2, 64)
	var items []Item
	if hit, err := cache.Get(ctx, r.cache, key, &items); err == nil && hit {
		return items, nil
	}
	err := r.db.NewSelect().Model(&items).Where("i.price = ?", price).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &items, cache.DefaultTTL)
	return items, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Item{}).Where("item_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "item:")
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model(&Item{}).Where("name = ?", name).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "item:")
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
	_ = cache.DeletePrefix(ctx, r.cache, "item:")
	return item, nil
}
