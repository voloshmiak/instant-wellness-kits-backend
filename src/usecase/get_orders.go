package usecase

import (
	"InstantWellnessKits/src/entity"
	"context"
)

type ListOrdersUseCase struct {
	orders Orders
}

func NewListOrdersUseCase(orders Orders) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orders: orders,
	}
}

func (uc *ListOrdersUseCase) Execute(ctx context.Context, params entity.ListParams) (*entity.ListResult, error) {
	return uc.orders.List(ctx, params)
}