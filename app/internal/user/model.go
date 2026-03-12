package user

import (
	"tili/app/internal/store"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:user,alias:u" swaggerignore:"true"`

	UserID       int          `bun:"user_id,pk,autoincrement" json:"user_id"`
	StoreID      int          `bun:"store_id"                 json:"store_id"`
	Name         string       `bun:"name,notnull"             json:"name"`
	Email        string       `bun:"email,unique,notnull"     json:"email"`
	PasswordHash string       `bun:"password,notnull"    json:"-"`
	AccessCode   string       `bun:"access_code"              json:"access_code,omitempty"`
	AccessLevel  int          `bun:"access_level"             json:"access_level"`
	Store        *store.Store `bun:"rel:belongs-to,join:store_id=store_id" json:"store,omitempty"`
}

type CreateUserInput struct {
	StoreID     int    `json:"store_id"`
	Name        string `json:"name"          binding:"required"`
	Email       string `json:"email"         binding:"required,email"`
	Password    string `json:"password"      binding:"required,min=6"`
	AccessCode  string `json:"access_code"`
	AccessLevel int    `json:"access_level"`
}

type UpdateUserInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"        binding:"omitempty,email"`
	AccessCode  string `json:"access_code"`
	AccessLevel int    `json:"access_level"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
