package sale

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var ErrInvalidSaleTotal     = errors.New("sale total must be positive")
var ErrInvalidPaymentsTotal  = errors.New("payments total must equal the sale total")
var ErrInvalidPaymentAmount  = errors.New("each payment amount must be positive")
var ErrPayementMethodInvalid = errors.New("payment method not found or inactive")

type pmChecker interface {
	FindActiveByID(ctx context.Context, id int) error
}

type Service struct {
	repo      *Repository
	pmChecker pmChecker
}

func NewService(repo *Repository, pmChecker pmChecker) *Service {
	return &Service{repo: repo, pmChecker: pmChecker}
}

func computeTotal(lines []SaleLine) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		qty := decimal.NewFromInt(int64(line.Quantity))
		total = total.Add(line.UnitPrice.Mul(qty))
	}
	return total.Round(2)
}

func validatePayments(payments []SalePayment, saleTotal decimal.Decimal) error {
	total := decimal.Zero
	for _, p := range payments {
		if !p.Amount.IsPositive() {
			return ErrInvalidPaymentAmount
		}
		total = total.Add(p.Amount)
	}
	if !total.Round(2).Equal(saleTotal) {
		return ErrInvalidPaymentsTotal
	}
	return nil
}

func (s *Service) CreateSale(ctx context.Context, input CreateSaleInput) (*Sale, error) {
	total := computeTotal(input.Lines)
	if !total.IsPositive() {
		return nil, ErrInvalidSaleTotal
	}
	if err := validatePayments(input.Payments, total); err != nil {
		return nil, err
	}
	for _, p := range input.Payments {
		if err := s.pmChecker.FindActiveByID(ctx, p.PayementMethodID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrPayementMethodInvalid
			}
			return nil, err
		}
	}
	sale := &Sale{
		Lines:     input.Lines,
		Price:     total,
		TimeStamp: time.Now(),
		Payments:  input.Payments,
	}
	return s.repo.Create(ctx, sale)
}

func (s *Service) GetSaleByID(ctx context.Context, id int) (*Sale, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetAllSales(ctx context.Context) ([]*Sale, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) UpdateSale(ctx context.Context, id int, input UpdateSaleInput) (*Sale, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lines := existing.Lines
	if input.Lines != nil {
		lines = *input.Lines
	}
	payments := existing.Payments
	if input.Payments != nil {
		payments = *input.Payments
	}

	total := computeTotal(lines)
	if !total.IsPositive() {
		return nil, ErrInvalidSaleTotal
	}
	if err := validatePayments(payments, total); err != nil {
		return nil, err
	}
	if input.Payments != nil {
		for _, p := range payments {
			if err := s.pmChecker.FindActiveByID(ctx, p.PayementMethodID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, ErrPayementMethodInvalid
				}
				return nil, err
			}
		}
	}

	existing.Lines = lines
	existing.Price = total
	existing.Payments = payments

	return s.repo.Update(ctx, existing)
}

func (s *Service) DeleteSale(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
