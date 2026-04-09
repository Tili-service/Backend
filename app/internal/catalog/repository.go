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

func NewRepository(d *db.Db, cacheClient *redis.Client) *Repository {
	return &Repository{db: d.DB, cache: cacheClient}
}

func (r *Repository) Create(ctx context.Context, c *catalog) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "catalog:")
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]catalog, error) {
	const key = "catalog:all"
	var catalogs []catalog
	if hit, err := cache.Get(ctx, r.cache, key, &catalogs); err == nil && hit {
		return catalogs, nil
	}
	err := r.db.NewSelect().Model(&catalogs).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &catalogs, cache.DefaultTTL)
	return catalogs, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*catalog, error) {
	key := fmt.Sprintf("catalog:id:%d", id)
	c := new(catalog)
	if hit, err := cache.Get(ctx, r.cache, key, c); err == nil && hit {
		return c, nil
	}
	err := r.db.NewSelect().Model(c).Where("c.catalog_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, c, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("catalog:name:%s", c.Name), c, cache.DefaultTTL)
	return c, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*catalog, error) {
	key := fmt.Sprintf("catalog:name:%s", name)
	c := new(catalog)
	if hit, err := cache.Get(ctx, r.cache, key, c); err == nil && hit {
		return c, nil
	}
	err := r.db.NewSelect().Model(c).Where("c.name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, c, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("catalog:id:%d", c.CatalogID), c, cache.DefaultTTL)
	return c, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&catalog{}).Where("catalog_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "catalog:")
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model(&catalog{}).Where("name = ?", name).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "catalog:")
	return err
}

func (r *Repository) Update(ctx context.Context, id int, input catalogUpdate) (*catalog, error) {
	catalog := &catalog{}
	err := r.db.NewSelect().Model(catalog).Where("c.catalog_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		catalog.Name = *input.Name
	}
	if input.Description != nil {
		catalog.Description = *input.Description
	}
	_, err = r.db.NewUpdate().Model(catalog).Where("catalog_id = ?", id).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "catalog:")
	return catalog, nil
}
