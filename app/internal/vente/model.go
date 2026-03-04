package sale

import (
	"time"

	"tili/app/internal/payementmethod"

	"github.com/uptrace/bun"
)

type Sales struct {
	bun.BaseModel `bun:"table:sale,alias:v" swaggerignore:"true"`

	Sale_ID          int64                          `bun:"sale_id,pk,autoincrement"                          json:"sale_id"`
	Element          map[string]interface{}         `bun:"element,type:jsonb"                                 json:"element,omitempty"`
	Price            float64                        `bun:"price,type:decimal(10,2)"                           json:"price"`
	TimeStamp        time.Time                      `bun:"time_stamp,default:current_timestamp"               json:"time_stamp"`
	PayementMethodID int64                          `bun:"payement_method_id"                                 json:"payement_method_id"`
	PayementMethod   *payementmethod.PayementMethod `bun:"rel:belongs-to,join:payement_method_id=payement_method_id" json:"payement_method,omitempty"`
}
