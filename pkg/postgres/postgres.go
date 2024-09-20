package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/db"
	"wb-kafka-service/internal/models"

	"github.com/jackc/pgx/v4"
)

func ConnectToDB(config *config.AppConfig) *pgx.Conn {
	conn, err := db.ConnectDB(*config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return conn
}

func InsertOrderToDB(ctx context.Context, conn *pgx.Conn, order *models.Order) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	err = db.InsertDelivery(ctx, tx, &order.Delivery)
	if err != nil {
		log.Fatalf("Error inserting delivery: %v", err)
	}

	err = db.InsertPayment(ctx, tx, &order.Payment)
	if err != nil {
		log.Fatalf("Error inserting payment: %v", err)
	}

	for _, item := range order.Items {
		err = db.InsertItem(ctx, tx, &item)
		if err != nil {
			log.Fatalf("Error inserting item: %v", err)
		}
	}

	err = db.InsertOrder(ctx, tx, order)
	if err != nil {
		log.Fatalf("Error inserting order: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Fatalf("Error committing transaction: %v", err)
	}
}