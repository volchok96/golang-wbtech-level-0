package handler

import (
	"html/template"
	"encoding/json"
	"net/http"
	"strconv"
	"wb-nats-service/internal/cache"
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

	orderItem, err := cache.MemCache.Get("order:" + strconv.Itoa(orderID))
	if err != nil {
		log.Warn().Msgf("Order not found for ID: %d", orderID)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	order := models.Order{}
	err = json.Unmarshal(orderItem.Value, &order)
	if err != nil {
		http.Error(w, "Error unmarshalling order", http.StatusInternalServerError)
		return
	}

	delivery := order.Delivery
	payment := order.Payment
	items := order.Items

	for i, item := range items {
		item.ID = i + 1
		items[i] = item
	}

	tmpl := template.Must(template.ParseFiles(".././order.html"))
	data := OrderPageData{
		Order:    order,
		Delivery: delivery,
		Payment:  payment,
		Items:    items,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Error().Err(err).Msg("Error executing template")
		return
	}
}
