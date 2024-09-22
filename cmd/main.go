package main

import (
	"context"
	"net/http"

	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/handlers"
	"wb-kafka-service/internal/kafka"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"
)

func main() {
	// Инициализация логгера с выводом в файл и консоль
	log, err := logger.NewLogger("app.log", true) // true - логирование и в консоль, и в файл
	if err != nil {
		log.Fatal("Failed to create logger", err)
	}
	defer log.Close()

	// Получаем конфигурацию
	config, err := config.GetConfig(log)
	if err != nil {
		log.Fatal("Failed to get config", err)
	}

	// Подключение к базе данных
	pool, err := postgres.ConnectToDB(&config, log)
	if err != nil {
		log.Fatal("Failed to connect to DB", err)
	}
	defer pool.Close()

	// Запускаем Kafka-консюмера в отдельной горутине
	go func() {
		log.Info("Starting Kafka consumer...")
		kafka.InitKafka(&config, pool, log)
	}()

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerOrder(log, w, r)
	})

	log.Info("Starting HTTP server on :8080")
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP server failed", err)
		}
	}()

	// Читаем существующий заказ из базы данных
	order, err := postgres.GetOrderFromDB(context.Background(), pool, 1, log) // Предполагаем, что заказ с ID 1 существует
	if err != nil {
		log.Error("Failed to get order from DB", err)
	} else {
		// Отправляем заказ в Kafka
		err = kafka.ProduceOrder(&config, order, log)
		if err != nil {
			log.Error("Failed to produce order", err)
		} else {
			log.Info("Order produced successfully")
		}
	}

	// Бесконечный цикл для поддержания работы основного потока
	select {}
}
