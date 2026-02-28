package controller

import (
	"InstantWellnessKits/src/usecase"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type createOrderRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Subtotal  *int     `json:"subtotal"`
	Timestamp *string  `json:"timestamp"`
}

type CreateController struct {
	uc *usecase.CreateOrderUseCase
}

func NewCreateController(uc *usecase.CreateOrderUseCase) *CreateController {
	return &CreateController{
		uc: uc,
	}
}

func (h *CreateController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var body createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Latitude == nil || body.Longitude == nil || body.Subtotal == nil || body.Timestamp == nil {
		http.Error(rw, "Missing required fields", http.StatusBadRequest)
		return
	}

	order, err := h.uc.Execute(r.Context(), *body.Latitude, *body.Longitude,
		*body.Subtotal, *body.Timestamp)
	if err != nil {
		if errors.Is(err, usecase.ErrFailedParsingTimestamp) {
			http.Error(rw, "Invalid timestamp format", http.StatusBadRequest)
			return
		}
		http.Error(rw, "Failed to create order", http.StatusInternalServerError)
		log.Println("Error executing use case:", err)
		return
	}

	encodedOrder, err := json.Marshal(order)
	if err != nil {
		http.Error(rw, "Failed to encode order", http.StatusInternalServerError)
		log.Println("Error encoding order:", err)
		return
	}

	_, err = rw.Write(encodedOrder)
	if err != nil {
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusAccepted)
}
