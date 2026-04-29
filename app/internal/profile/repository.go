package profile

import (
	"context"

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

func (r *Repository) Create(ctx context.Context, p *Profile) error {
	_, err := r.db.NewInsert().Model(p).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Profile, error) {
	p := new(Profile)
	err := r.db.NewSelect().Model(p).Where("p.profile_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (r *Repository) FindByStoreAndPin(ctx context.Context, storeID int, pin string) (*Profile, error) {
	p := new(Profile)
	err := r.db.NewSelect().Model(p).
		Where("p.store_id = ?", storeID).
		Where("p.pin = ?", pin).
		Where("p.is_active = ?", true).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
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
	exists, err := r.db.NewSelect().Model(&Profile{}).
		Where("store_id = ?", storeID).
		Where("pin = ?", pin).
		Exists(ctx)
	if err != nil {
		return false, err
	}
	return exists, err
}

func (r *Repository) Update(ctx context.Context, p *Profile) error {
	_, err := r.db.NewUpdate().Model(p).Where("profile_id = ?", p.ProfileID).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "profile:")
	return err
}

func (r *Repository) FindAllProfilesByStoreID(ctx context.Context, storeID int) ([]*Profile, error) {
	var profiles []*Profile
	err := r.db.NewSelect().Model(&profiles).Where("store_id = ?", storeID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return profiles, err
}
