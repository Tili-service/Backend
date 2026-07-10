package catalog

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Catalog struct {
	bun.BaseModel `bun:"table:catalog,alias:c" swaggerignore:"true"`

	CatalogID   uuid.UUID `bun:"catalog_id,pk,type:uuid,default:gen_random_uuid()" json:"catalog_id"  example:"00000000-0000-0000-0000-000000000000"`
	Name        string    `bun:"name,notnull"                                      json:"name"        example:"Winter 2026 Collection"`
	Description string    `bun:"description"                                       json:"description" example:"All items available for the winter 2026 season"`
	StoreID     uuid.UUID `bun:"store_id,notnull,type:uuid"                        json:"store_id"    example:"00000000-0000-0000-0000-000000000000"`
}

type catalogUpdate struct {
	Name        *string `json:"name"        example:"Winter 2026 Collection"`
	Description *string `json:"description" example:"All items available for the winter 2026 season"`
}
