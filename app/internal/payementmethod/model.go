package payementmethod

import "github.com/uptrace/bun"

type PayementMethod struct {
	bun.BaseModel `bun:"table:payementmethod,alias:pm" swaggerignore:"true"`

	PayementMethodID int64  `bun:"payement_method_id,pk,autoincrement" json:"payement_method_id"`
	Name             string `bun:"name,notnull"                        json:"name"`
}
