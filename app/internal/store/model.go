package store

import (
	"tili/app/internal/account"

	"github.com/uptrace/bun"
)

type Store struct {
	bun.BaseModel `bun:"table:store,alias:s" swaggerignore:"true"`

	StoreID   int64            `bun:"store_id,pk,autoincrement" json:"store_id"`
	StoreName string           `bun:"store_name,notnull"        json:"store_name"`
	AccountID int64            `bun:"account_id"                json:"account_id"`
	Account   *account.Account `bun:"rel:belongs-to,join:account_id=account_id" json:"account,omitempty"`
}

type CreateStoreInput struct {
	StoreName string `json:"store_name" binding:"required"`
	AccountID int64  `json:"account_id" binding:"required"`
}

type UpdateStoreInput struct {
	StoreName string `json:"store_name" binding:"required"`
}
