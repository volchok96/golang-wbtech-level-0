package postgres

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/db"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger" 

	"github.com/jackc/pgx/v4/pgxpool"
)

func ConnectToDB(config *config.AppConfig, log *logger.Logger) (*pgxpool.Pool, error) {
	pool, err := db.ConnectDB(log, *config)
	if err != nil {
		log.Error("Unable to connect to database", err)
		return nil, err
	}
	return pool, nil
}

func InsertOrderToDB(ctx context.Context, pool *pgxpool.Pool, order *models.Order, log *logger.Logger) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Error("Error starting transaction", err)
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			log.Error("Transaction rolled back due to panic", nil)
		}
	}()

	_, err = db.InsertDelivery(log, tx, &order.Delivery)
	if err != nil {
		tx.Rollback(ctx)
		log.Error("Error inserting delivery", err)
		return fmt.Errorf("error inserting delivery: %v", err)
	}

	_, err = db.InsertPayment(log, tx, &order.Payment)
	if err != nil {
		tx.Rollback(ctx)
		log.Error("Error inserting payment", err)
		return fmt.Errorf("error inserting payment: %v", err)
	}

	for _, item := range order.Items {
		err = db.InsertItem(log, tx, &item)  
		if err != nil {
			tx.Rollback(ctx)
			log.Error("Error inserting item", err)
			return fmt.Errorf("error inserting item: %v", err)
		}
	}

	err = db.InsertOrder(log, tx, order)
	if err != nil {
		tx.Rollback(ctx)
		log.Error("Error inserting order", err)
		return fmt.Errorf("error inserting order: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Error committing transaction", err)
		return fmt.Errorf("error committing transaction: %v", err)
	}

	log.Info("Order successfully inserted")
	return nil
}

func GetOrderFromDB(ctx context.Context, pool *pgxpool.Pool, orderID int, log *logger.Logger) (*models.Order, error) {
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
		log.Error("Error getting order from DB", err)
		return nil, err
	}

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
		log.Error("Error getting delivery from DB", err)
		return nil, err
	}

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
		log.Error("Error getting payment from DB", err)
		return nil, err
	}

	rows, err := pool.Query(ctx, "SELECT id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items")
	if err != nil {
		log.Error("Error getting items from DB", err)
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
			log.Error("Error scanning item from DB", err)
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating over items", err)
		return nil, err
	}

	log.Info("Order successfully retrieved from DB")
	return &order, nil
}
