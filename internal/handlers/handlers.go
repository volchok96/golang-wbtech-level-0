package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/postgres"

	"github.com/bradfitz/gomemcache/memcache"
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

	// Попытка получить заказ из кэша
	orderItem, err := cache.MemCache.Get("order:" + strconv.Itoa(orderID))
	if err == nil {
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
		return
	}

	// Если заказ не найден в кэше, попытка получить его из базы данных
	config, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get config")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	conn, err := postgres.ConnectToDB(&config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to DB")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())

	order, err := postgres.GetOrderFromDB(context.Background(), conn, orderID)
	if err != nil {
		log.Warn().Msgf("Order not found for ID: %d", orderID)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Обновление кэша
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling order")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = cache.MemCache.Set(&memcache.Item{Key: "order:" + strconv.Itoa(orderID), Value: orderData})
	if err != nil {
		log.Error().Err(err).Msg("Error saving order to memcache")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
		Order:    *order,
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
