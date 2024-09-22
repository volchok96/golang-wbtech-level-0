package db

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func ConnectDB(log *logger.Logger, config config.AppConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host, config.Postgres.Port, config.Postgres.User, config.Postgres.Password, config.Postgres.DBName)
	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Error("Failed to connect to the database", err)
		return nil, err
	}

	log.Info("Successfully connected to the database")
	return pool, nil
}

func InsertDelivery(log *logger.Logger, tx pgx.Tx, delivery *models.Delivery) (int, error) {
	var id int
	// Проверка существования записи доставки перед вставкой
	err := tx.QueryRow(context.Background(),
		"SELECT id FROM delivery WHERE name=$1 AND phone=$2 AND zip=$3 AND city=$4 AND address=$5 AND region=$6 AND email=$7",
		delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email).Scan(&id)

	if err == pgx.ErrNoRows {
		// Если запись не найдена, выполняем вставку новой
		err = tx.QueryRow(context.Background(),
			"INSERT INTO delivery (name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email).Scan(&id)

		if err != nil {
			log.Error("Failed to insert delivery", err)
			return 0, err
		}

		log.Info(fmt.Sprintf("Inserted delivery with ID: %d", id))
	} else if err != nil {
		log.Error("Failed to check delivery existence", err)
		return 0, err
	} else {
		log.Info(fmt.Sprintf("Delivery already exists with ID: %d", id))
	}

	delivery.ID = id
	return id, nil
}

func InsertPayment(log *logger.Logger, tx pgx.Tx, payment *models.Payment) (int, error) {
	var id int
	// Проверка существования платежа перед вставкой
	err := tx.QueryRow(context.Background(),
		"SELECT id FROM payment WHERE transaction=$1 AND request_id=$2 AND amount=$3",
		payment.Transaction, payment.RequestID, payment.Amount).Scan(&id)

	if err == pgx.ErrNoRows {
		// Если запись не найдена, выполняем вставку новой
		err = tx.QueryRow(context.Background(),
			"INSERT INTO payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
			payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee).Scan(&id)

		if err != nil {
			log.Error("Failed to insert payment", err)
			return 0, err
		}

		log.Info(fmt.Sprintf("Inserted payment with ID: %d", id))
	} else if err != nil {
		log.Error("Failed to check payment existence", err)
		return 0, err
	} else {
		log.Info(fmt.Sprintf("Payment already exists with ID: %d", id))
	}

	payment.ID = id
	return id, nil
}

func InsertItem(log *logger.Logger, tx pgx.Tx, item *models.Items) error {
	// Попытка получить ID элемента, если он уже существует в базе данных
	err := tx.QueryRow(context.Background(),
		`SELECT id FROM items WHERE chrt_id = $1 AND track_number = $2 AND price = $3`,
		item.ChrtID, item.TrackNumber, item.Price).Scan(&item.ID)

	if err == pgx.ErrNoRows {
		// Если элемент не найден, выполняем вставку нового элемента
		err = tx.QueryRow(context.Background(),
			`INSERT INTO items (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status).Scan(&item.ID)

		if err != nil {
			log.Error("Failed to insert item", err)
			return err
		}
		log.Info(fmt.Sprintf("Inserted new item with ID: %d", item.ID))
	} else if err != nil {
		log.Error("Failed to check item existence", err)
		return err
	} else {
		log.Info(fmt.Sprintf("Item already exists with ID: %d", item.ID))
	}

	return nil
}

func InsertOrder(log *logger.Logger, tx pgx.Tx, order *models.Order) error {
	var id int
	// Проверка существования заказа перед вставкой
	err := tx.QueryRow(context.Background(),
		"SELECT id FROM orders WHERE order_uid = $1", order.OrderUid).Scan(&id)

	if err == pgx.ErrNoRows {
		// Если заказ не найден, вставляем новый
		_, err = tx.Exec(context.Background(),
			"INSERT INTO orders (order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
			order.OrderUid, order.TrackNumber, order.Entry, order.Delivery.ID, order.Payment.ID, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)

		if err != nil {
			log.Error("Failed to insert order", err)
			return err
		}

		log.Info("Inserted order successfully")
	} else if err != nil {
		log.Error("Failed to check order existence", err)
		return err
	} else {
		log.Info(fmt.Sprintf("Order already exists with ID: %d", id))
	}

	return nil
}
