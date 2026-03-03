package vente

import (
	"time"

	"tili/app/internal/payementmethod"

	"github.com/uptrace/bun"
)

type Vente struct {
	bun.BaseModel `bun:"table:vente,alias:v" swaggerignore:"true"`

	VenteID          int64                          `bun:"vente_id,pk,autoincrement"                          json:"vente_id"`
	Element          map[string]interface{}         `bun:"element,type:jsonb"                                 json:"element,omitempty"`
	Price            float64                        `bun:"price,type:decimal(10,2)"                           json:"price"`
	TimeStamp        time.Time                      `bun:"time_stamp,default:current_timestamp"               json:"time_stamp"`
	PayementMethodID int64                          `bun:"payement_method_id"                                 json:"payement_method_id"`
	PayementMethod   *payementmethod.PayementMethod `bun:"rel:belongs-to,join:payement_method_id=payement_method_id" json:"payement_method,omitempty"`
}
