package user

import (
	"tili/backend/pkg/db"
)

type Repository struct {
}

func NewRepository(db *db.Db) *Repository {
	return &Repository{}
}