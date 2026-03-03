package license

import (
	"context"
	"time"
	"crypto/rand"
	"math/big"
	"fmt"
	"github.com/google/uuid"

	"tili/backend/internal/store"
	"tili/backend/internal/user"
	"tili/backend/internal/account"
)

type Service struct {
	repo *Repository
	userService *user.Service
}

func NewService(repo *Repository, userService *user.Service) *Service {
	return &Service{
		repo: repo,
		userService: userService,
	}
}

func generatePIN(length int) (string, error) {
	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", length, n), nil
}

func (s *Service) Create(ctx context.Context, input AccountRegistrationInput) (*bodyResponse, error) {
	account := &account.Account{
		LicenceID: uuid.New(),
		ExpiresAt: time.Now().Add(time.Duration(input.LicenceActive) * 24 * time.Hour),
		IsActive:  true,
	}


	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, err
	}

	store := &store.Store{
		StoreName: input.StoreName,
		AccountID: account.AccountID,
	}

	if err := s.repo.CreateStore(ctx, store); err != nil {
		return nil, err
	}
	pin, err := generatePIN(6)
	if err != nil {
		panic(err)
	}
	_, err = s.userService.Create(ctx, user.CreateUserInput{
		StoreID:   store.StoreID,
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password,
		AccessCode: pin,
		AccessLevel: 1,
	})
	if err != nil {
		return nil, err
	}
	bodyResponse := bodyResponse{
		AccountID: account.AccountID,
		UserAccessCode: pin,
		ExpiresAt: account.ExpiresAt,
	}
	return &bodyResponse, nil
}

// func (s *Service) Delete(ctx context.Context, id int64) error {
// 	_, err := s.repo.FindByID(ctx, id)
// 	if err != nil {
// 		return errors.New("account not found")
// 	}
// 	return s.repo.Delete(ctx, id)
// }
