package payementmethod

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PayementMethod struct {
	bun.BaseModel `bun:"table:payementmethod,alias:pm" swaggerignore:"true"`

	PayementMethodID uuid.UUID `bun:"payement_method_id,pk,type:uuid,default:gen_random_uuid()" json:"payement_method_id" example:"00000000-0000-0000-0000-000000000000"`
	Name             string    `bun:"name,notnull"                                              json:"name"               example:"Credit Card"`
	IsActive         bool      `bun:"is_active,notnull,default:true"                            json:"is_active"          example:"true"`
}

type PayementMethodInput struct {
	Name string `json:"name" example:"Credit Card"`
}
