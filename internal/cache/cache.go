package cache

import (
	"encoding/json"
	"strconv"
	"wb-kafka-service/internal/models"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/rs/zerolog/log"
)

var MemCache *memcache.Client

func init() {
	MemCache = memcache.New("127.0.0.1:11211")
}

func SaveToCache(order *models.Order) error {
	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling order")
		return err
	}

	err = MemCache.Set(&memcache.Item{Key: "order:" + strconv.Itoa(order.ID), Value: orderData})
	if err != nil {
		log.Error().Err(err).Msg("Error saving order to memcache")
		return err
	}

	for _, item := range order.Items {
		err := MemCache.Set(&memcache.Item{Key: "item:" + strconv.Itoa(item.ID), Value: []byte(strconv.Itoa(item.ChrtID))})
		if err != nil {
			log.Error().Err(err).Msg("Error saving item to memcache")
			return err
		}
	}

	err = MemCache.Set(&memcache.Item{Key: "delivery:" + strconv.Itoa(order.Delivery.ID), Value: []byte(order.Delivery.Name)})
	if err != nil {
		log.Error().Err(err).Msg("Error saving delivery to memcache")
		return err
	}

	err = MemCache.Set(&memcache.Item{Key: "payment:" + strconv.Itoa(order.Payment.ID), Value: []byte(order.Payment.Transaction)})
	if err != nil {
		log.Error().Err(err).Msg("Error saving payment to memcache")
		return err
	}

	return nil
}
