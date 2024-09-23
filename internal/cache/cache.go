package cache

import (
	"encoding/json"
	"strconv"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemCacheClient - интерфейс для работы с Memcache
type MemCacheClient interface {
	Set(item *memcache.Item) error
	Get(key string) (*memcache.Item, error)
	Delete(key string) error
}

// MemCache - структура, реализующая интерфейс MemCacheClient
type MemCache struct {
	Client *memcache.Client
}

// NewMemCache - конструктор для создания нового экземпляра MemCache
func NewMemCache(server string) *MemCache {
	return &MemCache{Client: memcache.New(server)}
}

// Set - метод для установки значения в кэш
func (m *MemCache) Set(item *memcache.Item) error {
	return m.Client.Set(item)
}

// Get - метод для получения значения из кэша
func (m *MemCache) Get(key string) (*memcache.Item, error) {
	return m.Client.Get(key)
}

// Delete - метод для удаления значения из кэша
func (m *MemCache) Delete(key string) error {
	return m.Client.Delete(key)
}

// SaveToCache - функция для сохранения заказа в кэш
func SaveToCache(log logger.Logger, memCache MemCacheClient, order *models.Order) error {
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error("Error marshalling order", err)
		return err
	}

	err = memCache.Set(&memcache.Item{Key: "order:" + strconv.Itoa(order.ID), Value: orderData})
	if err != nil {
		log.Error("Error saving order to memcache", err)
		return err
	}

	for _, item := range order.Items {
		err := memCache.Set(&memcache.Item{Key: "item:" + strconv.Itoa(item.ID), Value: []byte(strconv.Itoa(item.ChrtID))})
		if err != nil {
			log.Error("Error saving item to memcache", err)
			return err
		}
	}

	err = memCache.Set(&memcache.Item{Key: "delivery:" + strconv.Itoa(order.Delivery.ID), Value: []byte(order.Delivery.Name)})
	if err != nil {
		log.Error("Error saving delivery to memcache", err)
		return err
	}

	err = memCache.Set(&memcache.Item{Key: "payment:" + strconv.Itoa(order.Payment.ID), Value: []byte(order.Payment.Transaction)})
	if err != nil {
		log.Error("Error saving payment to memcache", err)
		return err
	}

	return nil
}
