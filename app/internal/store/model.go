package store

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Store struct {
	bun.BaseModel `bun:"table:store,alias:s" swaggerignore:"true"`

	StoreID      int       `bun:"store_id,pk,autoincrement"               json:"store_id"`
	Name         string    `bun:"name,notnull"                            json:"name"`
	BuyerID      int       `bun:"buyer_id,notnull"                        json:"buyer_id"`
	LicenceID    uuid.UUID `bun:"licence_id,notnull,type:uuid"            json:"licence_id"`
	DateCreation time.Time `bun:"date_creation,default:current_timestamp" json:"date_creation"`
	NumeroTVA    string    `bun:"numero_tva"                              json:"numero_tva,omitempty"`
	Siret        string    `bun:"siret"                                   json:"siret,omitempty"`
}

type CreateStoreInput struct {
	Name      string    `json:"name"       binding:"required"`
	BuyerID   int       `json:"buyer_id"`
	LicenceID uuid.UUID `json:"licence_id" binding:"required"`
	NumeroTVA string    `json:"numero_tva"`
	Siret     string    `json:"siret"`
}
