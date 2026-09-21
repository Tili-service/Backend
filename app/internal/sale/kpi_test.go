package sale

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBucketFor_Daily(t *testing.T) {
	ts := time.Date(2024, 3, 15, 13, 45, 0, 0, time.UTC)
	key, start := bucketFor(ts, GranularityDaily)
	assert.Equal(t, "2024-03-15", key)
	assert.Equal(t, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), start)
}

func TestBucketFor_Monthly(t *testing.T) {
	ts := time.Date(2024, 3, 15, 13, 45, 0, 0, time.UTC)
	key, start := bucketFor(ts, GranularityMonthly)
	assert.Equal(t, "2024-03", key)
	assert.Equal(t, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), start)
}

func TestBucketFor_Weekly_MondayAligned(t *testing.T) {
	// 2024-03-13 is a Wednesday; the ISO week starts Monday 2024-03-11.
	ts := time.Date(2024, 3, 13, 9, 0, 0, 0, time.UTC)
	key, start := bucketFor(ts, GranularityWeekly)
	assert.Equal(t, "2024-W11", key)
	assert.Equal(t, time.Date(2024, 3, 11, 0, 0, 0, 0, time.UTC), start)

	// Sunday 2024-03-17, the last day of the same ISO week, must land in the
	// same bucket as the Wednesday above.
	sundayKey, sundayStart := bucketFor(time.Date(2024, 3, 17, 23, 0, 0, 0, time.UTC), GranularityWeekly)
	assert.Equal(t, key, sundayKey)
	assert.Equal(t, start, sundayStart)
}

func TestBucketFor_Weekly_YearBoundary(t *testing.T) {
	// 2023-01-01 is a Sunday and belongs to ISO week 52 of 2022, not 2023.
	key, _ := bucketFor(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), GranularityWeekly)
	assert.Equal(t, "2022-W52", key)

	// 2021-01-01 is a Friday and belongs to ISO week 53 of 2020.
	key, _ = bucketFor(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), GranularityWeekly)
	assert.Equal(t, "2020-W53", key)
}

func TestTaxAmountFromInclusive(t *testing.T) {
	// 12.00 inclusive of 20% tax -> 2.00 tax.
	got := taxAmountFromInclusive(decimal.NewFromInt(12), decimal.NewFromFloat(0.20))
	assert.True(t, decimal.NewFromInt(2).Equal(got), "got %s", got)

	assert.True(t, decimal.Zero.Equal(taxAmountFromInclusive(decimal.NewFromInt(12), decimal.Zero)))

	// rate == -1 would divide by zero; must not panic, only ever reachable
	// via data stored before validateTaxRates existed.
	assert.NotPanics(t, func() {
		taxAmountFromInclusive(decimal.NewFromInt(12), decimal.NewFromInt(-1))
	})
}

func TestTaxBracketLabel(t *testing.T) {
	assert.Equal(t, "5%", taxBracketLabel(decimal.NewFromFloat(0.05)))
	assert.Equal(t, "10%", taxBracketLabel(decimal.NewFromFloat(0.10)))
	assert.Equal(t, "20%", taxBracketLabel(decimal.NewFromFloat(0.20)))
	assert.Equal(t, otherTaxBracketLabel, taxBracketLabel(decimal.NewFromFloat(0.15)))
}

func TestAggregateSalesKPI_TotalsAndBrackets(t *testing.T) {
	ts := time.Date(2024, 3, 13, 9, 0, 0, 0, time.UTC)
	itemA := uuid.New()
	itemB := uuid.New()

	sale1Lines := []SaleLine{
		{ItemID: itemA, Name: "Coffee", Quantity: 2, UnitPrice: decimal.NewFromFloat(4.80), TaxRate: decimal.NewFromFloat(0.20)},
		{ItemID: itemB, Name: "Book", Quantity: 1, UnitPrice: decimal.NewFromFloat(10.50), TaxRate: decimal.NewFromFloat(0.05)},
	}
	sale2Lines := []SaleLine{
		{ItemID: itemA, Name: "Coffee", Quantity: 1, UnitPrice: decimal.NewFromFloat(4.80), TaxRate: decimal.NewFromFloat(0.20)},
		{ItemID: itemB, Name: "Book", Quantity: 1, UnitPrice: decimal.NewFromFloat(9.00), TaxRate: decimal.NewFromFloat(0.15)},
	}
	sales := []*Sale{
		{
			TimeStamp: ts,
			Lines:     sale1Lines,
			// Price mirrors what CreateSale would have stored via computeTotal,
			// since aggregateSalesKPI now sources TotalRevenue from it directly.
			Price: computeTotal(sale1Lines),
		},
		{
			// Same day, second sale: item A repeats (should accumulate) and an
			// "other" tax rate line is introduced.
			TimeStamp: ts.Add(2 * time.Hour),
			Lines:     sale2Lines,
			Price:     computeTotal(sale2Lines),
		},
	}

	report := aggregateSalesKPI(sales, GranularityDaily)

	assert.Len(t, report.Periods, 1)
	period := report.Periods[0]
	assert.Equal(t, "2024-03-13", period.Period)
	assert.Equal(t, 2, period.SalesCount)

	// Total: 2*4.80 + 10.50 + 1*4.80 + 9.00 = 33.90
	assert.True(t, decimal.NewFromFloat(33.90).Equal(period.TotalRevenue), "got %s", period.TotalRevenue)

	bracket20, ok := period.RevenueByTaxBracket["20%"]
	assert.True(t, ok)
	assert.True(t, decimal.NewFromFloat(14.40).Equal(bracket20.Revenue), "got %s", bracket20.Revenue)

	bracket5, ok := period.RevenueByTaxBracket["5%"]
	assert.True(t, ok)
	assert.True(t, decimal.NewFromFloat(10.50).Equal(bracket5.Revenue), "got %s", bracket5.Revenue)

	other, ok := period.RevenueByTaxBracket[otherTaxBracketLabel]
	assert.True(t, ok)
	assert.True(t, decimal.NewFromFloat(9.00).Equal(other.Revenue), "got %s", other.Revenue)
	if assert.NotNil(t, other.Rate, "the other bracket must report the real rate, not zero") {
		assert.True(t, decimal.NewFromFloat(0.15).Equal(*other.Rate))
	}

	var revenueByItem = map[uuid.UUID]*ProductBreakdown{}
	for _, p := range period.RevenueByProduct {
		revenueByItem[p.ItemID] = p
	}
	assert.Equal(t, 3, revenueByItem[itemA].Quantity)
	assert.True(t, decimal.NewFromFloat(14.40).Equal(revenueByItem[itemA].Revenue))
	assert.Equal(t, 2, revenueByItem[itemB].Quantity)
	assert.True(t, decimal.NewFromFloat(19.50).Equal(revenueByItem[itemB].Revenue))
}

func TestAggregateSalesKPI_RoundingReconciles(t *testing.T) {
	ts := time.Date(2024, 3, 13, 9, 0, 0, 0, time.UTC)
	item := uuid.New()

	// Unit prices with sub-cent precision: each line's raw revenue rounds
	// differently, so the parts must still sum to the reported total.
	lines := []SaleLine{
		{ItemID: item, Name: "A", Quantity: 1, UnitPrice: decimal.NewFromFloat(4.995), TaxRate: decimal.NewFromFloat(0.20)},
		{ItemID: item, Name: "A", Quantity: 1, UnitPrice: decimal.NewFromFloat(4.995), TaxRate: decimal.NewFromFloat(0.20)},
		{ItemID: item, Name: "A", Quantity: 1, UnitPrice: decimal.NewFromFloat(4.995), TaxRate: decimal.NewFromFloat(0.20)},
	}
	sales := []*Sale{
		{
			TimeStamp: ts,
			Lines:     lines,
			Price:     computeTotal(lines),
		},
	}

	report := aggregateSalesKPI(sales, GranularityDaily)
	period := report.Periods[0]

	// The report's total must match what was actually charged (computeTotal
	// sums the unrounded 4.995 lines and rounds once: 14.985 -> 14.99), not
	// what summing three independently-rounded lines (5.00 each) would give.
	assert.True(t, decimal.NewFromFloat(14.99).Equal(period.TotalRevenue), "got %s", period.TotalRevenue)
	assert.True(t, sales[0].Price.Equal(period.TotalRevenue))

	sumBrackets := decimal.Zero
	for _, b := range period.RevenueByTaxBracket {
		sumBrackets = sumBrackets.Add(b.Revenue)
	}
	sumProducts := decimal.Zero
	for _, p := range period.RevenueByProduct {
		sumProducts = sumProducts.Add(p.Revenue)
	}

	assert.True(t, period.TotalRevenue.Equal(sumBrackets), "bracket revenues (%s) must sum to total (%s)", sumBrackets, period.TotalRevenue)
	assert.True(t, period.TotalRevenue.Equal(sumProducts), "product revenues (%s) must sum to total (%s)", sumProducts, period.TotalRevenue)
}

func TestAggregateSalesKPI_OtherBracketMixedRates_RateIsNil(t *testing.T) {
	ts := time.Date(2024, 3, 13, 9, 0, 0, 0, time.UTC)
	item := uuid.New()

	sales := []*Sale{
		{
			TimeStamp: ts,
			Lines: []SaleLine{
				{ItemID: item, Name: "A", Quantity: 1, UnitPrice: decimal.NewFromFloat(10), TaxRate: decimal.NewFromFloat(0.15)},
				{ItemID: item, Name: "A", Quantity: 1, UnitPrice: decimal.NewFromFloat(10), TaxRate: decimal.NewFromFloat(0.07)},
			},
		},
	}

	report := aggregateSalesKPI(sales, GranularityDaily)
	other, ok := report.Periods[0].RevenueByTaxBracket[otherTaxBracketLabel]
	assert.True(t, ok)
	assert.Nil(t, other.Rate, "mixing distinct non-standard rates in \"other\" must not report a misleading single rate")
	assert.True(t, decimal.NewFromFloat(20).Equal(other.Revenue), "got %s", other.Revenue)
}

func TestGetSalesKPI_InvalidGranularity(t *testing.T) {
	svc := &Service{}
	_, err := svc.GetSalesKPI(context.Background(), uuid.New(), Granularity("yearly"), nil, nil)
	assert.ErrorIs(t, err, ErrInvalidGranularity)
}

func TestGetSalesKPI_InvalidRange(t *testing.T) {
	svc := &Service{}
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.GetSalesKPI(context.Background(), uuid.New(), GranularityDaily, &from, &to)
	assert.ErrorIs(t, err, ErrKPIRangeInvalid)
}
