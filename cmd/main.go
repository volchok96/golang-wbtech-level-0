package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"net/http"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/handlers"
	"wb-kafka-service/internal/kafka"
	"wb-kafka-service/pkg/postgres"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config")
	}

	conn := postgres.ConnectToDB(&config)
	defer conn.Close(context.Background())

	kafka.InitKafka(&config, conn)

	http.HandleFunc("/order", handlers.HandlerOrder)
	http.ListenAndServe(":8082", nil)
}
