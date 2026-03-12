package profile

import (
	"github.com/uptrace/bun"
)

type Profile struct {
	bun.BaseModel `bun:"table:profile,alias:p" swaggerignore:"true"`

	ProfileID   int    `bun:"profile_id,pk,autoincrement"    json:"profile_id"`
	StoreID     int    `bun:"store_id,notnull"               json:"store_id"`
	Name        string `bun:"name,notnull"                   json:"name"`
	Pin         string `bun:"pin,notnull"                    json:"-"`
	LevelAccess int    `bun:"level_access,notnull,default:4" json:"level_access"`
	IsActive    bool   `bun:"is_active,default:true"         json:"is_active"`
}

type ProfileWithPin struct {
	ProfileID   int    `json:"profile_id"`
	StoreID     int    `json:"store_id"`
	Name        string `json:"name"`
	Pin         string `json:"pin"`
	LevelAccess int    `json:"level_access"`
	IsActive    bool   `json:"is_active"`
}

type CreateProfileInput struct {
	StoreID     int    `json:"-"`
	Name        string `json:"name"         binding:"required"`
	LevelAccess int    `json:"level_access"`
}

type PinLoginInput struct {
	StoreID int    `json:"store_id" binding:"required"`
	Pin     string `json:"pin"      binding:"required"`
}

type updateProfileInput struct {
	Name        *string `json:"name,omitempty"`
	Pin 	    *string `json:"pin,omitempty"`
	LevelAccess *int    `json:"level_access,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}