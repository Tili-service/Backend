package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"tili/app/pkg/email"

	"github.com/google/uuid"
)

var (
	ErrStoreNotFound          = errors.New("store not found")
	ErrStoreOwnershipMismatch = errors.New("you are not the owner of this store")

	SumupStatusConnected    = "connected"
	SumupStatusDisconnected = "disconnected"
)

type AccountRepository interface {
	FindByID(ctx context.Context, id int) (*AccountData, error)
}

type AccountData struct {
	Email string
}

type Service struct {
	repo        *Repository
	accountRepo AccountRepository
	emailClient email.Sender
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithEmail(repo *Repository, accountRepo AccountRepository, emailClient email.Sender) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
		emailClient: emailClient,
	}
}

func decorateSumupStatus(store *Store) *Store {
	if store == nil {
		return nil
	}

	if store.SumupMerchantCode != "" && store.SumupAccessToken != "" {
		store.SumupStatus = SumupStatusConnected
		return store
	}

	store.SumupStatus = SumupStatusDisconnected
	return store
}

func decorateSumupStatusList(stores []Store) []Store {
	for index := range stores {
		decorateSumupStatus(&stores[index])
	}
	return stores
}

func decorateSumupStatusPointers(stores []*Store) []*Store {
	for _, store := range stores {
		decorateSumupStatus(store)
	}
	return stores
}

func (s *Service) Create(ctx context.Context, input CreateStoreInput, accountID int) (*Store, error) {
	store := &Store{
		Name:      input.Name,
		BuyerID:   accountID,
		LicenceID: input.LicenceID,
		NumeroTVA: input.NumeroTVA,
		Siret:     input.Siret,
	}
	createdStore, err := s.repo.Create(ctx, store)
	if err != nil {
		return nil, err
	}
	decorateSumupStatus(createdStore)

	if s.emailClient != nil && s.accountRepo != nil {
		go s.sendStoreCreatedEmail(ctx, accountID, createdStore.Name, createdStore.StoreID)
	}

	return createdStore, nil
}

func (s *Service) sendStoreCreatedEmail(ctx context.Context, accountID int, storeName string, storeID int) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		log.Printf("Failed to fetch account: %v", err)
		return
	}

	if account == nil {
		log.Printf("Account not found")
		return
	}

	emailContent, err := email.GetNewStoreCreatedEmailContent(storeName, storeID)
	if err != nil {
		log.Printf("Failed to generate email content: %v", err)
		return
	}

	if err := s.emailClient.SendEmail(account.Email, "New Store Created", emailContent); err != nil {
		log.Printf("Failed to send store creation email: %v", err)
	}
}

func (s *Service) LinkSumupCredentials(ctx context.Context, storeID int, accountID int, merchantCode string, accessToken string) (*Store, error) {
	store, err := s.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}

	if store.BuyerID != accountID {
		return nil, ErrStoreOwnershipMismatch
	}

	store.SumupMerchantCode = merchantCode
	store.SumupAccessToken = accessToken
	decorateSumupStatus(store)

	return s.repo.Update(ctx, store)
}

func (s *Service) FindByID(ctx context.Context, id int) (*Store, error) {
	s2, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}
	return decorateSumupStatus(s2), nil
}

func (s *Service) FindByBuyerID(ctx context.Context, buyerID int) ([]Store, error) {
	stores, err := s.repo.FindByBuyerID(ctx, buyerID)
	if err != nil {
		return nil, err
	}
	return decorateSumupStatusList(stores), nil
}

func (s *Service) FindByLicenceID(ctx context.Context, licenceID uuid.UUID) (*Store, error) {
	st, err := s.repo.FindByLicenceID(ctx, licenceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}
	return decorateSumupStatus(st), nil
}

func (s *Service) FindAll(ctx context.Context) ([]*Store, error) {
	stores, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return decorateSumupStatusPointers(stores), nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrStoreNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) DeleteByID(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, input UpdateStoreInput) (*Store, error) {
	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}

	if input.Name != nil {
		store.Name = *input.Name
	}
	if input.NumeroTVA != nil {
		store.NumeroTVA = *input.NumeroTVA
	}
	if input.Siret != nil {
		store.Siret = *input.Siret
	}
	decorateSumupStatus(store)

	return s.repo.Update(ctx, store)
}
