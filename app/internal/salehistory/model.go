package salehistory

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type SaleLineSnapshot struct {
	ItemID    int             `json:"item_id"`
	Name      string          `json:"name"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	TaxRate   decimal.Decimal `json:"tax_rate,omitempty"`
}

type SaleHistory struct {
	bun.BaseModel `bun:"table:sale_history,alias:sh" swaggerignore:"true"`

	HistoryID                int                `bun:"history_id,pk,autoincrement"          json:"history_id"`
	SaleID                   int                `bun:"sale_id,notnull"                      json:"sale_id"`
	ChangedAt                time.Time          `bun:"changed_at,default:current_timestamp" json:"changed_at"`
	ChangedByAccountID       *int               `bun:"changed_by_account_id"                json:"changed_by_account_id,omitempty"`
	PreviousLines            []SaleLineSnapshot `bun:"previous_lines,type:jsonb"         json:"previous_lines"`
	PreviousPrice            decimal.Decimal    `bun:"previous_price,type:decimal(10,2)"    json:"previous_price"`
	PreviousPayementMethodID int                `bun:"previous_payement_method_id"          json:"previous_payement_method_id"`
	PreviousTimeStamp        time.Time          `bun:"previous_time_stamp"                  json:"previous_time_stamp"`
}
