package sale

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Granularity string

const (
	GranularityDaily   Granularity = "daily"
	GranularityWeekly  Granularity = "weekly"
	GranularityMonthly Granularity = "monthly"
)

const otherTaxBracketLabel = "other"

// taxBracketDefs are the tax brackets broken out explicitly in reports; any
// line with a different rate is grouped under otherTaxBracketLabel so totals
// still reconcile with TotalRevenue.
var taxBracketDefs = []struct {
	Label string
	Rate  decimal.Decimal
}{
	{Label: "5%", Rate: decimal.NewFromFloat(0.05)},
	{Label: "10%", Rate: decimal.NewFromFloat(0.10)},
	{Label: "20%", Rate: decimal.NewFromFloat(0.20)},
}

func taxBracketLabel(rate decimal.Decimal) string {
	for _, b := range taxBracketDefs {
		if rate.Equal(b.Rate) {
			return b.Label
		}
	}
	return otherTaxBracketLabel
}

// taxAmountFromInclusive extracts the tax portion of a tax-inclusive amount:
// tax = amount * rate / (1 + rate).
func taxAmountFromInclusive(amount, rate decimal.Decimal) decimal.Decimal {
	if rate.IsZero() {
		return decimal.Zero
	}
	return amount.Mul(rate).Div(decimal.NewFromInt(1).Add(rate))
}

type TaxBracketBreakdown struct {
	Rate      decimal.Decimal `json:"rate"`
	Revenue   decimal.Decimal `json:"revenue"`
	TaxAmount decimal.Decimal `json:"tax_amount"`
}

type ProductBreakdown struct {
	ItemID   uuid.UUID       `json:"item_id"`
	Name     string          `json:"name"`
	Quantity int             `json:"quantity"`
	Revenue  decimal.Decimal `json:"revenue"`
}

type KPIPeriod struct {
	Period              string                          `json:"period"`
	PeriodStart         time.Time                       `json:"period_start"`
	TotalRevenue        decimal.Decimal                 `json:"total_revenue"`
	SalesCount          int                             `json:"sales_count"`
	RevenueByTaxBracket map[string]*TaxBracketBreakdown `json:"revenue_by_tax_bracket"`
	RevenueByProduct    []*ProductBreakdown             `json:"revenue_by_product"`
}

type KPIReport struct {
	Granularity Granularity  `json:"granularity"`
	Periods     []*KPIPeriod `json:"periods"`
}

// bucketFor returns the period key and period start for ts at the given granularity.
// Weekly buckets are ISO weeks (Monday start), labeled "<iso-year>-W<iso-week>".
func bucketFor(ts time.Time, g Granularity) (key string, start time.Time) {
	loc := ts.Location()
	switch g {
	case GranularityMonthly:
		start = time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, loc)
		key = start.Format("2006-01")
	case GranularityWeekly:
		dayStart := time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
		offsetFromMonday := (int(ts.Weekday()) + 6) % 7
		start = dayStart.AddDate(0, 0, -offsetFromMonday)
		isoYear, isoWeek := start.ISOWeek()
		key = fmt.Sprintf("%d-W%02d", isoYear, isoWeek)
	default:
		start = time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, loc)
		key = start.Format("2006-01-02")
	}
	return key, start
}

// aggregateSalesKPI groups sales into periods and computes, per period, total
// revenue, revenue/tax-amount per tax bracket, and revenue per product. Line
// amounts are treated as tax-inclusive (UnitPrice already includes tax).
func aggregateSalesKPI(sales []*Sale, granularity Granularity) *KPIReport {
	periods := make([]*KPIPeriod, 0)
	periodIndex := make(map[string]*KPIPeriod)
	productIndex := make(map[string]map[uuid.UUID]*ProductBreakdown)

	for _, sl := range sales {
		key, start := bucketFor(sl.TimeStamp, granularity)
		period, ok := periodIndex[key]
		if !ok {
			period = &KPIPeriod{
				Period:              key,
				PeriodStart:         start,
				TotalRevenue:        decimal.Zero,
				RevenueByTaxBracket: map[string]*TaxBracketBreakdown{},
				RevenueByProduct:    []*ProductBreakdown{},
			}
			periodIndex[key] = period
			productIndex[key] = map[uuid.UUID]*ProductBreakdown{}
			periods = append(periods, period)
		}
		period.SalesCount++

		for _, line := range sl.Lines {
			qty := decimal.NewFromInt(int64(line.Quantity))
			lineRevenue := line.UnitPrice.Mul(qty)
			period.TotalRevenue = period.TotalRevenue.Add(lineRevenue)

			label := taxBracketLabel(line.TaxRate)
			bracket, ok := period.RevenueByTaxBracket[label]
			if !ok {
				rate := line.TaxRate
				if label == otherTaxBracketLabel {
					rate = decimal.Zero
				}
				bracket = &TaxBracketBreakdown{Rate: rate, Revenue: decimal.Zero, TaxAmount: decimal.Zero}
				period.RevenueByTaxBracket[label] = bracket
			}
			bracket.Revenue = bracket.Revenue.Add(lineRevenue)
			bracket.TaxAmount = bracket.TaxAmount.Add(taxAmountFromInclusive(lineRevenue, line.TaxRate))

			prod, ok := productIndex[key][line.ItemID]
			if !ok {
				prod = &ProductBreakdown{ItemID: line.ItemID, Name: line.Name, Revenue: decimal.Zero}
				productIndex[key][line.ItemID] = prod
				period.RevenueByProduct = append(period.RevenueByProduct, prod)
			}
			prod.Quantity += line.Quantity
			prod.Revenue = prod.Revenue.Add(lineRevenue)
		}
	}

	for _, p := range periods {
		p.TotalRevenue = p.TotalRevenue.Round(2)
		for _, b := range p.RevenueByTaxBracket {
			b.Revenue = b.Revenue.Round(2)
			b.TaxAmount = b.TaxAmount.Round(2)
		}
		for _, pr := range p.RevenueByProduct {
			pr.Revenue = pr.Revenue.Round(2)
		}
		sort.Slice(p.RevenueByProduct, func(i, j int) bool {
			return p.RevenueByProduct[i].Revenue.GreaterThan(p.RevenueByProduct[j].Revenue)
		})
	}

	return &KPIReport{Granularity: granularity, Periods: periods}
}

func isValidGranularity(g Granularity) bool {
	return g == GranularityDaily || g == GranularityWeekly || g == GranularityMonthly
}

func (s *Service) GetSalesKPI(ctx context.Context, granularity Granularity, from, to *time.Time) (*KPIReport, error) {
	if !isValidGranularity(granularity) {
		return nil, ErrInvalidGranularity
	}

	sales, err := s.repo.FindInRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return aggregateSalesKPI(sales, granularity), nil
}
