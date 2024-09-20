package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"wb-nats-service/internal/cache"
	"wb-nats-service/internal/config"
	"wb-nats-service/internal/models"
	"wb-nats-service/pkg/postgres"
	"wb-nats-service/pkg/unmarshal"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

func Nats() {
	config, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get config")
	}

	clusterID := config.Nats.ClusterID
	clientID := config.Nats.ClientID

	orders := unmarshal.ReadOrdersFromDirectory(".././materials")

	conn := postgres.ConnectToDB(&config)
	defer conn.Close(context.Background())

	natsconn := ConnectToNats(clusterID, clientID)
	if natsconn == nil {
		log.Error().Msg("Failed to connect to NATS")
		return
	}
	defer natsconn.Close()

	for _, order := range orders {
		postgres.InsertOrderToDB(conn, &order)

		err := cache.SaveToCache(&order)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save cache")
		}

		PublishOrder(natsconn, &order)
	}

	SubscribeToOrder(natsconn)

}

func ConnectToNats(clusterID, clientID string) *nats.Conn {
	natsconn, err := nats.Connect(clusterID)
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to NATS")
	}

	return natsconn
}

func PublishOrder(natsconn *nats.Conn, order *models.Order) {
	if natsconn == nil {
		log.Error().Msg("NATS connection is nil")
		return
	}

	orderData, err := json.Marshal(order)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling order")
	}

	err = natsconn.Publish("order", orderData)
	if err != nil {
		log.Fatal().Err(err).Msg("Error publishing to channel")
	} else {
		log.Info().Msg("Successfully published to channel")
	}
}

func SubscribeToOrder(natsconn *nats.Conn) {
	if natsconn == nil {
		log.Error().Msg("NATS connection is nil")
		return
	}

	sub, err := natsconn.Subscribe("order", func(m *nats.Msg) {
		fmt.Printf("Received a message from channel: %s\n", string(m.Data))
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Error subscribing to channel")
	} else {
		log.Info().Msg("Successfully subscribed to channel")
	}

	defer sub.Unsubscribe()
}
