package controller

import (
	"InstantWellnessKits/src/entity"
	"InstantWellnessKits/src/usecase"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type listOrdersResponse struct {
	Orders      []*entity.Order `json:"orders"`
	Pagination  pagination      `json:"pagination"`
	GlobalTotal total           `json:"globalTotal"`
	Last24h     total           `json:"last24h"`
}

type pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

type total struct {
	Orders int             `json:"orders"`
	Tax    decimal.Decimal `json:"tax"`
	Grand  decimal.Decimal `json:"grand"`
}

type GetController struct {
	uc *usecase.ListOrdersUseCase
}

func NewGetController(uc *usecase.ListOrdersUseCase) *GetController {
	return &GetController{
		uc: uc,
	}
}

func (h *GetController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := entity.ListParams{
		Page:   parseIntParam(q.Get("page"), 1),
		Limit:  parseIntParam(q.Get("limit"), 20),
		State:  q.Get("state"),
		County: q.Get("county"),
		City:   q.Get("city"),
	}

	if from := q.Get("from"); from != "" {
		if t, err := time.Parse(time.DateOnly, from); err == nil {
			params.From = &t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse(time.DateOnly, to); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			params.To = &t
		}
	}

	result, err := h.uc.Execute(r.Context(), params)
	if err != nil {
		http.Error(rw, "Failed to list orders", http.StatusInternalServerError)
		log.Println("Error executing use case:", err)
		return
	}

	totalPages := result.Total / params.Limit
	if result.Total%params.Limit != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	response := listOrdersResponse{
		Orders: result.Orders,
		Pagination: pagination{
			Total:      result.Total,
			Page:       params.Page,
			Limit:      params.Limit,
			TotalPages: totalPages,
		},
		GlobalTotal: total{
			Orders: result.GlobalOrders,
			Tax:    result.GlobalTax,
			Grand:  result.GlobalGrand,
		},
		Last24h: total{
			Orders: result.Last24hOrders,
			Tax:    result.Last24hTax,
			Grand:  result.Last24hGrand,
		},
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, "Failed to encode orders", http.StatusInternalServerError)
		log.Println("Error encoding orders:", err)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	_, err = rw.Write(encoded)
	if err != nil {
		return
	}
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
