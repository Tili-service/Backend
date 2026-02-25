package account

import (
	"time"

	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:account,alias:a" swaggerignore:"true"`

	AccountID int64     `bun:"account_id,pk,autoincrement" json:"account_id"`
	LicenceID int64     `bun:"licence_id,notnull"          json:"licence_id"`
	ExpiresAt time.Time `bun:"expires_at,notnull"          json:"expires_at"`
	IsActive  bool      `bun:"is_active,default:true"      json:"is_active"`
}
