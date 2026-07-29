package image

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Image struct {
	bun.BaseModel `bun:"table:image,alias:img" swaggerignore:"true"`

	ImageID uuid.UUID `bun:"image_id,pk,type:uuid,default:gen_random_uuid()" json:"image_id"`
	Name    string    `bun:"name"                      json:"name"`
	URL     string    `bun:"url"                       json:"url"`
}
