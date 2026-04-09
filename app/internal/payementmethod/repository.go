package payementmethod

import (
	"context"
	"database/sql"
	"errors"
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

func (r *Repository) Create(ctx context.Context, pm *PayementMethod) error {
	existing := new(PayementMethod)
	if hit, err := cache.Get(ctx, r.cache, fmt.Sprintf("payementmethod:name:%s", pm.Name), existing); err == nil && hit {
		return errors.New("payement method already exists")
	}
	err := r.db.NewSelect().Model(existing).Where("name = ?", pm.Name).Scan(ctx)
	if err == nil {
		return errors.New("payement method already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = r.db.NewInsert().Model(pm).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "payementmethod:")
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]PayementMethod, error) {
	const key = "payementmethod:all"
	var payementMethods []PayementMethod
	if hit, err := cache.Get(ctx, r.cache, key, &payementMethods); err == nil && hit {
		return payementMethods, nil
	}
	err := r.db.NewSelect().Model(&payementMethods).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &payementMethods, cache.DefaultTTL)
	return payementMethods, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*PayementMethod, error) {
	key := fmt.Sprintf("payementmethod:name:%s", name)
	pm := new(PayementMethod)
	if hit, err := cache.Get(ctx, r.cache, key, pm); err == nil && hit {
		return pm, nil
	}
	err := r.db.NewSelect().Model(pm).Where("name = ?", name).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, pm, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("payementmethod:id:%d", pm.PayementMethodID), pm, cache.DefaultTTL)
	return pm, err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model((*PayementMethod)(nil)).Where("name = ?", name).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "payementmethod:")
	return err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*PayementMethod, error) {
	key := fmt.Sprintf("payementmethod:id:%d", id)
	pm := new(PayementMethod)
	if hit, err := cache.Get(ctx, r.cache, key, pm); err == nil && hit {
		return pm, nil
	}
	err := r.db.NewSelect().Model(pm).Where("payement_method_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, pm, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("payementmethod:name:%s", pm.Name), pm, cache.DefaultTTL)
	return pm, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model((*PayementMethod)(nil)).Where("payement_method_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "payementmethod:")
	return err
}

func (r *Repository) Update(ctx context.Context, pm *PayementMethod) error {
	existingPM := new(PayementMethod)
	err := r.db.NewSelect().Model(existingPM).Where("name = ?", pm.Name).Scan(ctx)
	if err == nil && existingPM.PayementMethodID != pm.PayementMethodID {
		return errors.New("payement method name already in use")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	res, err := r.db.NewUpdate().Model(pm).Where("payement_method_id = ?", pm.PayementMethodID).Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	_ = cache.DeletePrefix(ctx, r.cache, "payementmethod:")
	return nil
}
