package token

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test_secret")
	secretKey = []byte("test_secret")
	accountID := uuid.New()

	token, err := CreateAccountToken(accountID, "Test User", "test@example.com", "cus_123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateAccountToken(token)
	assert.NoError(t, err)
	assert.Equal(t, accountID, claims.AccountID)
	assert.Equal(t, "Test User", claims.Name)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "cus_123", claims.CustomerID)

	claims2, err := ValidateAccountToken("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, accountID, claims2.AccountID)

	_, err = ValidateAccountToken("invalid_token")
	assert.Error(t, err)
}

func TestProfileToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test_secret")
	secretKey = []byte("test_secret")

	profileID := uuid.New()
	storeID := uuid.New()
	token, err := CreateProfileToken(profileID, "Profile User", 3, storeID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateProfileToken(token)
	assert.NoError(t, err)
	assert.Equal(t, profileID, claims.ProfileID)
	assert.Equal(t, "Profile User", claims.Name)
	assert.Equal(t, 3, claims.LevelAccess)
	assert.Equal(t, storeID, claims.StoreID)

	claims2, err := ValidateProfileToken("Bearer " + token)
	assert.NoError(t, err)
	assert.Equal(t, profileID, claims2.ProfileID)

	_, err = ValidateProfileToken("invalid.token.string")
	assert.Error(t, err)
}
