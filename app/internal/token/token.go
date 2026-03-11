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

type Claims struct {
	UserID      int64
	Name        string
	Email       string
	AccessLevel int
}

func Create(userID int64, name, email string, accessLevel int) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userID":      userID,
			"name":        name,
			"email":       email,
			"accessLevel": accessLevel,
			"exp":         time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := t.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func Validate(tokenString string) (Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	t, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return Claims{}, fmt.Errorf("invalid token: %w", err)
	}
	if !t.Valid {
		return Claims{}, fmt.Errorf("token is not valid")
	}

	claims, ok := t.Claims.(*jwt.MapClaims)
	if !ok {
		return Claims{}, fmt.Errorf("invalid claims")
	}

	userID, ok := (*claims)["userID"].(float64)
	if !ok {
		return Claims{}, fmt.Errorf("userID not found in token")
	}

	accessLevel, ok := (*claims)["accessLevel"].(float64)
	if !ok {
		return Claims{}, fmt.Errorf("accessLevel not found in token")
	}

	return Claims{
		UserID:      int64(userID),
		Name:        (*claims)["name"].(string),
		Email:       (*claims)["email"].(string),
		AccessLevel: int(accessLevel),
	}, nil
}
