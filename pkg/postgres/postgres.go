package postgres

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/db"
	"wb-kafka-service/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

// ConnectToDB - функция для подключения к базе данных, возвращает ошибку, если не удается подключиться
func ConnectToDB(config *config.AppConfig) (*pgxpool.Pool, error) {
	pool, err := db.ConnectDB(*config)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to connect to database: %v", err)
		return nil, err
	}
	return pool, nil
}

// InsertOrderToDB - вставка заказа в базу данных с обработкой ошибок
func InsertOrderToDB(ctx context.Context, pool *pgxpool.Pool, order *models.Order) error {
	tx, err := pool.Begin(ctx)
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
	_, err = db.InsertDelivery(tx, &order.Delivery)
	if err != nil {
		tx.Rollback(ctx)
		log.Error().Err(err).Msg("Error inserting delivery")
		return fmt.Errorf("error inserting delivery: %v", err)
	}

	// Вставка данных об оплате
	_, err = db.InsertPayment(tx, &order.Payment)
	if err != nil {
		tx.Rollback(ctx)
		log.Error().Err(err).Msg("Error inserting payment")
		return fmt.Errorf("error inserting payment: %v", err)
	}

	// Вставка позиций заказа
	for _, item := range order.Items {
		err = db.InsertItem(tx, &item)
		if err != nil {
			tx.Rollback(ctx)
			log.Error().Err(err).Msg("Error inserting item")
			return fmt.Errorf("error inserting item: %v", err)
		}
	}

	// Вставка самого заказа
	err = db.InsertOrder(tx, order)
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
func GetOrderFromDB(ctx context.Context, pool *pgxpool.Pool, orderID int) (*models.Order, error) {
	var order models.Order
	err := pool.QueryRow(ctx, "SELECT id, order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE id = $1", orderID).Scan(
		&order.ID,
		&order.OrderUid,
		&order.TrackNumber,
		&order.Entry,
		&order.Delivery.ID,
		&order.Payment.ID,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting order from DB: %v", err)
		return nil, err
	}

	// Читаем данные о доставке
	err = pool.QueryRow(ctx, "SELECT id, name, phone, zip, city, address, region, email FROM delivery WHERE id = $1", order.Delivery.ID).Scan(
		&order.Delivery.ID,
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting delivery from DB: %v", err)
		return nil, err
	}

	// Читаем данные об оплате
	err = pool.QueryRow(ctx, "SELECT id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE id = $1", order.Payment.ID).Scan(
		&order.Payment.ID,
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting payment from DB: %v", err)
		return nil, err
	}

	// Читаем позиции заказа
	rows, err := pool.Query(ctx, "SELECT id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items")
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
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
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
