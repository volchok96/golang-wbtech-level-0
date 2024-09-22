package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/pkg/logger" 
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/postgres"
	"wb-kafka-service/pkg/unmarshal"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/segmentio/kafka-go"
)

var Cache = make(map[int]models.Order)

func InitKafka(cfg *config.AppConfig, pool *pgxpool.Pool, log *logger.Logger) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Kafka.Broker},
		Topic:    cfg.Kafka.Topic,
		GroupID:  "order-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Info("Kafka consumer initialized")

	// Читаем заказы из директории
	orders := unmarshal.ReadOrdersFromDirectory(log, ".././materials")

	// Обрабатываем каждый заказ отдельно
	for _, order := range orders {
		// Сохраняем заказ в базу данных
		err := postgres.InsertOrderToDB(context.Background(), pool, &order, log)  
		if err != nil {
			log.Error(fmt.Sprintf("Error inserting order into DB: %v", order.ID), err)
			continue
		}

		// Сохраняем заказ в кэш
		err = cache.SaveToCache(log, &order)  
		if err != nil {
			log.Error(fmt.Sprintf("Error saving order to cache: %v", order.ID), err)
			continue
		}

		// Сохраняем в локальный кэш
		Cache[order.ID] = order

		log.Info(fmt.Sprintf("Processed order from directory: %v", order))
	}

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Error("Error reading message from Kafka", err)
			continue
		}

		// Unmarshal данных из Kafka сообщения
		var order models.Order
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Error("Error unmarshalling order", err)
			continue
		}

		// Сохраняем заказ в базу данных
		err = postgres.InsertOrderToDB(context.Background(), pool, &order, log)
		if err != nil {
			log.Error(fmt.Sprintf("Error inserting order into DB: %v", order.ID), err)
			continue
		}

		// Сохраняем заказ в кэш
		err = cache.SaveToCache(log, &order)
		if err != nil {
			log.Error(fmt.Sprintf("Error saving order to cache: %v", order.ID), err)
			continue
		}

		// Сохраняем в локальный кэш
		Cache[order.ID] = order

		log.Info(fmt.Sprintf("Processed order from Kafka: %v", order))
	}
}

func ProduceOrder(cfg *config.AppConfig, order *models.Order, log *logger.Logger) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.Topic,
	})
	defer writer.Close()

	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error("Error marshalling order", err)
		return err
	}

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: orderData,
	})
	if err != nil {
		log.Error("Error writing message to Kafka", err)
		return err
	}

	log.Info(fmt.Sprintf("Produced order: %v", order))
	return nil
}
