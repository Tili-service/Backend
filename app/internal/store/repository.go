package store

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

func (r *Repository) Create(ctx context.Context, s *Store) (*Store, error) {
	_, err := r.db.NewInsert().Model(s).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "store:")
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return s, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Store, error) {
	key := fmt.Sprintf("store:id:%s", id)
	store := &Store{}
	if hit, err := cache.Get(ctx, r.cache, key, store); err == nil && hit {
		return store, nil
	}
	err := r.db.NewSelect().Model(store).Where("store_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, store, cache.DefaultTTL)
	return store, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]*Store, error) {
	const key = "store:all"
	var stores []*Store
	if hit, err := cache.Get(ctx, r.cache, key, &stores); err == nil && hit {
		return stores, nil
	}
	err := r.db.NewSelect().Model(&stores).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &stores, cache.DefaultTTL)
	return stores, nil
}

func (r *Repository) FindByBuyerID(ctx context.Context, buyerID uuid.UUID) ([]Store, error) {
	key := fmt.Sprintf("store:buyer:%s", buyerID)
	var stores []Store
	if hit, err := cache.Get(ctx, r.cache, key, &stores); err == nil && hit {
		return stores, nil
	}
	err := r.db.NewSelect().Model(&stores).Where("buyer_id = ?", buyerID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &stores, cache.DefaultTTL)
	return stores, nil
}

func (r *Repository) FindByLicenceID(ctx context.Context, licenceID uuid.UUID) (*Store, error) {
	store := &Store{}
	err := r.db.NewSelect().Model(store).Where("licence_id = ?", licenceID).Scan(ctx)
	return store, err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model(&Store{}).Where("store_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "store:")
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return err
}

func (r *Repository) Update(ctx context.Context, s *Store) (*Store, error) {
	_, err := r.db.NewUpdate().Model(s).Where("store_id = ?", s.StoreID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "store:")
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return s, nil
}

type StoreProfileData struct {
	BuyerID uuid.UUID
}

func (r *Repository) FindByIDForProfile(ctx context.Context, id uuid.UUID) (*StoreProfileData, error) {
	store := &Store{}
	err := r.db.NewSelect().Model(store).Where("store_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &StoreProfileData{BuyerID: store.BuyerID}, nil
}
