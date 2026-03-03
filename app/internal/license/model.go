package license

import (
	"time"
)

type AccountRegistrationInput struct {
	StoreName        string `json:"store_name"          binding:"required"`
	Name        string `json:"name"          binding:"required"`
	Email       string `json:"email"         binding:"required,email"`
	Password	string `json:"password"      binding:"required,min=6"`
	LicenceActive int    `json:"licence_active"  binding:"required"`
}

type AccountDeleting struct {
	AccountID        int64 `json:"account_id"          binding:"required"`
}

type bodyResponse struct {
	AccountID  int64  `json:"account_id"`
	UserAccessCode string `json:"user_access_code"`
	ExpiresAt  time.Time  `json:"expires_at"`
}
