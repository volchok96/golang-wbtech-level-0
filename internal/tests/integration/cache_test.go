package tests

import (
	"encoding/json"
	"testing"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/pkg/logger"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSaveToCache_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := cache.NewMockMemCacheClient(ctrl)
	mockLogger := logger.NewMockLogger(ctrl) // Используем новый мок для интерфейса Logger

	order := &models.Order{
		ID:       1,
		OrderUid: "test-uid",
		Delivery: models.Delivery{ID: 1, Name: "John Doe"},
		Payment:  models.Payment{ID: 1, Transaction: "trans123"},
		Items:    []models.Items{{ID: 1, Name: "item1", Price: 100}},
	}

	orderData, _ := json.Marshal(order)

	mockCache.EXPECT().Set(&memcache.Item{Key: "order:1", Value: orderData}).Return(nil)
	mockCache.EXPECT().Set(&memcache.Item{Key: "item:1", Value: []byte("0")}).Return(nil)
	mockCache.EXPECT().Set(&memcache.Item{Key: "delivery:1", Value: []byte("John Doe")}).Return(nil)
	mockCache.EXPECT().Set(&memcache.Item{Key: "payment:1", Value: []byte("trans123")}).Return(nil)

	err := cache.SaveToCache(mockLogger, mockCache, order)

	assert.NoError(t, err)
}
