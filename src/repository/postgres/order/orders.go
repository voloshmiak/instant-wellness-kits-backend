package order

import (
	"InstantWellnessKits/src/entity"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) Create(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	query := `
		INSERT INTO orders (id, latitude, longitude, subtotal, composite_tax_rate, tax_amount, total_amount, breakdown, jurisdictions, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	breakdownJSON, err := json.Marshal(order.Breakdown)
	if err != nil {
		return nil, err
	}
	jurisdictionJSON, err := json.Marshal(order.Jurisdiction)
	if err != nil {
		return nil, err
	}
	_, err = r.conn.ExecContext(ctx, query, order.Id, order.Latitude, order.Longitude,
		order.Subtotal, order.CompositeTaxRate, order.TaxAmount, order.TotalAmount,
		breakdownJSON, jurisdictionJSON, order.Timestamp)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *Repository) List(ctx context.Context, params entity.ListParams) (*entity.ListResult, error) {
	var conditions []string
	var args []interface{}
	i := 1

	if params.State != "" {
		conditions = append(conditions, fmt.Sprintf("jurisdictions->>'state' = $%d", i))
		args = append(args, params.State)
		i++
	}
	if params.County != "" {
		conditions = append(conditions, fmt.Sprintf("jurisdictions->>'county' = $%d", i))
		args = append(args, params.County)
		i++
	}
	if params.City != "" {
		conditions = append(conditions, fmt.Sprintf("jurisdictions->>'city' = $%d", i))
		args = append(args, params.City)
		i++
	}
	if params.From != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", i))
		args = append(args, *params.From)
		i++
	}
	if params.To != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", i))
		args = append(args, *params.To)
		i++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.conn.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where), args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	var globalOrders int
	var globalTax, globalGrand decimal.Decimal
	if err := r.conn.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(tax_amount), 0), COALESCE(SUM(total_amount), 0) FROM orders",
	).Scan(&globalOrders, &globalTax, &globalGrand); err != nil {
		return nil, err
	}

	var last24hOrders int
	var last24hTax, last24hGrand decimal.Decimal
	if err := r.conn.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(tax_amount), 0), COALESCE(SUM(total_amount), 0) FROM orders WHERE timestamp >= NOW() - INTERVAL '24 hours'",
	).Scan(&last24hOrders, &last24hTax, &last24hGrand); err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	query := fmt.Sprintf(`
		SELECT id, latitude, longitude, subtotal, composite_tax_rate, 
		       tax_amount, total_amount, breakdown, jurisdictions, timestamp
		FROM orders
		%s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d
	`, where, i, i+1)

	dataArgs := append(args, params.Limit, offset)
	rows, err := r.conn.QueryContext(ctx, query, dataArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*entity.Order, 0)
	for rows.Next() {
		var order entity.Order
		var breakdownData []byte
		var jurisdictionsData []byte

		err := rows.Scan(&order.Id, &order.Latitude, &order.Longitude, &order.Subtotal,
			&order.CompositeTaxRate, &order.TaxAmount, &order.TotalAmount,
			&breakdownData, &jurisdictionsData, &order.Timestamp)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(breakdownData, &order.Breakdown); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(jurisdictionsData, &order.Jurisdiction); err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &entity.ListResult{
		Orders:        orders,
		Total:         total,
		GlobalOrders:  globalOrders,
		GlobalTax:     globalTax,
		GlobalGrand:   globalGrand,
		Last24hOrders: last24hOrders,
		Last24hTax:    last24hTax,
		Last24hGrand:  last24hGrand,
	}, nil
}

func (r *Repository) CreateBatch(ctx context.Context, orders []*entity.Order) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (id, latitude, longitude, subtotal, composite_tax_rate, tax_amount, total_amount, breakdown, jurisdictions, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	for _, order := range orders {
		breakdownJSON, err := json.Marshal(order.Breakdown)
		if err != nil {
			return err
		}
		jurisdictionJSON, err := json.Marshal(order.Jurisdiction)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, order.Id, order.Latitude, order.Longitude,
			order.Subtotal, order.CompositeTaxRate, order.TaxAmount, order.TotalAmount,
			breakdownJSON, jurisdictionJSON, order.Timestamp)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
