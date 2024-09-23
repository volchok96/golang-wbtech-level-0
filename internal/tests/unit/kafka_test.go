package tests

import (
	"testing"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/kafka"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"

	"github.com/golang/mock/gomock"
)

func TestProduceOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для логгера
	mockLogger := logger.NewMockLogger(ctrl)

	// Пример заказа
	order := &models.Order{
		ID:       1,
		OrderUid: "test-uid",
		Delivery: models.Delivery{ID: 1, Name: "John Doe"},
		Payment:  models.Payment{ID: 1, Transaction: "trans123"},
		Items:    []models.Items{{ID: 1, Name: "item1", Price: 100}},
	}

	// Настраиваем ожидания для логгера
	mockLogger.EXPECT().Info(gomock.Any()).Times(1)

	// Создаем конфигурацию Kafka
	cfg := config.AppConfig{
		Kafka: struct {
			Broker  string `yaml:"broker"`
			GroupID string `yaml:"group_id"`
			Topic   string `yaml:"topic"`
		}{
			Broker:  "localhost:9092",
			GroupID: "order-group",
			Topic:   "test-topic",
		},
	}

	// Вызываем функцию ProduceOrder
	err := kafka.ProduceOrder(cfg, order, mockLogger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestInitKafka(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки для логгера, базы данных и кэша
	mockLogger := logger.NewMockLogger(ctrl)
	mockDB := postgres.NewMockPostgresDB(ctrl)
	mockCache := cache.NewMockMemCacheClient(ctrl)

	// Пример заказа
	order := models.Order{
		ID:       1,
		OrderUid: "test-uid",
		Delivery: models.Delivery{ID: 1, Name: "New New"},
		Payment:  models.Payment{ID: 1, Transaction: "t123"},
		Items:    []models.Items{{ID: 1, Name: "item1", Price: 100}},
	}

	// Настраиваем ожидания для базы данных и кэша
	mockDB.EXPECT().InsertOrderToDB(gomock.Any(), &order).Return(nil).Times(1)

	// Создаем конфигурацию Kafka
	cfg := config.AppConfig{
		Kafka: struct {
			Broker  string `yaml:"broker"`
			GroupID string `yaml:"group_id"`
			Topic   string `yaml:"topic"`
		}{
			Broker:  "localhost:9092",
			GroupID: "order-group",
			Topic:   "test-topic",
		},
	}

	// Вызов функции InitKafka (это будет блокирующий вызов, нужно рассмотреть использование горутин для тестирования)
	go kafka.InitKafka(cfg, mockDB, mockLogger, mockCache)
}
