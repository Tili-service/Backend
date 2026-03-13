package token

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type AccessLevel int

const (
	SuperAdmin AccessLevel = 1
	Admin      AccessLevel = 2
	Manager    AccessLevel = 3
	UserLevel  AccessLevel = 4
)

type AccountClaims struct {
	AccountID  int
	Name       string
	Email      string
	CustomerID string
}

type ProfileClaims struct {
	ProfileID   int
	Name        string
	LevelAccess int
	StoreID     int
}

func CreateAccountToken(accountID int, name, email string, customerID string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"accountID":  accountID,
			"name":       name,
			"email":      email,
			"customerID": customerID,
			"type":       "account",
			"exp":        time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign account token: %w", err)
	}
	return tokenString, nil
}

func ValidateAccountToken(tokenString string) (AccountClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	t, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return AccountClaims{}, fmt.Errorf("invalid token: %w", err)
	}
	if !t.Valid {
		return AccountClaims{}, fmt.Errorf("token is not valid")
	}

	claims, ok := t.Claims.(*jwt.MapClaims)
	if !ok {
		return AccountClaims{}, fmt.Errorf("invalid claims")
	}

	accountID, ok := (*claims)["accountID"].(float64)
	if !ok {
		return AccountClaims{}, fmt.Errorf("accountID not found in token")
	}

	name, ok := (*claims)["name"].(string)
	if !ok {
		return AccountClaims{}, fmt.Errorf("name not found or invalid in token")
	}
	email, ok := (*claims)["email"].(string)
	if !ok {
		return AccountClaims{}, fmt.Errorf("email not found or invalid in token")
	}
	customerID := ""
	if rawCustomerID, exists := (*claims)["customerID"]; exists {
		if cid, ok := rawCustomerID.(string); ok {
			customerID = cid
		}
	}

	return AccountClaims{
		AccountID:  int(accountID),
		Name:       name,
		Email:      email,
		CustomerID: customerID,
	}, nil
}

func CreateProfileToken(profileID int, name string, levelAccess int, storeID int) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"profileID":   profileID,
			"name":        name,
			"levelAccess": levelAccess,
			"storeID":     storeID,
			"type":        "profile",
			"exp":         time.Now().Add(time.Hour * 12).Unix(),
		})

	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign profile token: %w", err)
	}
	return tokenString, nil
}

func ValidateProfileToken(tokenString string) (ProfileClaims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	t, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return ProfileClaims{}, fmt.Errorf("invalid token: %w", err)
	}
	if !t.Valid {
		return ProfileClaims{}, fmt.Errorf("token is not valid")
	}

	claims, ok := t.Claims.(*jwt.MapClaims)
	if !ok {
		return ProfileClaims{}, fmt.Errorf("invalid claims")
	}

	profileID, ok := (*claims)["profileID"].(float64)
	if !ok {
		return ProfileClaims{}, fmt.Errorf("profileID not found in token")
	}

	levelAccess, ok := (*claims)["levelAccess"].(float64)
	if !ok {
		return ProfileClaims{}, fmt.Errorf("levelAccess not found in token")
	}

	storeID, ok := (*claims)["storeID"].(float64)
	if !ok {
		return ProfileClaims{}, fmt.Errorf("storeID not found in token")
	}

	return ProfileClaims{
		ProfileID:   int(profileID),
		Name:        (*claims)["name"].(string),
		LevelAccess: int(levelAccess),
		StoreID:     int(storeID),
	}, nil
}
