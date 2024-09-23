package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"wb-kafka-service/internal/handlers"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки для MemCache, Logger и PostgresDB
	mockCache := cache.NewMockMemCacheClient(ctrl)  // Мок для кэша
	mockLogger := logger.NewMockLogger(ctrl)        // Мок для логгера
	mockDB := postgres.NewMockPostgresDB(ctrl)     // Мок для PostgresDB

	// Пример заказа
	order := models.Order{
		ID:       1,
		OrderUid: "test-uid",
		Delivery: models.Delivery{ID: 1, Name: "John Doe"},
		Payment:  models.Payment{ID: 1, Transaction: "trans123"},
		Items:    []models.Items{{ID: 1, Name: "item1", Price: 100}},
	}

	// Мокируем возвращаемые данные из кэша
	orderData, _ := json.Marshal(order)
	mockCache.EXPECT().Get("order:" + strconv.Itoa(order.ID)).Return(&memcache.Item{Value: orderData}, nil)

	// Мокируем возвращение заказа из базы данных
	mockDB.EXPECT().GetOrderFromDB(gomock.Any(), order.ID).Return(&order, nil)

	// Настройка логгера
	mockLogger.EXPECT().Info(gomock.Any()).Times(1) // Пример вызова метода Info для логгирования

	// Создаем запрос
	req, err := http.NewRequest("GET", "/order?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем рекордер для записи HTTP-ответа
	rr := httptest.NewRecorder()

	// Вызов обработчика
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerOrder(mockLogger, mockCache, mockDB, w, r) // Используем моки логгера, кэша и базы данных
	})
	handler.ServeHTTP(rr, req)

	// Проверяем, что статус-код успешный
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем, что тело ответа содержит ожидаемые данные
	assert.Contains(t, rr.Body.String(), "John Doe")
}
