package categorie

import "github.com/uptrace/bun"

type Categorie struct {
	bun.BaseModel `bun:"table:categorie,alias:cat" swaggerignore:"true"`

	CategorieID int64  `bun:"categorie_id,pk,autoincrement" json:"categorie_id"`
	Type        string `bun:"type"                          json:"type"`
}
