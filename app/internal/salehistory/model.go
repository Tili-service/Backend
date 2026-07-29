package salehistory

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type SaleLineSnapshot struct {
	ItemID    uuid.UUID       `json:"item_id"`
	Name      string          `json:"name"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	TaxRate   decimal.Decimal `json:"tax_rate,omitempty"`
}

type SalePaymentSnapshot struct {
	PayementMethodID uuid.UUID       `json:"payement_method_id"`
	Amount           decimal.Decimal `json:"amount"`
}

type SaleHistory struct {
	bun.BaseModel `bun:"table:sale_history,alias:sh" swaggerignore:"true"`

	HistoryID          uuid.UUID             `bun:"history_id,pk,type:uuid,default:gen_random_uuid()" json:"history_id"`
	SaleID             uuid.UUID             `bun:"sale_id,notnull,type:uuid"                      json:"sale_id"`
	ChangedAt          time.Time             `bun:"changed_at,default:current_timestamp" json:"changed_at"`
	ChangedByProfileID *uuid.UUID            `bun:"changed_by_profile_id,type:uuid"                json:"changed_by_profile_id,omitempty"`
	Lines              []SaleLineSnapshot    `bun:"lines,type:jsonb"                     json:"lines"`
	Payments           []SalePaymentSnapshot `bun:"payments,type:jsonb"                  json:"payments"`
	Price              decimal.Decimal       `bun:"price,type:decimal(10,2)"             json:"price"`
	TimeStamp          time.Time             `bun:"time_stamp"                           json:"time_stamp"`
	IsDeleted          bool                  `bun:"is_deleted,notnull,default:false"     json:"is_deleted"`
	Changes            map[string]any        `bun:"changes,type:jsonb"                   json:"changes,omitempty"`
}
