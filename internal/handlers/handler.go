package handlers

import (
	"net/http"
	"strconv"
	"wb-nats-service/internal/models"

	"github.com/rs/zerolog/log"
)

type OrderPageData struct {
	Order    models.Order
	Delivery models.Delivery
	Payment  models.Payment
	Items    []models.Items
}

func HandlerOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("id")
	orderID, err := strconv.Atoi(orderIDStr)

	if err != nil {
		log.Error().Err(err).Msg("Invalid order ID")
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}
}
