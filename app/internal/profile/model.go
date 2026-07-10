package profile

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Profile struct {
	bun.BaseModel `bun:"table:profile,alias:p" swaggerignore:"true"`

	ProfileID   uuid.UUID `bun:"profile_id,pk,type:uuid,default:gen_random_uuid()" json:"profile_id"`
	StoreID     uuid.UUID `bun:"store_id,notnull,type:uuid"                        json:"store_id"`
	Name        string    `bun:"name,notnull"                                      json:"name"`
	Pin         string    `bun:"pin,notnull"                                       json:"-"`
	LevelAccess int       `bun:"level_access,notnull,default:4"                    json:"level_access"`
	IsActive    bool      `bun:"is_active,default:true"                            json:"is_active"`
}

type ProfileWithPin struct {
	ProfileID   uuid.UUID `json:"profile_id"`
	StoreID     uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Pin         string    `json:"pin"`
	LevelAccess int       `json:"level_access"`
	IsActive    bool      `json:"is_active"`
}

type CreateProfileInput struct {
	StoreID     uuid.UUID `json:"-"`
	Name        string    `json:"name"         binding:"required"`
	LevelAccess int       `json:"level_access"`
}

type PinLoginInput struct {
	StoreID uuid.UUID `json:"store_id" binding:"required"`
	Pin     string    `json:"pin"      binding:"required"`
}

type updateProfileInput struct {
	Name        *string `json:"name,omitempty"`
	Pin         *string `json:"pin,omitempty"`
	LevelAccess *int    `json:"level_access,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
