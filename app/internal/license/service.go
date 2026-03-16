package license

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) DeleteByAccountID(ctx context.Context, accountID int) error {
	return s.repo.DeleteLicencesByAccountID(ctx, accountID)
}

func (s *Service) GetByAccountID(ctx context.Context, accountID int) ([]Licence, error) {
	return s.repo.FindLicencesByAccountID(ctx, accountID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Licence, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, accountID int, id uuid.UUID) error {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("licence not found")
	}
	if lic.AccountID != accountID {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, accountID int, id uuid.UUID, input UpdateLicenceInput) (*Licence, error) {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("licence not found")
	}
	if lic.AccountID != accountID {
		return nil, errors.New("forbidden")
	}
	if input.Transaction != nil {
		lic.Transaction = *input.Transaction
	}
	if input.IsActive != nil {
		lic.IsActive = *input.IsActive
	}
	return s.repo.Update(ctx, lic)
}

func (s *Service) Create(ctx context.Context, accountID int, input CreateLicenceInput) (*Licence, error) {

	lic := &Licence{
		LicenceID:   uuid.New(),
		AccountID:   accountID,
		Expiration:  time.Now().Add(time.Duration(input.DurationDays) * 24 * time.Hour),
		Transaction: input.Transaction,
		IsActive:    true,
	}
	if err := s.repo.CreateLicence(ctx, lic); err != nil {
		return nil, err
	}
	return lic, nil
}

func (s *Service) CreatePaymentLink(ctx context.Context, accountID int, customerID string, input CreatePaymentLinkInput) (string, error) {
	var priceID string

	switch input.Offer {
	case "mensuel":
		priceID = os.Getenv("STRIPE_PRICE_MENSUEL")
	case "semestriel":
		priceID = os.Getenv("STRIPE_PRICE_SEMESTRIEL")
	case "annuel":
		priceID = os.Getenv("STRIPE_PRICE_ANNUEL")
	default:
		return "", fmt.Errorf("offre invalide: %s", input.Offer)
	}

	if priceID == "" {
		return "", fmt.Errorf("config manquante pour l'offre: %s", input.Offer)
	}
	var customerPtr *string
 	if customerID != "" {
 		customerPtr = stripe.String(customerID)
 	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(os.Getenv("APP_URL") + "/admin/licenses?success=true"),
		CancelURL:  stripe.String(os.Getenv("APP_URL") + "/admin/licenses?canceled=true"),
		Customer:   customerPtr,
		Metadata: map[string]string{
			"account_id": fmt.Sprintf("%d", accountID),
			"offer":      input.Offer,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("erreur stripe: %w", err)
	}

	return sess.URL, nil
}
