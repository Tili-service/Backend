package categorie

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

func (r *Repository) Create(ctx context.Context, c *Categorie) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]Categorie, error) {
	var categories []Categorie
	err := r.db.NewSelect().Model(&categories).Scan(ctx)
	return categories, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*Categorie, error) {
	c := new(Categorie)
	err := r.db.NewSelect().Model(c).Where("cat.categorie_id = ?", id).Scan(ctx)
	return c, err
}

func (r *Repository) FindByType(ctx context.Context, typ string) (*Categorie, error) {
	c := new(Categorie)
	err := r.db.NewSelect().Model(c).Where("cat.type = ?", typ).Scan(ctx)
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
	return cat, nil
}

func (r *Repository) DeleteById(ctx context.Context, id int) error {
	cat := &Categorie{}
	err := r.db.NewSelect().Model(cat).Where("cat.categorie_id = ?", id).Scan(ctx)
	if err != nil {
		return err
	}
	_, er := r.db.NewDelete().Model(cat).Where("cat.categorie_id = ?", id).Exec(ctx)
	return er
}

func (r *Repository) DeleteByType(ctx context.Context, typ string) error {
	cat := &Categorie{}
	err := r.db.NewSelect().Model(cat).Where("cat.type = ?", typ).Scan(ctx)
	if err != nil {
		return err
	}
	_, er := r.db.NewDelete().Model(cat).Where("cat.type = ?", typ).Exec(ctx)
	return er
}
