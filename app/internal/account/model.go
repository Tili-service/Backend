package account

import (
	"time"

	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:account,alias:a" swaggerignore:"true"`

	AccountID int64     `bun:"account_id,pk,autoincrement"          json:"account_id"`
	Email     string    `bun:"email,unique,notnull"                 json:"email"`
	Password  string    `bun:"password,notnull"                     json:"-"`
	Name      string    `bun:"name,notnull"                         json:"name"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
}

type RegistrationInput struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
