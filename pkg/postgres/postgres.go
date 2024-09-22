package postgres

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/db"
	"wb-kafka-service/internal/models"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log" // Используем zerolog для логирования
)

// ConnectToDB - функция для подключения к базе данных, возвращает ошибку, если не удается подключиться
func ConnectToDB(config *config.AppConfig) (*pgx.Conn, error) {
	conn, err := db.ConnectDB(*config)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to connect to database: %v", err)
		return nil, err
	}
	return conn, nil
}

// InsertOrderToDB - вставка заказа в базу данных с обработкой ошибок
func InsertOrderToDB(ctx context.Context, conn *pgx.Conn, order *models.Order) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error starting transaction")
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		// Откат транзакции, если произошла ошибка
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			log.Error().Msg("Transaction rolled back due to panic")
		}
	}()

	// Вставка данных о доставке
	err = db.InsertDelivery(ctx, tx, &order.Delivery)
	if err != nil {
		tx.Rollback(ctx)
		log.Error().Err(err).Msg("Error inserting delivery")
		return fmt.Errorf("error inserting delivery: %v", err)
	}

	// Вставка данных об оплате
	err = db.InsertPayment(ctx, tx, &order.Payment)
	if err != nil {
		tx.Rollback(ctx)
		log.Error().Err(err).Msg("Error inserting payment")
		return fmt.Errorf("error inserting payment: %v", err)
	}

	// Вставка позиций заказа
	for _, item := range order.Items {
		err = db.InsertItem(ctx, tx, &item)
		if err != nil {
			tx.Rollback(ctx)
			log.Error().Err(err).Msg("Error inserting item")
			return fmt.Errorf("error inserting item: %v", err)
		}
	}

	// Вставка самого заказа
	err = db.InsertOrder(ctx, tx, order)
	if err != nil {
		tx.Rollback(ctx)
		log.Error().Err(err).Msg("Error inserting order")
		return fmt.Errorf("error inserting order: %v", err)
	}

	// Коммит транзакции
	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error committing transaction")
		return fmt.Errorf("error committing transaction: %v", err)
	}

	log.Info().Msgf("Order successfully inserted: %v", order)
	return nil
}

// GetOrderFromDB - чтение заказа из базы данных по ID
func GetOrderFromDB(ctx context.Context, conn *pgx.Conn, orderID int) (*models.Order, error) {
	var order models.Order
	err := conn.QueryRow(ctx, "SELECT id, user_id, product, quantity, price FROM orders WHERE id = $1", orderID).Scan(
		&order.ID,
		&order.OrderUid,
		&order.Entry,
		&order.Items,
		&order.Payment,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting order from DB: %v", err)
		return nil, err
	}

	// Читаем данные о доставке
	err = conn.QueryRow(ctx, "SELECT id, address, city, country FROM deliveries WHERE order_id = $1", orderID).Scan(
		&order.Delivery.ID,
		&order.Delivery.Address,
		&order.Delivery.City,
		&order.Delivery.Region,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting delivery from DB: %v", err)
		return nil, err
	}

	// Читаем данные об оплате
	err = conn.QueryRow(ctx, "SELECT id, method, amount FROM payments WHERE order_id = $1", orderID).Scan(
		&order.Payment.ID,
		&order.Payment.Currency,
		&order.Payment.Amount,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting payment from DB: %v", err)
		return nil, err
	}

	// Читаем позиции заказа
	rows, err := conn.Query(ctx, "SELECT id, product_id, quantity, price FROM items WHERE order_id = $1", orderID)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting items from DB: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Items
		err = rows.Scan(
			&item.ID,
			&item.ChrtID,
			&item.Size,
			&item.TotalPrice,
		)
		if err != nil {
			log.Error().Err(err).Msgf("Error scanning item from DB: %v", err)
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msgf("Error iterating over items: %v", err)
		return nil, err
	}

	log.Info().Msgf("Order successfully retrieved from DB: %v", order)
	return &order, nil
}
