package catalog

import (
	"context"
	"fmt"

	"tili/app/pkg/cache"
	"tili/app/pkg/db"

	"github.com/redis/go-redis/v9"

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

func (r *Repository) Create(ctx context.Context, c *catalog) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%d:", c.StoreID))
	return err
}

func (r *Repository) FindAll(ctx context.Context, storeID int) ([]catalog, error) {
	key := fmt.Sprintf("catalog:store:%d:all", storeID)
	var catalogs []catalog
	if hit, err := cache.Get(ctx, r.cache, key, &catalogs); err == nil && hit {
		return catalogs, nil
	}
	err := r.db.NewSelect().Model(&catalogs).Where("c.store_id = ?", storeID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &catalogs, cache.DefaultTTL)
	return catalogs, err
}

func (r *Repository) FindByID(ctx context.Context, id int, storeID int) (*catalog, error) {
	c := new(catalog)
	err := r.db.NewSelect().Model(c).
		Where("c.catalog_id = ?", id).
		Where("c.store_id = ?", storeID).
		Scan(ctx)
	return c, err
}

func (r *Repository) FindByName(ctx context.Context, name string, storeID int) (*catalog, error) {
	c := new(catalog)
	err := r.db.NewSelect().Model(c).
		Where("c.name = ?", name).
		Where("c.store_id = ?", storeID).
		Scan(ctx)
	return c, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int, storeID int) error {
	_, err := r.db.NewDelete().Model(&catalog{}).
		Where("catalog_id = ?", id).
		Where("store_id = ?", storeID).
		Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%d:", storeID))
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string, storeID int) error {
	_, err := r.db.NewDelete().Model(&catalog{}).
		Where("name = ?", name).
		Where("store_id = ?", storeID).
		Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%d:", storeID))
	return err
}

func (r *Repository) Update(ctx context.Context, id int, storeID int, input catalogUpdate) (*catalog, error) {
	c := &catalog{}
	err := r.db.NewSelect().Model(c).
		Where("c.catalog_id = ?", id).
		Where("c.store_id = ?", storeID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		c.Name = *input.Name
	}
	if input.Description != nil {
		c.Description = *input.Description
	}
	_, err = r.db.NewUpdate().Model(c).
		Where("catalog_id = ?", id).
		Where("store_id = ?", storeID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%d:", storeID))
	return c, nil
}
