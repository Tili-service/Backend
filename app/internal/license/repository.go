package license

import (
	"context"
	"fmt"
	"time"

	"tili/app/internal/account"
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

const licencesByAccountTTL = 20 * time.Second

func NewRepository(d *db.Db, cacheClients ...*redis.Client) *Repository {
	var cacheClient *redis.Client
	if len(cacheClients) > 0 {
		cacheClient = cacheClients[0]
	}
	return &Repository{db: d.DB, cache: cacheClient}
}

func (r *Repository) CreateLicence(ctx context.Context, l *Licence) error {
	_, err := r.db.NewInsert().Model(l).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return err
}

func (r *Repository) FindLicencesByAccountID(ctx context.Context, accountID int) ([]Licence, error) {
	key := fmt.Sprintf("license:account:%d", accountID)
	var licences []Licence
	if hit, err := cache.Get(ctx, r.cache, key, &licences); err == nil && hit {
		return licences, nil
	}
	err := r.db.NewSelect().
		Model(&licences).
		Relation("Store").
		Where("l.account_id = ?", accountID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &licences, licencesByAccountTTL)
	return licences, nil
}

func (r *Repository) DeleteLicencesByAccountID(ctx context.Context, accountID int) error {
	_, err := r.db.NewDelete().Model(&Licence{}).Where("account_id = ?", accountID).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Licence, error) {
	l := &Licence{}
	err := r.db.NewSelect().Model(l).Where("licence_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *Repository) GetAccountByID(ctx context.Context, accountID int) (*account.Account, error) {
	account := &account.Account{}
	err := r.db.NewSelect().Model(account).Where("account_id = ?", accountID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *Repository) FindByTransaction(ctx context.Context, transaction string) (*Licence, error) {
	l := &Licence{}
	err := r.db.NewSelect().Model(l).Where("transaction = ?", transaction).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model(&Licence{}).Where("licence_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return err
}

func (r *Repository) Update(ctx context.Context, l *Licence) (*Licence, error) {
	_, err := r.db.NewUpdate().Model(l).Where("licence_id = ?", l.LicenceID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "license:")
	return l, nil
}
