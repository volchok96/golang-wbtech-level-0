package main

import (
	"net/http"
	"time"
	"os"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"wb-nats-service/internal/nats"
	"wb-nats-service/internal/handlers"

)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting service")
	http.HandleFunc("/order", handler.HandlerOrder)
	nats.Nats()

	log.Info().Msg("Starting HTTP server on :8082")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal().Err(err).Msg("HTTP server error")
	}
}
