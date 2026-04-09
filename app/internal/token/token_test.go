package token

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test_secret")
	secretKey = []byte("test_secret")

	token, err := CreateAccountToken(1, "Test User", "test@example.com", "cus_123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateAccountToken(token)
	assert.NoError(t, err)
	assert.Equal(t, 1, claims.AccountID)
	assert.Equal(t, "Test User", claims.Name)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "cus_123", claims.CustomerID)

	claims2, err := ValidateAccountToken("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, 1, claims2.AccountID)

	_, err = ValidateAccountToken("invalid_token")
	assert.Error(t, err)
}

func TestProfileToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test_secret")
	secretKey = []byte("test_secret")

	token, err := CreateProfileToken(2, "Profile User", 3, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateProfileToken(token)
	assert.NoError(t, err)
	assert.Equal(t, 2, claims.ProfileID)
	assert.Equal(t, "Profile User", claims.Name)
	assert.Equal(t, 3, claims.LevelAccess)
	assert.Equal(t, 10, claims.StoreID)

	claims2, err := ValidateProfileToken("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, 2, claims2.ProfileID)

	_, err = ValidateProfileToken("invalid.token.string")
	assert.Error(t, err)
}
