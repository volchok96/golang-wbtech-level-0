package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"wb-nats-service/internal/models"
	"wb-nats-service/internal/nats"

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
	if order, ok := nats.Cache[orderID]; ok {

		delivery := order.Delivery
		payment := order.Payment
		items := order.Items

		for i, item := range items {
			item.ID = i + 1
			items[i] = item
		}

		tmpl := template.Must(template.ParseFiles("order.html"))
		data := OrderPageData{
			Order:    order,
			Delivery: delivery,
			Payment:  payment,
			Items:    items,
		}
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Error().Err(err).Msg("Error executing template")
			return
		}
	} else {
		log.Warn().Msgf("Order not found for ID: %d", orderID)
		http.Error(w, "Order not found", http.StatusNotFound)
	}
}
