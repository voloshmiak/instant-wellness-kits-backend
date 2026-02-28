package usecase

import (
	"InstantWellnessKits/src/entity"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrFailedParsingTimestamp = errors.New(`failed parsing timestamp`)
)

type GeocodingService interface {
	GetJurisdiction(latitude, longitude float64) (*entity.Jurisdiction, error)
}

type Orders interface {
	Create(ctx context.Context, order *entity.Order) (*entity.Order, error)
	List(ctx context.Context, params entity.ListParams) (*entity.ListResult, error)
	CreateBatch(ctx context.Context, orders []*entity.Order) error
}

type TaxRates interface {
	Get(ctx context.Context, jurisdiction *entity.Jurisdiction) (decimal.Decimal,
		*entity.TaxBreakdown, error)
}

type CreateOrderUseCase struct {
	geocodingService GeocodingService
	orders           Orders
	taxRates         TaxRates
}

func NewCreateOrderUseCase(geocodingService GeocodingService,
	orders Orders, taxRates TaxRates) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		geocodingService: geocodingService,
		orders:           orders,
		taxRates:         taxRates,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context,
	latitude, longitude float64, subtotal int, timestamp string) (*entity.Order, error) {
	juris, err := uc.geocodingService.GetJurisdiction(latitude, longitude)
	if err != nil {
		return nil, err
	}

	if juris.State != "New York" {
		return nil, fmt.Errorf("delivery location is outside New York State (got: %s)", juris.State)
	}

	compositeTaxRate, taxBreakdown, err := uc.taxRates.Get(ctx, juris)
	if err != nil {
		return nil, err
	}

	decimalSubtotal := decimal.NewFromInt(int64(subtotal))

	taxAmount := decimalSubtotal.Mul(compositeTaxRate)
	taxAmount = taxAmount.Round(2)

	totalAmount := decimalSubtotal.Add(taxAmount)

	parsedTimestamp, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		return nil, ErrFailedParsingTimestamp
	}

	order := entity.NewOrder(latitude, longitude,
		decimalSubtotal, compositeTaxRate, taxAmount, totalAmount,
		taxBreakdown, juris, parsedTimestamp)

	return uc.orders.Create(ctx, order)
}
