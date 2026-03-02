package catalogue

import (
	"tili/app/internal/categorie"
	"tili/app/internal/image"

	"github.com/uptrace/bun"
)

type Catalogue struct {
	bun.BaseModel `bun:"table:catalogue,alias:c" swaggerignore:"true"`

	CatalogueID int64                `bun:"catalogue_id,pk,autoincrement"                     json:"catalogue_id"`
	Name        string               `bun:"name,notnull"                                      json:"name"`
	Price       float64              `bun:"price,type:decimal(10,2)"                          json:"price"`
	TVA         int16                `bun:"tva,type:smallint"                                 json:"tva"`
	CategorieID int64                `bun:"categorie_id"                                      json:"categorie_id"`
	ImageID     int64                `bun:"image_id"                                          json:"image_id"`
	Categorie   *categorie.Categorie `bun:"rel:belongs-to,join:categorie_id=categorie_id"    json:"categorie,omitempty"`
	Image       *image.Image         `bun:"rel:belongs-to,join:image_id=image_id"             json:"image,omitempty"`
}
