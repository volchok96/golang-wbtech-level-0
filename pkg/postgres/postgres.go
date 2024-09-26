package postgres

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/database"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/pkg/logger"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/go-playground/validator/v10"
)

type PostgresDB interface {
	InsertOrderToDB(ctx context.Context, order *models.Order) error
	GetOrderFromDB(ctx context.Context, orderID int) (*models.Order, error)
}

type PostgresDBImpl struct {
	Pool *pgxpool.Pool
	Log  logger.Logger
}

func NewPostgresDB(pool *pgxpool.Pool, log logger.Logger) *PostgresDBImpl {
	return &PostgresDBImpl{Pool: pool, Log: log}
}

func ConnectDB(log logger.Logger, config config.AppConfig) (*pgxpool.Pool, error) {
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

func (db *PostgresDBImpl) InsertOrderToDB(ctx context.Context, order *models.Order) error {
	validate := validator.New()
	err := validate.Struct(order)
	if err != nil {
		db.Log.Error("Validation failed", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		db.Log.Error("Error starting transaction", err)
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			db.Log.Error("Transaction rolled back due to panic", nil)
		}
	}()

	_, err = database.InsertDelivery(db.Log, tx, &order.Delivery)
	if err != nil {
		tx.Rollback(ctx)
		db.Log.Error("Error inserting delivery", err)
		return fmt.Errorf("error inserting delivery: %v", err)
	}

	_, err = database.InsertPayment(db.Log, tx, &order.Payment)
	if err != nil {
		tx.Rollback(ctx)
		db.Log.Error("Error inserting payment", err)
		return fmt.Errorf("error inserting payment: %v", err)
	}

	for _, item := range order.Items {
		err = database.InsertItem(db.Log, tx, &item)
		if err != nil {
			tx.Rollback(ctx)
			db.Log.Error("Error inserting item", err)
			return fmt.Errorf("error inserting item: %v", err)
		}
	}

	err = database.InsertOrder(db.Log, tx, order)
	if err != nil {
		tx.Rollback(ctx)
		db.Log.Error("Error inserting order", err)
		return fmt.Errorf("error inserting order: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		db.Log.Error("Error committing transaction", err)
		return fmt.Errorf("error committing transaction: %v", err)
	}

	db.Log.Info("Order successfully inserted")
	return nil
}

func (db *PostgresDBImpl) GetOrderFromDB(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	err := db.Pool.QueryRow(ctx, "SELECT id, order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE id = $1", orderID).Scan(
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
		db.Log.Error("Error getting order from DB", err)
		return nil, err
	}

	err = db.Pool.QueryRow(ctx, "SELECT id, name, phone, zip, city, address, region, email FROM delivery WHERE id = $1", order.Delivery.ID).Scan(
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
		db.Log.Error("Error getting delivery from DB", err)
		return nil, err
	}

	err = db.Pool.QueryRow(ctx, "SELECT id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payment WHERE id = $1", order.Payment.ID).Scan(
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
		db.Log.Error("Error getting payment from DB", err)
		return nil, err
	}

	rows, err := db.Pool.Query(ctx, "SELECT id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE track_number = $1", order.TrackNumber)
	if err != nil {
		db.Log.Error("Error getting items from DB", err)
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
			db.Log.Error("Error scanning item from DB", err)
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		db.Log.Error("Error iterating over items", err)
		return nil, err
	}

	db.Log.Info("Order successfully retrieved from DB")
	return &order, nil
}
