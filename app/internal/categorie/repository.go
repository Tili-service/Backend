package categorie

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

func (r *Repository) Create(ctx context.Context, c *Categorie) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "categorie:")
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]Categorie, error) {
	const key = "categorie:all"
	var categories []Categorie
	if hit, err := cache.Get(ctx, r.cache, key, &categories); err == nil && hit {
		return categories, nil
	}
	err := r.db.NewSelect().Model(&categories).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, &categories, cache.DefaultTTL)
	return categories, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Categorie, error) {
	key := fmt.Sprintf("categorie:id:%d", id)
	c := new(Categorie)
	if hit, err := cache.Get(ctx, r.cache, key, c); err == nil && hit {
		return c, nil
	}
	err := r.db.NewSelect().Model(c).Where("cat.categorie_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, c, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("categorie:type:%s", c.Type), c, cache.DefaultTTL)
	return c, err
}

func (r *Repository) FindByType(ctx context.Context, typ string) (*Categorie, error) {
	key := fmt.Sprintf("categorie:type:%s", typ)
	c := new(Categorie)
	if hit, err := cache.Get(ctx, r.cache, key, c); err == nil && hit {
		return c, nil
	}
	err := r.db.NewSelect().Model(c).Where("cat.type = ?", typ).Scan(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.Set(ctx, r.cache, key, c, cache.DefaultTTL)
	_ = cache.Set(ctx, r.cache, fmt.Sprintf("categorie:id:%d", c.CategorieID), c, cache.DefaultTTL)
	return c, err
}

func (r *Repository) Update(ctx context.Context, id int, c *Categorie) (*Categorie, error) {
	cat := &Categorie{}
	err := r.db.NewSelect().Model(cat).Where("cat.categorie_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	cat.Type = c.Type
	_, err = r.db.NewUpdate().Model(cat).Where("cat.categorie_id = ?", id).Exec(ctx)
	if err != nil {
		return nil, err
	}
	_ = cache.DeletePrefix(ctx, r.cache, "categorie:")
	return cat, nil
}

func (r *Repository) DeleteById(ctx context.Context, id int) error {
	cat := &Categorie{}
	err := r.db.NewSelect().Model(cat).Where("cat.categorie_id = ?", id).Scan(ctx)
	if err != nil {
		return err
	}
	_, er := r.db.NewDelete().Model(cat).Where("cat.categorie_id = ?", id).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "categorie:")
	return er
}

func (r *Repository) DeleteByType(ctx context.Context, typ string) error {
	cat := &Categorie{}
	err := r.db.NewSelect().Model(cat).Where("cat.type = ?", typ).Scan(ctx)
	if err != nil {
		return err
	}
	_, er := r.db.NewDelete().Model(cat).Where("cat.type = ?", typ).Exec(ctx)
	_ = cache.DeletePrefix(ctx, r.cache, "categorie:")
	return er
}
