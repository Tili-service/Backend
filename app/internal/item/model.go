package item

import (
	"tili/app/internal/categorie"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Item struct {
	bun.BaseModel `bun:"table:item,alias:i" swaggerignore:"true"`

	ItemId     int             `bun:"item_id,pk,autoincrement"         json:"item_id"      example:"1"`
	Name       string          `bun:"name,notnull"                     json:"name"         example:"Laptop Pro 15"`
	Price      decimal.Decimal `bun:"price,notnull,type:decimal(10,2)" json:"price"        example:"999.99"`
	Tax        decimal.Decimal `bun:"tax,notnull,type:decimal(5,4)"    json:"tax"          example:"0.2000"`
	Tax_amount decimal.Decimal `bun:"tax_amount,notnull,type:decimal(10,2)" json:"tax_amount" example:"199.99"`

	CategorieID int                  `bun:"categorie_id,notnull"                          json:"categorie_id"       example:"1"`
	Categorie   *categorie.Categorie `bun:"rel:belongs-to,join:categorie_id=categorie_id" json:"categorie,omitempty"`
}

type ItemUpdate struct {
	Name        *string          `json:"name"         example:"Laptop Pro 15"`
	Price       *decimal.Decimal `json:"price"        example:"999.99"`
	Tax         *decimal.Decimal `json:"tax"          example:"0.2000"`
	CategorieID *int             `json:"categorie_id" example:"1"`
}
