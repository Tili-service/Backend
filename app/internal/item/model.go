package item

import (

	"tili/app/internal/categorie"

	"github.com/uptrace/bun"
	"github.com/shopspring/decimal"
)

type Item struct {
	bun.BaseModel `bun:"table:item,alias:i"`

	ItemId int `bun:"item_id,pk,autoincrement" json:"item_id"`
	Name string `bun:"name,notnull" json:"name"`
	Price decimal.Decimal  `bun:"price,notnull,type:decimal(10,2)" json:"price"`
	Tax decimal.Decimal `bun:"tax,notnull,type:decimal(5,4)" json:"tax"`
	Tax_amount decimal.Decimal `bun:"tax_amount,notnull,type:decimal(10,2)" json:"tax_amount"`

	CategorieID int        `bun:"categorie_id,notnull"   json:"categorie_id"`
    Categorie   *categorie.Categorie `bun:"rel:belongs-to,join:categorie_id=categorie_id" json:"categorie,omitempty" swaggerignore:"true"`
}
