package license

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v84"

	"tili/app/internal/profile"
	"tili/app/internal/store"
	"tili/app/pkg/db"
)

func TestService_GetByID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByID(context.Background(), uuid.New())

	assert.ErrorIs(t, err, ErrLicenceNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_GetByID_WithStripeInfo(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origRetrieveCheckoutSession := retrieveCheckoutSession
	origRetrieveSubscriptionForLicence := retrieveSubscriptionForLicence
	t.Cleanup(func() {
		retrieveCheckoutSession = origRetrieveCheckoutSession
		retrieveSubscriptionForLicence = origRetrieveSubscriptionForLicence
	})

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		assert.Equal(t, "cs_test_123", sessionID)
		return &stripe.CheckoutSession{
			ID:           "cs_test_123",
			Subscription: &stripe.Subscription{ID: "sub_123"},
		}, nil
	}
	retrieveSubscriptionForLicence = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_123", subscriptionID)
		return &stripe.Subscription{
			ID:                "sub_123",
			Status:            "active",
			CancelAtPeriodEnd: true,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						CurrentPeriodEnd: 1714406400,
						Price: &stripe.Price{
							ID:         "price_123",
							UnitAmount: 1999,
							Currency:   stripe.Currency("eur"),
							Recurring: &stripe.PriceRecurring{
								Interval: stripe.PriceRecurringIntervalMonth,
							},
							Product: &stripe.Product{
								ID:   "prod_123",
								Name: "Pro",
							},
						},
					},
				},
			},
		}, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	rows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, accID, "cs_test_123")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	lic, err := svc.GetByID(context.Background(), licID)

	assert.NoError(t, err)
	if assert.NotNil(t, lic) {
		if assert.NotNil(t, lic.Stripe) {
			assert.Equal(t, "sub_123", lic.Stripe.SubscriptionID)
			assert.Equal(t, "active", lic.Stripe.Status)
			assert.True(t, lic.Stripe.CancelAtPeriodEnd)
			assert.Equal(t, int64(1999), lic.Stripe.PriceAmount)
			assert.Equal(t, "EUR", lic.Stripe.PriceCurrency)
			assert.Equal(t, "month", lic.Stripe.PriceInterval)
			assert.Equal(t, "prod_123", lic.Stripe.PriceProductID)
			assert.Equal(t, "Pro", lic.Stripe.PriceProductName)
			if assert.NotNil(t, lic.Stripe.NextPaymentAt) {
				assert.Equal(t, time.Unix(1714406400, 0).UTC(), *lic.Stripe.NextPaymentAt)
			}
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Delete_Forbidden(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	licID := uuid.New()
	forbiddenAccID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, forbiddenAccID)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	err := svc.Delete(context.Background(), accID, licID)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Update_Forbidden(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	licID := uuid.New()
	forbiddenAccID := uuid.MustParse("00000000-0000-0000-0000-000000000099")
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, forbiddenAccID)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(rows)

	_, err := svc.Update(context.Background(), accID, licID, UpdateLicenceInput{})

	assert.ErrorIs(t, err, ErrForbidden)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_CreatePaymentLink_InvalidOffer(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	_, err := svc.CreatePaymentLink(context.Background(), accID, "", CreatePaymentLinkInput{Offer: "weekly"})

	assert.EqualError(t, err, "offre invalide: weekly")
}

func TestService_CreatePaymentLink_MissingConfig(t *testing.T) {
	bunDB, _ := setupMockDB(t)
	defer bunDB.Close()

	t.Setenv("STRIPE_PRICE_MENSUEL", "")
	t.Setenv("APP_URL", "https://example.com")

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	_, err := svc.CreatePaymentLink(context.Background(), accID, "", CreatePaymentLinkInput{Offer: "mensuel"})

	assert.EqualError(t, err, "config manquante pour l'offre: mensuel")
}

func TestService_GetByAccountID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	rows := sqlmock.NewRows([]string{"licence_id", "account_id"}).AddRow(licID, accID)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l"`).WillReturnRows(rows)

	list, err := svc.GetByAccountID(context.Background(), accID)

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, accID, list[0].AccountID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByAccountID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(account_id = .+\)`).WillReturnResult(sqlmock.NewResult(1, 1))

	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	err := svc.DeleteByAccountID(context.Background(), accID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Refund_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origRetrieveCheckoutSession := retrieveCheckoutSession
	origCancelSubscription := cancelSubscription
	origCreateRefund := createRefund
	t.Cleanup(func() {
		retrieveCheckoutSession = origRetrieveCheckoutSession
		cancelSubscription = origCancelSubscription
		createRefund = origCreateRefund
	})

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		assert.Equal(t, "cs_test_123", sessionID)
		return &stripe.CheckoutSession{
			Subscription:  &stripe.Subscription{ID: "sub_123"},
			PaymentIntent: &stripe.PaymentIntent{ID: "pi_123"},
		}, nil
	}
	cancelSubscription = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_123", subscriptionID)
		return &stripe.Subscription{ID: subscriptionID}, nil
	}
	createRefund = func(ctx context.Context, paymentIntentID string) (*stripe.Refund, error) {
		assert.Equal(t, "pi_123", paymentIntentID)
		return &stripe.Refund{ID: "re_123"}, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(store.NewRepository(&db.Db{DB: bunDB}))
	profileSvc := profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))
	svc := NewService(repo, &MockEmailSender{})
	svc.SetDependencies(storeSvc, profileSvc)

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).AddRow(storeID, "Refund Store", accID, licID, time.Now(), "", "")
	licRows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, accID, "cs_test_123")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(licRows)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(licence_id = .+\)$`).WillReturnRows(storeRows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Refund(context.Background(), accID, licID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Refund_Success_ResolvePaymentIntentFromSubscription(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origRetrieveCheckoutSession := retrieveCheckoutSession
	origRetrieveSubscriptionForRefund := retrieveSubscriptionForRefund
	origCancelSubscription := cancelSubscription
	origCreateRefund := createRefund
	t.Cleanup(func() {
		retrieveCheckoutSession = origRetrieveCheckoutSession
		retrieveSubscriptionForRefund = origRetrieveSubscriptionForRefund
		cancelSubscription = origCancelSubscription
		createRefund = origCreateRefund
	})

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		assert.Equal(t, "cs_test_456", sessionID)
		return &stripe.CheckoutSession{
			Subscription: &stripe.Subscription{ID: "sub_456"},
		}, nil
	}
	retrieveSubscriptionForRefund = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_456", subscriptionID)
		return &stripe.Subscription{
			LatestInvoice: &stripe.Invoice{
				Payments: &stripe.InvoicePaymentList{
					Data: []*stripe.InvoicePayment{
						{
							Payment: &stripe.InvoicePaymentPayment{
								PaymentIntent: &stripe.PaymentIntent{ID: "pi_456"},
							},
						},
					},
				},
			},
		}, nil
	}
	cancelSubscription = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_456", subscriptionID)
		return &stripe.Subscription{ID: subscriptionID}, nil
	}
	createRefund = func(ctx context.Context, paymentIntentID string) (*stripe.Refund, error) {
		assert.Equal(t, "pi_456", paymentIntentID)
		return &stripe.Refund{ID: "re_456"}, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(store.NewRepository(&db.Db{DB: bunDB}))
	profileSvc := profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))
	svc := NewService(repo, &MockEmailSender{})
	svc.SetDependencies(storeSvc, profileSvc)

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).AddRow(storeID, "Refund Store 2", accID, licID, time.Now(), "", "")
	licRows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, accID, "cs_test_456")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(licRows)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(licence_id = .+\)$`).WillReturnRows(storeRows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Refund(context.Background(), accID, licID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Refund_Success_NoPaymentReference(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origRetrieveCheckoutSession := retrieveCheckoutSession
	origRetrieveSubscriptionForRefund := retrieveSubscriptionForRefund
	origCancelSubscription := cancelSubscription
	origCreateRefund := createRefund
	t.Cleanup(func() {
		retrieveCheckoutSession = origRetrieveCheckoutSession
		retrieveSubscriptionForRefund = origRetrieveSubscriptionForRefund
		cancelSubscription = origCancelSubscription
		createRefund = origCreateRefund
	})

	retrieveCheckoutSession = func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
		assert.Equal(t, "cs_test_789", sessionID)
		return &stripe.CheckoutSession{
			Subscription: &stripe.Subscription{ID: "sub_789"},
		}, nil
	}
	retrieveSubscriptionForRefund = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_789", subscriptionID)
		return &stripe.Subscription{LatestInvoice: &stripe.Invoice{}}, nil
	}
	cancelSubscription = func(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_789", subscriptionID)
		return &stripe.Subscription{ID: subscriptionID}, nil
	}
	createRefund = func(ctx context.Context, paymentIntentID string) (*stripe.Refund, error) {
		assert.Fail(t, "createRefund should not be called when there is no payment reference")
		return nil, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	storeSvc := store.NewService(store.NewRepository(&db.Db{DB: bunDB}))
	profileSvc := profile.NewService(profile.NewRepository(&db.Db{DB: bunDB}))
	svc := NewService(repo, &MockEmailSender{})
	svc.SetDependencies(storeSvc, profileSvc)

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	storeID := uuid.MustParse("00000000-0000-0000-0000-000000000012")
	storeRows := sqlmock.NewRows([]string{"store_id", "name", "buyer_id", "licence_id", "date_creation", "numero_tva", "siret"}).AddRow(storeID, "Refund Store 3", accID, licID, time.Now(), "", "")
	licRows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, accID, "cs_test_789")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnRows(licRows)
	mock.ExpectQuery(`^SELECT .* FROM "store" AS "s" WHERE \(licence_id = .+\)$`).WillReturnRows(storeRows)
	mock.ExpectExec(`^DELETE FROM "profile" AS "p" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "store" AS "s" WHERE \(store_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.Refund(context.Background(), accID, licID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByStripeSubscriptionID_Success(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origListCheckoutSessionsBySubscription := listCheckoutSessionsBySubscription
	t.Cleanup(func() {
		listCheckoutSessionsBySubscription = origListCheckoutSessionsBySubscription
	})

	listCheckoutSessionsBySubscription = func(ctx context.Context, subscriptionID string) ([]*stripe.CheckoutSession, error) {
		assert.Equal(t, "sub_123", subscriptionID)
		return []*stripe.CheckoutSession{{ID: "cs_test_123"}}, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	licID := uuid.New()
	accID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(transaction = .+\)$`).WillReturnError(sql.ErrNoRows)
	rows := sqlmock.NewRows([]string{"licence_id", "account_id", "transaction"}).AddRow(licID, accID, "cs_test_123")
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(transaction = .+\)$`).WillReturnRows(rows)
	mock.ExpectExec(`^DELETE FROM "licence" AS "l" WHERE \(licence_id = .+\)$`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.DeleteByStripeSubscriptionID(context.Background(), "sub_123")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_DeleteByStripeSubscriptionID_NotFound(t *testing.T) {
	bunDB, mock := setupMockDB(t)
	defer bunDB.Close()

	origListCheckoutSessionsBySubscription := listCheckoutSessionsBySubscription
	t.Cleanup(func() {
		listCheckoutSessionsBySubscription = origListCheckoutSessionsBySubscription
	})

	listCheckoutSessionsBySubscription = func(ctx context.Context, subscriptionID string) ([]*stripe.CheckoutSession, error) {
		assert.Equal(t, "sub_404", subscriptionID)
		return []*stripe.CheckoutSession{{ID: "cs_unknown"}}, nil
	}

	repo := NewRepository(&db.Db{DB: bunDB})
	svc := NewService(repo, &MockEmailSender{})

	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(transaction = .+\)$`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`^SELECT .* FROM "licence" AS "l" WHERE \(transaction = .+\)$`).WillReturnError(sql.ErrNoRows)

	err := svc.DeleteByStripeSubscriptionID(context.Background(), "sub_404")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
