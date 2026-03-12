package user

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

func (r *Repository) Create(ctx context.Context, u *User) error {
	_, err := r.db.NewInsert().Model(u).Exec(ctx)
	return err
}

func (r *Repository) FindAll(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.NewSelect().Model(&users).Scan(ctx)
	return users, err
}

func (r *Repository) FindByID(ctx context.Context, id int) (*User, error) {
	u := new(User)
	err := r.db.NewSelect().Model(u).Where("u.user_id = ?", id).Scan(ctx)
	return u, err
}

func (r *Repository) FindByStoreID(ctx context.Context, storeID int) (*User, error) {
	u := new(User)
	err := r.db.NewSelect().Model(u).Where("u.store_id = ?", storeID).Scan(ctx)
	return u, err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u := new(User)
	err := r.db.NewSelect().Model(u).Where("u.email = ?", email).Scan(ctx)
	return u, err
}

func (r *Repository) Update(ctx context.Context, u *User) error {
	_, err := r.db.NewUpdate().Model(u).WherePK().Exec(ctx)
	return err
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model(&User{}).Where("user_id = ?", id).Exec(ctx)
	return err
}
