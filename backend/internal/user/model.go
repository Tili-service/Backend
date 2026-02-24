package user

import (
	"time"
)

type User struct {
    ID        int64     `bun:",pk,autoincrement" json:"id"`
    Name      string    `bun:",notnull" json:"name"`
    Email     string    `bun:",unique,notnull" json:"email"`
    CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}
