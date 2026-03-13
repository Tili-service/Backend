package license

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Licence struct {
	bun.BaseModel `bun:"table:licence,alias:l" swaggerignore:"true"`

	LicenceID   uuid.UUID `bun:"licence_id,pk,type:uuid,default:gen_random_uuid()" json:"licence_id"`
	AccountID   int       `bun:"account_id,notnull"                                json:"account_id"`
	Expiration  time.Time `bun:"expiration,notnull"                                json:"expiration"`
	Transaction string    `bun:"transaction"                                       json:"transaction,omitempty"`
	IsActive    bool      `bun:"is_active,default:true"                            json:"is_active"`
}

type CreateLicenceInput struct {
	DurationDays int    `json:"duration_days" binding:"required,min=1"`
	Transaction  string `json:"transaction"`
}

type CreatePaymentLinkInput struct {
	Offer string `json:"offer"`
}
