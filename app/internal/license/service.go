package license

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tili/app/pkg/email"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/customer"
	refundapi "github.com/stripe/stripe-go/v84/refund"
	subscriptionapi "github.com/stripe/stripe-go/v84/subscription"

	"tili/app/internal/profile"
	"tili/app/internal/store"
)

var (
	ErrLicenceNotFound = errors.New("licence not found")
	ErrForbidden       = errors.New("forbidden")

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		_ = ctx
		return session.Get(sessionID, nil)
	}

	cancelSubscription = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		_ = ctx
		return subscriptionapi.Cancel(subscriptionID, nil)
	}

	retrieveSubscriptionForRefund = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		_ = ctx
		params := &stripe.SubscriptionParams{}
		params.AddExpand("latest_invoice.payments")
		// Stripe max expansion depth is 4 levels.
		params.AddExpand("latest_invoice.payments.data.payment")
		return subscriptionapi.Get(subscriptionID, params)
	}

	retrieveSubscriptionForLicence = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		_ = ctx
		params := &stripe.SubscriptionParams{}
		params.AddExpand("items.data.price.product")
		return subscriptionapi.Get(subscriptionID, params)
	}

	listCheckoutSessionsBySubscription = func(ctx context.Context, subscriptionID string) ([]*stripe.CheckoutSession, error) {
		_ = ctx
		params := &stripe.CheckoutSessionListParams{Subscription: stripe.String(subscriptionID)}
		iter := session.List(params)
		result := make([]*stripe.CheckoutSession, 0)
		for iter.Next() {
			result = append(result, iter.CheckoutSession())
		}
		if err := iter.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}

	createRefund = func(ctx context.Context, paymentIntentID string) (*stripe.Refund, error) {
		_ = ctx
		return refundapi.New(&stripe.RefundParams{
			PaymentIntent: stripe.String(paymentIntentID),
			Reason:        stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
		})
	}
)

type Service struct {
	repo           *Repository
	emailClient    email.Sender
	storeService   *store.Service
	profileService *profile.Service
}

func NewService(repo *Repository, emailClient email.Sender) *Service {
	return &Service{repo: repo, emailClient: emailClient}
}

func (s *Service) SetDependencies(storeService *store.Service, profileService *profile.Service) {
	s.storeService = storeService
	s.profileService = profileService
}

func (s *Service) DeleteByAccountID(ctx context.Context, accountID uuid.UUID) error {
	return s.repo.DeleteLicencesByAccountID(ctx, accountID)
}

func (s *Service) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]Licence, error) {
	licences, err := s.repo.FindLicencesByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if err := s.enrichLicencesWithStripe(ctx, licences); err != nil {
		return nil, err
	}
	return licences, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Licence, error) {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenceNotFound
		}
		return nil, err
	}
	if err := s.enrichLicenceWithStripe(ctx, lic); err != nil {
		return nil, err
	}
	return lic, nil
}

func (s *Service) enrichLicencesWithStripe(ctx context.Context, licences []Licence) error {
	for i := range licences {
		if err := s.enrichLicenceWithStripe(ctx, &licences[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) enrichLicenceWithStripe(ctx context.Context, lic *Licence) error {
	if lic == nil {
		return nil
	}
	transaction := strings.TrimSpace(lic.Transaction)
	if transaction == "" {
		return nil
	}

	stripeInfo, err := s.fetchStripeLicenceInfo(ctx, transaction)
	if err != nil {
		return err
	}
	lic.Stripe = stripeInfo
	return nil
}

func (s *Service) fetchStripeLicenceInfo(ctx context.Context, transaction string) (*StripeLicenceInfo, error) {
	session, err := retrieveCheckoutSession(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("retrieve checkout session: %w", err)
	}
	if session == nil || session.Subscription == nil || session.Subscription.ID == "" {
		return nil, fmt.Errorf("checkout session %s does not contain a subscription", transaction)
	}

	subscription, err := retrieveSubscriptionForLicence(ctx, session.Subscription.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieve stripe subscription: %w", err)
	}

	info := &StripeLicenceInfo{
		CheckoutSessionID: session.ID,
		SubscriptionID:    subscription.ID,
		Status:            string(subscription.Status),
		CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
	}

	if subscription.Items != nil && len(subscription.Items.Data) > 0 {
		item := subscription.Items.Data[0]
		if item != nil && item.Price != nil {
			if item.CurrentPeriodEnd > 0 {
				nextPayment := time.Unix(item.CurrentPeriodEnd, 0).UTC()
				info.NextPaymentAt = &nextPayment
			}
			info.PriceAmount = item.Price.UnitAmount
			info.PriceCurrency = strings.ToUpper(string(item.Price.Currency))
			if item.Price.Recurring != nil {
				info.PriceInterval = string(item.Price.Recurring.Interval)
			}
			if item.Price.Product != nil {
				info.PriceProductID = item.Price.Product.ID
				info.PriceProductName = item.Price.Product.Name
			}
		}
	}

	return info, nil
}

func (s *Service) Delete(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLicenceNotFound
		}
		return err
	}
	if lic.AccountID != accountID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Refund(ctx context.Context, accountID uuid.UUID, id uuid.UUID) error {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLicenceNotFound
		}
		return err
	}
	if lic.AccountID != accountID {
		return ErrForbidden
	}

	if s.storeService == nil || s.profileService == nil {
		return fmt.Errorf("refund dependencies not configured")
	}

	if strings.TrimSpace(lic.Transaction) == "" {
		return fmt.Errorf("licence transaction is missing")
	}

	sess, err := retrieveCheckoutSession(ctx, lic.Transaction)
	if err != nil {
		return fmt.Errorf("retrieve checkout session: %w", err)
	}

	if sess.Subscription == nil || sess.Subscription.ID == "" {
		return fmt.Errorf("checkout session %s does not contain a subscription", lic.Transaction)
	}

	if _, err := cancelSubscription(ctx, sess.Subscription.ID); err != nil {
		return fmt.Errorf("cancel stripe subscription: %w", err)
	}

	paymentIntentID := ""
	chargeID := ""
	if sess.PaymentIntent != nil {
		paymentIntentID = sess.PaymentIntent.ID
	}
	if paymentIntentID == "" {
		sub, err := retrieveSubscriptionForRefund(ctx, sess.Subscription.ID)
		if err != nil {
			return fmt.Errorf("retrieve stripe subscription for refund: %w", err)
		}
		if sub != nil && sub.LatestInvoice != nil && sub.LatestInvoice.Payments != nil {
			for _, invPayment := range sub.LatestInvoice.Payments.Data {
				if invPayment == nil || invPayment.Payment == nil {
					continue
				}
				if invPayment.Payment.PaymentIntent != nil && invPayment.Payment.PaymentIntent.ID != "" {
					paymentIntentID = invPayment.Payment.PaymentIntent.ID
					break
				}
				if chargeID == "" && invPayment.Payment.Charge != nil && invPayment.Payment.Charge.ID != "" {
					chargeID = invPayment.Payment.Charge.ID
				}
			}
		}
	}

	if paymentIntentID != "" {
		if _, err := createRefund(ctx, paymentIntentID); err != nil {
			return fmt.Errorf("create stripe refund: %w", err)
		}
	} else if chargeID != "" {
		if _, err := refundapi.New(&stripe.RefundParams{
			Charge: stripe.String(chargeID),
			Reason: stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
		}); err != nil {
			return fmt.Errorf("create stripe refund from charge: %w", err)
		}
	}

	st, err := s.storeService.FindByLicenceID(ctx, id)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			return s.repo.Delete(ctx, id)
		}
		return err
	}
	if st.BuyerID != accountID {
		return ErrForbidden
	}

	if err := s.profileService.DeleteByStoreID(ctx, st.StoreID); err != nil {
		return err
	}

	if err := s.storeService.DeleteByID(ctx, st.StoreID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, accountID uuid.UUID, id uuid.UUID, input UpdateLicenceInput) (*Licence, error) {
	lic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenceNotFound
		}
		return nil, err
	}
	if lic.AccountID != accountID {
		return nil, ErrForbidden
	}
	if input.Transaction != nil {
		lic.Transaction = *input.Transaction
	}
	if input.IsActive != nil {
		lic.IsActive = *input.IsActive
	}
	return s.repo.Update(ctx, lic)
}

func (s *Service) Create(ctx context.Context, accountID uuid.UUID, input CreateLicenceInput) (*Licence, error) {

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
	emailContent, err := email.GetNewLicenseActiveEmailContent(fmt.Sprintf("%s/admin/shop/new?licenceId=%s", os.Getenv("APP_URL"), lic.LicenceID))
	if err != nil {
	} else {
		account, err := s.repo.GetAccountByID(ctx, accountID)
		if err != nil {
			log.Printf("Error fetching account for email: %v", err)
		} else {
			if err := s.emailClient.SendEmail(account.Email, "Your new license is active!", emailContent); err != nil {
				log.Printf("Error sending email: %v", err)
			}
		}
	}
	return lic, nil
}

func (s *Service) CreatePaymentLink(ctx context.Context, accountID uuid.UUID, customerID string, input CreatePaymentLinkInput) (string, error) {
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
			"account_id": accountID.String(),
			"offer":      input.Offer,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("erreur stripe: %w", err)
	}

	var targetEmail string
	if customerID != "" {
		cust, err := customer.Get(customerID, nil)
		if err != nil {
			return "", fmt.Errorf("erreur lors de la récupération du client Stripe: %w", err)
		}
		targetEmail = cust.Email
	}

	if targetEmail != "" {
		content, err := email.GetNewPaymentLinkEmailContent(input.Offer, sess.URL)
		if err != nil {
			return "", fmt.Errorf("erreur lors de la génération de l'email: %w", err)
		}

		if err := s.emailClient.SendEmail(targetEmail, "New Payment Link Created", content); err != nil {
			return "", fmt.Errorf("erreur lors de l'envoi de l'email: %w", err)
		}
	}

	return sess.URL, nil
}

func (s *Service) DeleteByStripeSubscriptionID(ctx context.Context, subscriptionID string) error {
	subscriptionID = strings.TrimSpace(subscriptionID)
	if subscriptionID == "" {
		return fmt.Errorf("missing subscription ID")
	}

	// Fallback for any row that might already store subscription ID directly.
	if lic, err := s.repo.FindByTransaction(ctx, subscriptionID); err == nil {
		return s.repo.Delete(ctx, lic.LicenceID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	sessions, err := listCheckoutSessionsBySubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("list checkout sessions by subscription: %w", err)
	}

	for _, sess := range sessions {
		if sess == nil || sess.ID == "" {
			continue
		}
		lic, err := s.repo.FindByTransaction(ctx, sess.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		return s.repo.Delete(ctx, lic.LicenceID)
	}

	return nil
}
