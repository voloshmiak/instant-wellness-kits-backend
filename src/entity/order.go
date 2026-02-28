package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	Id               uuid.UUID       `json:"id"`
	Latitude         float64         `json:"latitude"`
	Longitude        float64         `json:"longitude"`
	Subtotal         decimal.Decimal `json:"subtotal"`
	CompositeTaxRate decimal.Decimal `json:"compositeTaxRate"`
	TaxAmount        decimal.Decimal `json:"taxAmount"`
	TotalAmount      decimal.Decimal `json:"totalAmount"`
	Breakdown        TaxBreakdown    `json:"breakdown"`
	Jurisdiction     Jurisdiction    `json:"jurisdiction"`
	Timestamp        time.Time       `json:"timestamp"`
}

func NewOrder(latitude, longitude float64, subtotal, compositeTaxRate,
	taxAmount, totalAmount decimal.Decimal, breakdown *TaxBreakdown,
	juris *Jurisdiction, timestamp time.Time) *Order {
	return &Order{
		Id:               uuid.New(),
		Latitude:         latitude,
		Longitude:        longitude,
		Subtotal:         subtotal,
		CompositeTaxRate: compositeTaxRate,
		TaxAmount:        taxAmount,
		TotalAmount:      totalAmount,
		Breakdown:        *breakdown,
		Jurisdiction:     *juris,
		Timestamp:        timestamp,
	}
}

type TaxBreakdown struct {
	StateRate   decimal.Decimal `json:"stateRate"`
	CountyRate  decimal.Decimal `json:"countyRate"`
	CityRate    decimal.Decimal `json:"cityRate"`
	SpecialRate decimal.Decimal `json:"specialRate"`
}

func NewTaxBreakdown(stateRate, countyRate, cityRate, specialRate decimal.Decimal) *TaxBreakdown {
	return &TaxBreakdown{
		StateRate:   stateRate,
		CountyRate:  countyRate,
		CityRate:    cityRate,
		SpecialRate: specialRate,
	}
}

type Jurisdiction struct {
	State   string `json:"state"`
	County  string `json:"county"`
	City    string `json:"city"`
	Special string `json:"special"`
}

func NewJurisdiction(state, county, city, special string) *Jurisdiction {
	return &Jurisdiction{
		State:   state,
		County:  county,
		City:    city,
		Special: special,
	}
}

type ListParams struct {
	Page   int
	Limit  int
	State  string
	City   string
	County string
	From   *time.Time
	To     *time.Time
}

type ListResult struct {
	Orders        []*Order
	Total         int
	GlobalOrders  int
	GlobalTax     decimal.Decimal
	GlobalGrand   decimal.Decimal
	Last24hOrders int
	Last24hTax    decimal.Decimal
	Last24hGrand  decimal.Decimal
}
