package account

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

func (r *Repository) Create(ctx context.Context, a *Account) error {
	_, err := r.db.NewInsert().Model(a).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "account:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Account, error) {
	a := &Account{}
	err := r.db.NewSelect().Model(a).Where("account_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*Account, error) {
	a := &Account{}
	err := r.db.NewSelect().Model(a).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
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
