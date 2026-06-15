package categorie

import "github.com/uptrace/bun"

type Categorie struct {
	bun.BaseModel `bun:"table:categorie,alias:cat" swaggerignore:"true"`

	CategorieID int    `bun:"categorie_id,pk,autoincrement" json:"categorie_id" example:"1"`
	Type        string `bun:"type"                          json:"type"         example:"Electronics"`
	CatalogID   int    `bun:"catalog_id,notnull"            json:"catalog_id"   example:"1"`
}
