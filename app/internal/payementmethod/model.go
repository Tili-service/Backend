package payementmethod

import "github.com/uptrace/bun"

type PayementMethod struct {
	bun.BaseModel `bun:"table:payementmethod,alias:pm" swaggerignore:"true"`

	PayementMethodID int    `bun:"payement_method_id,pk,autoincrement"  json:"payement_method_id" example:"1"`
	Name             string `bun:"name,notnull"                         json:"name"               example:"Credit Card"`
	IsActive         bool   `bun:"is_active,notnull,default:true"       json:"is_active"          example:"true"`
}

type PayementMethodInput struct {
	Name string `json:"name" example:"Credit Card"`
}
