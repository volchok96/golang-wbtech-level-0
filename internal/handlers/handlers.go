package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-playground/validator/v10" 
)

type OrderPageData struct {
	Order    models.Order
	Delivery models.Delivery
	Payment  models.Payment
	Items    []models.Items
}

var validate = validator.New()

func HandlerOrder(log logger.Logger, cacheClient cache.MemCacheClient, db postgres.PostgresDB, w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		log.Error("Invalid order ID", err)
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	orderItem, err := cacheClient.Get("order:" + strconv.Itoa(orderID))
	if err == nil {
		order := models.Order{}
		err = json.Unmarshal(orderItem.Value, &order)
		if err != nil {
			log.Error("Error unmarshalling order", err)
			http.Error(w, "Error unmarshalling order", http.StatusInternalServerError)
			return
		}

		if err := validateOrder(&order, log, w); err != nil {
			return
		}

		renderOrderPage(w, order, log)
		return
	}

	order, err := db.GetOrderFromDB(context.Background(), orderID)
	if err != nil {
		log.Warn("Order not found", nil)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if err := validateOrder(order, log, w); err != nil {
		return
	}

	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error("Error marshalling order", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = cacheClient.Set(&memcache.Item{Key: "order:" + strconv.Itoa(orderID), Value: orderData})
	if err != nil {
		log.Error("Error saving order to cache", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	renderOrderPage(w, *order, log)
}

func validateOrder(order *models.Order, log logger.Logger, w http.ResponseWriter) error {
	if err := validate.Struct(order); err != nil {
		log.Error("Order validation failed", err)
		http.Error(w, "Invalid order data", http.StatusBadRequest)
		return err
	}
	return nil
}

func renderOrderPage(w http.ResponseWriter, order models.Order, log logger.Logger) {
	delivery := order.Delivery
	payment := order.Payment
	items := order.Items

	for i, item := range items {
		item.ID = i + 1
		items[i] = item
	}

	tmpl := template.Must(template.ParseFiles("../.././ui.html"))
	data := OrderPageData{
		Order:    order,
		Delivery: delivery,
		Payment:  payment,
		Items:    items,
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Error("Error executing template", err)
	}
}
