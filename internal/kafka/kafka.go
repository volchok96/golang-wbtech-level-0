package kafka

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/postgres"
	"github.com/jackc/pgx/v4"
)

var Cache = make(map[int]models.Order)

func InitKafka(config *config.AppConfig, conn *pgx.Conn) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{config.Kafka.Broker},
		Topic:    config.Kafka.Topic,
		GroupID:  "order-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Info().Msg("Kafka consumer initialized")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Error reading message from Kafka")
			continue
		}

		var order models.Order
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			log.Error().Err(err).Msg("Error unmarshalling order")
			continue
		}

		postgres.InsertOrderToDB(context.Background(), conn, &order)
		Cache[order.ID] = order
		log.Info().Msgf("Received order: %v", order)
	}
}

