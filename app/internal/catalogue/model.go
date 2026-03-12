package catalogue

import (
	"github.com/uptrace/bun"
)

type Catalogue struct {
	bun.BaseModel `bun:"table:catalogue,alias:c" swaggerignore:"true"`

	CatalogueID int    `bun:"catalogue_id,pk,autoincrement"                     json:"catalogue_id"`
	Name        string `bun:"name,notnull"                                      json:"name"`
	Description string `bun:"description" json:"description"`
}

type CatalogueUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
