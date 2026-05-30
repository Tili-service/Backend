package sale

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var ErrInvalidSaleTotal = errors.New("sale total must be positive")
var ErrInvalidPaymentsTotal = errors.New("payments total must equal the sale total")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func computeTotal(lines []SaleLine) decimal.Decimal {
	total := decimal.Zero
	for _, line := range lines {
		qty := decimal.NewFromInt(int64(line.Quantity))
		total = total.Add(line.UnitPrice.Mul(qty))
	}
	return total.Round(2)
}

func computePaymentsTotal(payments []SalePayment) decimal.Decimal {
	total := decimal.Zero
	for _, p := range payments {
		total = total.Add(p.Amount)
	}
	return total.Round(2)
}

func (s *Service) CreateSale(ctx context.Context, input CreateSaleInput) (*Sale, error) {
	total := computeTotal(input.Lines)
	if !total.IsPositive() {
		return nil, ErrInvalidSaleTotal
	}
	if !computePaymentsTotal(input.Payments).Equal(total) {
		return nil, ErrInvalidPaymentsTotal
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
	if !computePaymentsTotal(payments).Equal(total) {
		return nil, ErrInvalidPaymentsTotal
	}

	existing.Lines = lines
	existing.Price = total
	existing.Payments = payments

	return s.repo.Update(ctx, existing)
}

func (s *Service) DeleteSale(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
