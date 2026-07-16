package categorie

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Categorie struct {
	bun.BaseModel `bun:"table:categorie,alias:cat" swaggerignore:"true"`

	CategorieID uuid.UUID `bun:"categorie_id,pk,type:uuid,default:gen_random_uuid()" json:"categorie_id" example:"00000000-0000-0000-0000-000000000000"`
	Type        string    `bun:"type"                                                 json:"type"         example:"Electronics"`
	CatalogID   uuid.UUID `bun:"catalog_id,notnull,type:uuid"                         json:"catalog_id"   example:"00000000-0000-0000-0000-000000000000"`
}
