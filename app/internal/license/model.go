package license

import (
	"time"

	"tili/app/internal/store"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Licence struct {
	bun.BaseModel `bun:"table:licence,alias:l" swaggerignore:"true"`

	LicenceID   uuid.UUID    `bun:"licence_id,pk,type:uuid,default:gen_random_uuid()" json:"licence_id"`
	AccountID   int          `bun:"account_id,notnull"                                json:"account_id"`
	Expiration  time.Time    `bun:"expiration,notnull"                                json:"expiration"`
	Store       *store.Store `bun:"rel:has-one,join:licence_id=licence_id"            json:"store"`
	Transaction string       `bun:"transaction"                                       json:"-"`
	IsActive    bool         `bun:"is_active,default:true"                            json:"is_active"`
}

type CreateLicenceInput struct {
	DurationDays int    `json:"duration_days" binding:"required,min=1"`
	Transaction  string `json:"transaction"`
}

type CreatePaymentLinkInput struct {
	Offer string `json:"offer" binding:"required"`
}

type UpdateLicenceInput struct {
	Transaction *string `json:"transaction,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
