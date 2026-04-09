package account

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

func (r *Repository) Create(ctx context.Context, a *Account) error {
	_, err := r.db.NewInsert().Model(a).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "account:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Account, error) {
	key := fmt.Sprintf("account:id:%d", id)
	a := &Account{}
	if hit, err := cache.Get(ctx, r.cache, key, a); err == nil && hit {
		return a, nil
	}
	err := r.db.NewSelect().Model(a).Where("account_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, a, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("account:email:%s", a.Email), a, cache.DefaultTTL)
	return a, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	key := fmt.Sprintf("account:email:%s", email)
	a := &Account{}
	if hit, err := cache.Get(ctx, r.cache, key, a); err == nil && hit {
		return a, nil
	}
	err := r.db.NewSelect().Model(a).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, a, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("account:id:%d", a.AccountID), a, cache.DefaultTTL)
	return a, nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Account{}).Where("account_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "account:")
	return err
}

func (r *Repository) Update(ctx context.Context, a *Account) (*Account, error) {
	_, err := r.db.NewUpdate().Model(a).Where("account_id = ?", a.AccountID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "account:")
	return a, nil
}
