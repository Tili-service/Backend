package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:account,alias:a" swaggerignore:"true"`

	AccountID        uuid.UUID `bun:"account_id,pk,type:uuid,default:gen_random_uuid()" json:"account_id"`
	Email            string    `bun:"email,unique,notnull"                 json:"email"`
	Password         string    `bun:"password,notnull"                     json:"-"`
	Name             string    `bun:"name,notnull"                         json:"name"`
	StripeCustomerID string    `bun:"stripe_customer_id"                   json:"-"`
	CreatedAt        time.Time `bun:"created_at,default:current_timestamp" json:"created_at"`
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

type UpdateAccountInput struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty" binding:"required,email"`
}

type ResetPasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
