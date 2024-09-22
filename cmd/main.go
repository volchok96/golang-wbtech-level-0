package main

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"

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

	pool, err := postgres.ConnectToDB(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to DB")
	}
	defer pool.Close()

	// Запускаем Kafka-консюмера в отдельной горутине
	go func() {
		kafka.InitKafka(&config, pool)
	}()

	// Запускаем HTTP-сервер для обработки запросов
	http.HandleFunc("/order", handlers.HandlerOrder)
	log.Info().Msg("Starting HTTP server on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Читаем существующий заказ из базы данных
	order, err := postgres.GetOrderFromDB(context.Background(), pool, 1) // Предполагаем, что заказ с ID 1 существует
	if err != nil {
		log.Error().Err(err).Msg("Failed to get order from DB")
	} else {
		// Отправляем заказ в Kafka
		err = kafka.ProduceOrder(&config, order)
		if err != nil {
			log.Error().Err(err).Msg("Failed to produce order")
		}
	}

	// Бесконечный цикл для поддержания работы основного потока
	select {}
}
