package kafka

import (
	"context"
	"encoding/json"
	"os"
	"gopkg.in/yaml.v3"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/postgres"
	"wb-kafka-service/pkg/unmarshal"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var Cache = make(map[int]models.Order)

func InitKafka(cfg *config.AppConfig, pool *pgxpool.Pool) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Kafka.Broker},
		Topic:    cfg.Kafka.Topic,
		GroupID:  "orders-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Info().Msg("Kafka consumer initialized")

	// Читаем заказы из директории
	orders := unmarshal.ReadOrdersFromDirectory(".././materials")

	// Обрабатываем каждый заказ отдельно
	for _, order := range orders {
		// Сохраняем заказ в базу данных
		err := postgres.InsertOrderToDB(context.Background(), pool, &order)
		if err != nil {
			log.Error().Err(err).Msgf("Error inserting order into DB: %v", order.ID)
			continue
		}

		// Сохраняем заказ в кэш
		err = cache.SaveToCache(&order)
		if err != nil {
			log.Error().Err(err).Msgf("Error saving order to cache: %v", order.ID)
			continue
		}

		// Сохраняем в локальный кэш
		Cache[order.ID] = order

		log.Info().Msgf("Processed order from directory: %v", order)
	}

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Error reading message from Kafka")
			continue
		}

		// Unmarshal данных из Kafka сообщения
		var order models.Order
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshalling order")
			continue
		}

		// Сохраняем заказ в базу данных
		err = postgres.InsertOrderToDB(context.Background(), pool, &order)
		if err != nil {
			log.Error().Err(err).Msgf("Error inserting order into DB: %v", order.ID)
			continue
		}

		// Сохраняем заказ в кэш
		err = cache.SaveToCache(&order)
		if err != nil {
			log.Error().Err(err).Msgf("Error saving order to cache: %v", order.ID)
			continue
		}

		// Сохраняем в локальный кэш
		Cache[order.ID] = order

		log.Info().Msgf("Processed order from Kafka: %v", order)
	}
}

func ProduceOrder(cfg *config.AppConfig, order *models.Order) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{cfg.Kafka.Broker},
		Topic:   cfg.Kafka.Topic,
	})
	defer writer.Close()

	orderData, err := json.Marshal(order)
	if err != nil {
		log.Error().Err(err).Msg("Error marshalling order")
		return err
	}

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: orderData,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error writing message to Kafka")
		return err
	}

	log.Info().Msgf("Produced order: %v", order)
	return nil
}

func GetConfig() (*config.AppConfig, error) {
	var cfg config.AppConfig

	file, err := os.ReadFile(".././config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Error reading config file")
		return nil, err
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshalling config file")
		return nil, err
	}

	return &cfg, nil
}
