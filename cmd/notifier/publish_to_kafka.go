package main

import (
	"context"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/kafka"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"
)

func main() {
	log, err := logger.NewLogger("notify.log", true)
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer log.Close()

	cfg, err := config.GetConfig(log)
	if err != nil {
		log.Fatal("Failed to get config", err)
	}

	pool, err := postgres.ConnectDB(log, cfg)
	if err != nil {
		log.Fatal("Failed to connect to DB", err)
	}
	defer pool.Close()

	postgresDB := postgres.NewPostgresDB(pool, log)

	order, err := postgresDB.GetOrderFromDB(context.Background(), 1)
	if err != nil {
		log.Error("Failed to get order from DB", err)
	} else {
		err = kafka.ProduceOrder(cfg, order, log)
		if err != nil {
			log.Error("Failed to produce order to Kafka", err)
		} else {
			log.Info("Order produced successfully")
		}
	}
}
