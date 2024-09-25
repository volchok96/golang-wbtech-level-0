package postgres

import (
	"context"
	"fmt"
	"wb-kafka-service/internal/config"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBPoolIface interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

type PostgresDBImpl struct {
	db     DBPoolIface
	logger logger.Logger
}

func NewPostgresDB(db DBPoolIface, logger logger.Logger) *PostgresDBImpl {
	return &PostgresDBImpl{
		db:     db,
		logger: logger,
	}
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

func (db *PostgresDBImpl) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.db.Begin(ctx)
}

func (db *PostgresDBImpl) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.db.QueryRow(ctx, sql, args...)
}

func (db *PostgresDBImpl) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.db.Query(ctx, sql, args...)
}

func (db *PostgresDBImpl) InsertOrderToDB(ctx context.Context, order *models.Order) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		db.logger.Error("Error starting transaction", err)
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			db.logger.Error("Transaction rolled back due to panic", nil)
		}
	}()

	err = tx.Commit(ctx)
	if err != nil {
		db.logger.Error("Error committing transaction", err)
		return fmt.Errorf("error committing transaction: %v", err)
	}

	db.logger.Info("Order successfully inserted")
	return nil
}

func (db *PostgresDBImpl) GetOrderFromDB(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	err := db.QueryRow(ctx, "SELECT id, order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE id = $1", orderID).Scan(
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
		db.logger.Error("Error getting order from DB", err)
		return nil, err
	}

	db.logger.Info("Order successfully retrieved from DB")
	return &order, nil
}
