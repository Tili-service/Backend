package profile

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

func (r *Repository) Create(ctx context.Context, p *Profile) error {
	_, err := r.db.NewInsert().Model(p).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Profile, error) {
	key := fmt.Sprintf("profile:id:%d", id)
	p := new(Profile)
	if hit, err := cache.Get(ctx, r.cache, key, p); err == nil && hit {
		return p, nil
	}
	err := r.db.NewSelect().Model(p).Where("p.profile_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, p, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("profile:store:%d", p.StoreID), p, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("profile:store-pin:%d:%s", p.StoreID, p.Pin), p, cache.DefaultTTL)
	return p, err
}

func (r *Repository) FindByStoreAndPin(ctx context.Context, storeID int, pin string) (*Profile, error) {
	key := fmt.Sprintf("profile:store-pin:%d:%s", storeID, pin)
	p := new(Profile)
	if hit, err := cache.Get(ctx, r.cache, key, p); err == nil && hit {
		return p, nil
	}
	err := r.db.NewSelect().Model(p).
		Where("p.store_id = ?", storeID).
		Where("p.pin = ?", pin).
		Where("p.is_active = ?", true).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, p, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("profile:id:%d", p.ProfileID), p, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("profile:store:%d", p.StoreID), p, cache.DefaultTTL)
	return p, err
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Profile{}).Where("profile_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) DeleteByStoreID(ctx context.Context, storeID int) error {
	_, err := r.db.NewDelete().Model(&Profile{}).Where("store_id = ?", storeID).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) PinExistsInStore(ctx context.Context, storeID int, pin string) (bool, error) {
	key := fmt.Sprintf("profile:pin-exists:%d:%s", storeID, pin)
	var exists bool
	if hit, err := cache.Get(ctx, r.cache, key, &exists); err == nil && hit {
		return exists, nil
	}
	exists, err := r.db.NewSelect().Model(&Profile{}).
		Where("store_id = ?", storeID).
		Where("pin = ?", pin).
		Exists(ctx)
	if err != nil {
		return false, err
	}
	_ = cache.Set(ctx, r.cache, key, &exists, cache.DefaultTTL)
	return exists, err
}

func (r *Repository) Update(ctx context.Context, p *Profile) error {
	_, err := r.db.NewUpdate().Model(p).Where("profile_id = ?", p.ProfileID).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) FindAllProfilesByStoreID(ctx context.Context, storeID int) ([]*Profile, error) {
	key := fmt.Sprintf("profile:store:%d", storeID)
	var profiles []*Profile
	if hit, err := cache.Get(ctx, r.cache, key, &profiles); err == nil && hit {
		return profiles, nil
	}
	err := r.db.NewSelect().Model(&profiles).Where("store_id = ?", storeID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &profiles, cache.DefaultTTL)
	return profiles, err
}
