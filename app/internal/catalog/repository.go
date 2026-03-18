package catalog

import (
	"context"

	"tili/app/pkg/db"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(d *db.Db) *Repository {
	return &Repository{db: d.DB}
}

func (r *Repository) Create(ctx context.Context, c *catalog) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]catalog, error) {
	var catalogs []catalog
	err := r.db.NewSelect().Model(&catalogs).Scan(ctx)
	return catalogs, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*catalog, error) {
	c := new(catalog)
	err := r.db.NewSelect().Model(c).Where("c.catalog_id = ?", id).Scan(ctx)
	return c, err
}

func (r *Repository) FindByName(ctx context.Context, name string) (*catalog, error) {
	c := new(catalog)
	err := r.db.NewSelect().Model(c).Where("c.name = ?", name).Scan(ctx)
	return c, err
}

func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&catalog{}).Where("catalog_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model(&catalog{}).Where("name = ?", name).Exec(ctx)
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
	return catalog, nil
}
