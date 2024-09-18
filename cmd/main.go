package main

import (
	"net/http"
	"wb-nats-service/internal/nats"
	"wb-nats-service/internal/handlers"

)

func main() {
	http.HandleFunc("/order", handler.HandlerOrder)

	nats.Nats()

	http.ListenAndServe(":8082", nil)
}
