package sale

import (
	"errors"
	"time"

	"tili/app/internal/payementmethod"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

var ErrSaleNotFound = errors.New("sale not found")
var ErrNewLineIncomplete = errors.New("new line items require name and unit_price")
var ErrNewLineZeroQuantity = errors.New("new line items must have quantity >= 1")
var ErrSaleWouldBeEmpty = errors.New("sale must have at least one item with quantity > 0")

type SaleLine struct {
	ItemID    int             `json:"item_id"    binding:"required"       example:"1"`
	Name      string          `json:"name"                                 example:"Coffee"`
	Quantity  int             `json:"quantity"   binding:"required,min=1"  example:"2"`
	UnitPrice decimal.Decimal `json:"unit_price" binding:"required"        example:"4.50"`
	TaxRate   decimal.Decimal `json:"tax_rate,omitempty"                   example:"0.20"`
}

type Sale struct {
	bun.BaseModel `bun:"table:sales,alias:s" swaggerignore:"true"`

	SaleID           int                            `bun:"sale_id,pk,autoincrement"                                  json:"sale_id"`
	Lines            []SaleLine                     `bun:"element,type:jsonb"                                        json:"lines"`
	Price            decimal.Decimal                `bun:"price,type:decimal(10,2)"                                  json:"price"`
	TimeStamp        time.Time                      `bun:"time_stamp,default:current_timestamp"                      json:"time_stamp"`
	PayementMethodID int                            `bun:"payement_method_id"                                        json:"payement_method_id"`
	IsDeleted        bool                           `bun:"is_deleted,default:false"                                  json:"is_deleted"`
	PayementMethod   *payementmethod.PayementMethod `bun:"rel:belongs-to,join:payement_method_id=payement_method_id" json:"payement_method,omitempty"`
}

type CreateSaleInput struct {
	Lines            []SaleLine `json:"lines"              binding:"required,min=1,dive"`
	PayementMethodID int        `json:"payement_method_id" binding:"required"             example:"1"`
}

type UpdateSaleLine struct {
	ItemID    int              `json:"item_id"             binding:"required"      example:"2"`
	Name      string           `json:"name,omitempty"                              example:"Tea"`
	Quantity  *int             `json:"quantity"            binding:"required,min=0" example:"1"`
	UnitPrice *decimal.Decimal `json:"unit_price,omitempty"                        example:"3.00"`
	TaxRate   *decimal.Decimal `json:"tax_rate,omitempty"                          example:"0.20"`
}

type UpdateSaleInput struct {
	Lines            *[]UpdateSaleLine `json:"lines,omitempty"              binding:"omitempty,min=1,dive"`
	PayementMethodID *int              `json:"payement_method_id,omitempty" example:"2"`
}
