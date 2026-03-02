package image

import "github.com/uptrace/bun"

type Image struct {
	bun.BaseModel `bun:"table:image,alias:img" swaggerignore:"true"`

	ImageID int64  `bun:"image_id,pk,autoincrement" json:"image_id"`
	Name    string `bun:"name"                      json:"name"`
	URL     string `bun:"url"                       json:"url"`
}
