package catalogue

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

// create
func (r *Repository) Create(ctx context.Context, c *Catalogue) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	return err
}

// find all
func (r *Repository) FindAll(ctx context.Context) ([]Catalogue, error) {
	var catalogues []Catalogue
	err := r.db.NewSelect().Model(&catalogues).Scan(ctx)
	return catalogues, err
}

// find by Id
func (r *Repository) FindByID(ctx context.Context, id int) (*Catalogue, error) {
	c := new(Catalogue)
	err := r.db.NewSelect().Model(c).Where("c.catalogue_id = ?", id).Scan(ctx)
	return c, err
}

// find by name
func (r *Repository) FindByName(ctx context.Context, name string) (*Catalogue, error) {
	c := new(Catalogue)
	err := r.db.NewSelect().Model(c).Where("c.name = ?", name).Scan(ctx)
	return c, err
}

// delete
func (r *Repository) DeleteByID(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&Catalogue{}).Where("catalogue_id = ?", id).Exec(ctx)
	return err
}

func (r *Repository) DeleteByName(ctx context.Context, name string) error {
	_, err := r.db.NewDelete().Model(&Catalogue{}).Where("name = ?", name).Exec(ctx)
	return err
}

// Update - Name or Descripton
func (r *Repository) Update(ctx context.Context, id int, input CatalogueUpdate) (*Catalogue, error) {
	catalogue := &Catalogue{}
	err := r.db.NewSelect().Model(catalogue).Where("c.catalogue_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		catalogue.Name = *input.Name
	}
	if input.Description != nil {
		catalogue.Description = *input.Description
	}
	_, err = r.db.NewUpdate().Model(catalogue).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return catalogue, nil
}
