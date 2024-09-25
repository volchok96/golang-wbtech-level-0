package main

import (
	"context"
	"net/http"
	"wb-kafka-service/internal/cache"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/handlers"
	"wb-kafka-service/internal/kafka"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"
)

func main() {
	// Initialize the logger
	log, err := logger.NewLogger("app.log", true)
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer log.Close()

	// Get configuration
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

	memCacheClient := cache.NewMemCache("127.0.0.1:11211")

	go func() {
		log.Info("Starting Kafka consumer...")
		kafka.InitKafka(cfg, postgresDB, log, memCacheClient)
	}()

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerOrder(log, memCacheClient, postgresDB, w, r)
	})

	log.Info("Starting HTTP server on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP server failed", err)
		}
	}()

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

	select {}
}
