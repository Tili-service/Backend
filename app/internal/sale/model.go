package sale

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

var ErrSaleNotFound = errors.New("sale not found")

type SaleLine struct {
	ItemID    int             `json:"item_id"    binding:"required"`
	Name      string          `json:"name"`
	Quantity  int             `json:"quantity"   binding:"required,min=1"`
	UnitPrice decimal.Decimal `json:"unit_price" binding:"required"`
	TaxRate   decimal.Decimal `json:"tax_rate,omitempty"`
}

type SalePayment struct {
	PayementMethodID int             `json:"payement_method_id" binding:"required"`
	Amount           decimal.Decimal `json:"amount"             binding:"required"`
}

type Sale struct {
	bun.BaseModel `bun:"table:sale,alias:s" swaggerignore:"true"`

	SaleID    int             `bun:"sale_id,pk,autoincrement"             json:"sale_id"`
	Lines     []SaleLine      `bun:"element,type:jsonb"                   json:"lines"`
	Price     decimal.Decimal `bun:"price,type:decimal(10,2)"             json:"price"`
	TimeStamp time.Time       `bun:"time_stamp,default:current_timestamp" json:"time_stamp"`
	Payments  []SalePayment   `bun:"payments,type:jsonb"                  json:"payments"`
}

type CreateSaleInput struct {
	Lines    []SaleLine    `json:"lines"    binding:"required,min=1,dive"`
	Payments []SalePayment `json:"payments" binding:"required,min=1,dive"`
}

type UpdateSaleInput struct {
	Lines    *[]SaleLine    `json:"lines,omitempty"    binding:"omitempty,min=1,dive"`
	Payments *[]SalePayment `json:"payments,omitempty" binding:"omitempty,min=1,dive"`
}
