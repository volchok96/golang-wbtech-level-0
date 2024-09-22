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
	"wb-kafka-service/pkg/logger" 

	"github.com/bradfitz/gomemcache/memcache"
)

type OrderPageData struct {
	Order    models.Order
	Delivery models.Delivery
	Payment  models.Payment
	Items    []models.Items
}

func HandlerOrder(log *logger.Logger, w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("id")
	orderID, err := strconv.Atoi(orderIDStr)

	if err != nil {
		log.Error("Invalid order ID", err)
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Попытка получить заказ из кэша
	orderItem, err := cache.MemCache.Get("order:" + strconv.Itoa(orderID))
	if err == nil {
		order := models.Order{}
		err = json.Unmarshal(orderItem.Value, &order)
		if err != nil {
			log.Error("Error unmarshalling order", err)
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

		tmpl := template.Must(template.ParseFiles(".././ui.html"))
		data := OrderPageData{
			Order:    order,
			Delivery: delivery,
			Payment:  payment,
			Items:    items,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Error("Error executing template", err)
			return
		}
		return
	}

	// Если заказ не найден в кэше, попытка получить его из базы данных
	cfg, err := config.GetConfig(log) 
	if err != nil {
		log.Error("Failed to get config", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	conn, err := postgres.ConnectToDB(&cfg, log)
	if err != nil {
		log.Error("Failed to connect to DB", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	order, err := postgres.GetOrderFromDB(context.Background(), conn, orderID, log) // Передаем логгер
	if err != nil {
		log.Warn("Order not found", nil)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Обновление кэша
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error("Error marshalling order", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = cache.MemCache.Set(&memcache.Item{Key: "order:" + strconv.Itoa(orderID), Value: orderData})
	if err != nil {
		log.Error("Error saving order to memcache", err)
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

	tmpl := template.Must(template.ParseFiles(".././ui.html"))
	data := OrderPageData{
		Order:    *order,
		Delivery: delivery,
		Payment:  payment,
		Items:    items,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Error("Error executing template", err)
		return
	}
}
