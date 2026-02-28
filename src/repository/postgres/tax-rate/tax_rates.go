package tax_rate

import (
	"InstantWellnessKits/src/entity"
	"context"
	"database/sql"

	"github.com/shopspring/decimal"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) Get(ctx context.Context, jurisdiction *entity.Jurisdiction) (decimal.Decimal,
	*entity.TaxBreakdown, error) {
	compositeRate, taxBreakdown, err := r.findRate(ctx, jurisdiction.City)
	if err == nil {
		return compositeRate, taxBreakdown, nil
	}

	compositeRate, taxBreakdown, err = r.findRate(ctx, jurisdiction.County)
	if err != nil {
		return r.findRate(ctx, "New York State")
	}

	return compositeRate, taxBreakdown, nil
}

func (r *Repository) findRate(ctx context.Context, jurisdictionName string) (decimal.Decimal,
	*entity.TaxBreakdown, error) {
	query := `
		SELECT composite_rate, state_rate, county_rate, city_rate, special_rate
		FROM tax_rates
		WHERE jurisdiction_name = $1
	`
	var compositeRate string
	var taxBreakdown entity.TaxBreakdown
	err := r.conn.QueryRowContext(ctx, query, jurisdictionName).
		Scan(&compositeRate, &taxBreakdown.StateRate, &taxBreakdown.CountyRate,
			&taxBreakdown.CityRate, &taxBreakdown.SpecialRate)
	if err != nil {
		return decimal.Zero, nil, err
	}
	return decimal.RequireFromString(compositeRate), &taxBreakdown, nil
}
