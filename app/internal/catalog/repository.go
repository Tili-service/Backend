package catalog

import (
	"context"
	"fmt"

	"tili/app/pkg/cache"
	"tili/app/pkg/db"

	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
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

func (r *Repository) Create(ctx context.Context, c *Catalog) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%s:", c.StoreID))
	return err
}

func (r *Repository) FindAll(ctx context.Context, storeID uuid.UUID) ([]Catalog, error) {
	key := fmt.Sprintf("catalog:store:%s:all", storeID)
	var catalogs []Catalog
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

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID, storeID uuid.UUID) (*Catalog, error) {
	c := new(Catalog)
	err := r.db.NewSelect().Model(c).
		Where("c.catalog_id = ?", id).
		Where("c.store_id = ?", storeID).
		Scan(ctx)
	return c, err
}

func (r *Repository) FindByName(ctx context.Context, name string, storeID uuid.UUID) (*Catalog, error) {
	c := new(Catalog)
	err := r.db.NewSelect().Model(c).
		Where("c.name = ?", name).
		Where("c.store_id = ?", storeID).
		Scan(ctx)
	return c, err
}

func (r *Repository) DeleteByID(ctx context.Context, id uuid.UUID, storeID uuid.UUID) error {
	_, err := r.db.NewDelete().Model(&Catalog{}).
		Where("catalog_id = ?", id).
		Where("store_id = ?", storeID).
		Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%s:", storeID))
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string, storeID uuid.UUID) error {
	_, err := r.db.NewDelete().Model(&Catalog{}).
		Where("name = ?", name).
		Where("store_id = ?", storeID).
		Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%s:", storeID))
	return err
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, storeID uuid.UUID, input catalogUpdate) (*Catalog, error) {
	c := &Catalog{}
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
	_ = cache.DeletePrefix(ctx, r.cache, fmt.Sprintf("catalog:store:%s:", storeID))
	return c, nil
}
