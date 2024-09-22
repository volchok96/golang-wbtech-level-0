package cache

import (
	"encoding/json"
	"strconv"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger" 
	"github.com/bradfitz/gomemcache/memcache"
)

var MemCache *memcache.Client

func init() {
	MemCache = memcache.New("127.0.0.1:11211")
}

func SaveToCache(log *logger.Logger, order *models.Order) error {
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error("Error marshalling order", err)
		return err
	}

	err = MemCache.Set(&memcache.Item{Key: "order:" + strconv.Itoa(order.ID), Value: orderData})
	if err != nil {
		log.Error("Error saving order to memcache", err)
		return err
	}

	for _, item := range order.Items {
		err := MemCache.Set(&memcache.Item{Key: "item:" + strconv.Itoa(item.ID), Value: []byte(strconv.Itoa(item.ChrtID))})
		if err != nil {
			log.Error("Error saving item to memcache", err)
			return err
		}
	}

	err = MemCache.Set(&memcache.Item{Key: "delivery:" + strconv.Itoa(order.Delivery.ID), Value: []byte(order.Delivery.Name)})
	if err != nil {
		log.Error("Error saving delivery to memcache", err)
		return err
	}

	err = MemCache.Set(&memcache.Item{Key: "payment:" + strconv.Itoa(order.Payment.ID), Value: []byte(order.Payment.Transaction)})
	if err != nil {
		log.Error("Error saving payment to memcache", err)
		return err
	}

	return nil
}
