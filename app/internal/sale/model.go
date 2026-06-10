package sale

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

var ErrSaleNotFound = errors.New("sale not found")
var ErrNewLineIncomplete = errors.New("new line items require a name")
var ErrNewLineZeroQuantity = errors.New("new line items must have quantity >= 1")
var ErrSaleWouldBeEmpty = errors.New("sale must have at least one item with quantity > 0")
var ErrInvalidSaleTotal = errors.New("sale total must be positive")
var ErrInvalidPaymentAmount = errors.New("each payment amount must be positive")
var ErrPayementMethodInvalid = errors.New("payment method not found or inactive")

type SaleLine struct {
	ItemID    int             `json:"item_id"    binding:"required"       example:"1"`
	Name      string          `json:"name"                                 example:"Coffee"`
	Quantity  int             `json:"quantity"   binding:"required,min=1"  example:"2"`
	UnitPrice decimal.Decimal `json:"unit_price" binding:"required"        example:"4.50"`
	TaxRate   decimal.Decimal `json:"tax_rate,omitempty"                   example:"0.20"`
}

type SalePayment struct {
	PayementMethodID int             `json:"payement_method_id" binding:"required"`
	Amount           decimal.Decimal `json:"amount"             binding:"required"`
}

type Sale struct {
	bun.BaseModel `bun:"table:sales,alias:s" swaggerignore:"true"`

	SaleID           int                            `bun:"sale_id,pk,autoincrement"                                  json:"sale_id"`
	Lines            []SaleLine                     `bun:"element,type:jsonb"                                        json:"lines"`
	Price            decimal.Decimal                `bun:"price,type:decimal(10,2)"                                  json:"price"`
	TimeStamp        time.Time                      `bun:"time_stamp,default:current_timestamp"                      json:"time_stamp"`
	PayementMethodID int                            `bun:"payement_method_id"                                        json:"payement_method_id"`
	IsDeleted        bool                           `bun:"is_deleted,default:false"                                  json:"is_deleted"`
	Payments   []SalePayment                   `json:"payments" binding:"required,min=1,dive"`
}

type CreateSaleInput struct {
	Lines            []SaleLine `json:"lines"              binding:"required,min=1,dive"`
	Payments []SalePayment `json:"payments" binding:"required,min=1,dive"`
}

type UpdateSaleLine struct {
	ItemID   int    `json:"item_id"  binding:"required"      example:"2"`
	Name     string `json:"name,omitempty"                   example:"Tea"`
	Quantity *int   `json:"quantity" binding:"required,min=0" example:"1"`
}

type UpdateSaleInput struct {
	Lines            *[]UpdateSaleLine `json:"lines,omitempty"              binding:"omitempty,min=1,dive"`
	Payments []SalePayment `json:"payments" binding:"required,min=1,dive"`
}
