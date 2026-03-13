package catalog

import (
	"github.com/uptrace/bun"
)

type catalog struct {
	bun.BaseModel `bun:"table:catalog,alias:c" swaggerignore:"true"`

	catalogID   int    `bun:"catalog_id,pk,autoincrement" json:"catalog_id"  example:"1"`
	Name        string `bun:"name,notnull"                json:"name"        example:"Winter 2026 Collection"`
	Description string `bun:"description"                 json:"description" example:"All items available for the winter 2026 season"`
}

type catalogUpdate struct {
	Name        *string `json:"name"        example:"Winter 2026 Collection"`
	Description *string `json:"description" example:"All items available for the winter 2026 season"`
}
