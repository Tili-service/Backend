package sale

import (
	"context"
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockPmChecker struct {
	err error
}

func (m *mockPmChecker) FindActiveByID(_ context.Context, _ int) error {
	return m.err
}

func TestService_CreateSale_InvalidPaymentAmount(t *testing.T) {
	svc := &Service{pmChecker: &mockPmChecker{}}
	input := CreateSaleInput{
		Lines:    []SaleLine{{Quantity: 1, UnitPrice: decimal.NewFromInt(10)}},
		Payments: []SalePayment{{PayementMethodID: 1, Amount: decimal.Zero}},
	}
	_, err := svc.CreateSale(context.Background(), input, nil)
	assert.ErrorIs(t, err, ErrInvalidPaymentAmount)
}

func TestService_CreateSale_PaymentsTotalMismatch(t *testing.T) {
	svc := &Service{pmChecker: &mockPmChecker{}}
	input := CreateSaleInput{
		Lines:    []SaleLine{{Quantity: 1, UnitPrice: decimal.NewFromInt(10)}},
		Payments: []SalePayment{{PayementMethodID: 1, Amount: decimal.NewFromInt(5)}},
	}
	_, err := svc.CreateSale(context.Background(), input, nil)
	assert.ErrorIs(t, err, ErrInvalidPaymentsTotal)
}

func TestService_CreateSale_PayementMethodInvalid(t *testing.T) {
	svc := &Service{pmChecker: &mockPmChecker{err: sql.ErrNoRows}}
	input := CreateSaleInput{
		Lines:    []SaleLine{{Quantity: 1, UnitPrice: decimal.NewFromInt(10)}},
		Payments: []SalePayment{{PayementMethodID: 99, Amount: decimal.NewFromInt(10)}},
	}
	_, err := svc.CreateSale(context.Background(), input, nil)
	assert.ErrorIs(t, err, ErrPayementMethodInvalid)
}

func TestValidatePayments_AmountNotPositive(t *testing.T) {
	err := validatePayments([]SalePayment{{PayementMethodID: 1, Amount: decimal.NewFromFloat(-5)}}, decimal.NewFromInt(10))
	assert.ErrorIs(t, err, ErrInvalidPaymentAmount)
}

func TestValidatePayments_TotalMismatch(t *testing.T) {
	err := validatePayments([]SalePayment{{PayementMethodID: 1, Amount: decimal.NewFromInt(5)}}, decimal.NewFromInt(10))
	assert.ErrorIs(t, err, ErrInvalidPaymentsTotal)
}

func TestValidatePayments_Valid(t *testing.T) {
	err := validatePayments([]SalePayment{
		{PayementMethodID: 1, Amount: decimal.NewFromInt(6)},
		{PayementMethodID: 2, Amount: decimal.NewFromInt(4)},
	}, decimal.NewFromInt(10))
	assert.NoError(t, err)
}
