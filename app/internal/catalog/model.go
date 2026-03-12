package catalog

import (
	"github.com/uptrace/bun"
)

type catalog struct {
	bun.BaseModel `bun:"table:catalog,alias:c" swaggerignore:"true"`

	catalogID   int    `bun:"catalog_id,pk,autoincrement"                     json:"catalog_id"`
	Name        string `bun:"name,notnull"                                      json:"name"`
	Description string `bun:"description" json:"description"`
}

type catalogUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
