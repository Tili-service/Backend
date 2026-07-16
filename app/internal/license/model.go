package license

import (
	"time"

	"tili/app/internal/store"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Licence struct {
	bun.BaseModel `bun:"table:licence,alias:l" swaggerignore:"true"`

	LicenceID   uuid.UUID          `bun:"licence_id,pk,type:uuid,default:gen_random_uuid()" json:"licence_id"`
	AccountID   uuid.UUID          `bun:"account_id,notnull,type:uuid"                     json:"account_id"`
	Expiration  time.Time          `bun:"expiration,notnull"                                json:"expiration"`
	NextPayment time.Time          `bun:"next_payment,notnull"                              json:"next_payment"`
	Store       *store.Store       `bun:"rel:has-one,join:licence_id=licence_id"            json:"store"`
	Transaction string             `bun:"transaction"                                       json:"-"`
	IsActive    bool               `bun:"is_active,default:true"                            json:"is_active"`
	Stripe      *StripeLicenceInfo `bun:"-" swaggerignore:"true"                           json:"stripe,omitempty"`
}

type StripeLicenceInfo struct {
	CheckoutSessionID string     `json:"checkout_session_id,omitempty"`
	SubscriptionID    string     `json:"subscription_id,omitempty"`
	Status            string     `json:"status,omitempty"`
	NextPaymentAt     *time.Time `json:"next_payment_at,omitempty"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end,omitempty"`
	PriceAmount       int64      `json:"price_amount,omitempty"`
	PriceCurrency     string     `json:"price_currency,omitempty"`
	PriceInterval     string     `json:"price_interval,omitempty"`
	PriceProductID    string     `json:"price_product_id,omitempty"`
	PriceProductName  string     `json:"price_product_name,omitempty"`
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
