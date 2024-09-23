package tests

import (
	"context"
	"testing"
	"wb-kafka-service/internal/models"
	"wb-kafka-service/pkg/logger"
	"wb-kafka-service/pkg/postgres"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInsertOrderToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl) 

	pgDB := &postgres.PostgresDBImpl{
		Pool: nil, 
		Log:  mockLogger, 
	}

	order := models.Order{
		OrderUid: "test-uid",
		Delivery: models.Delivery{ID: 1},
		Payment:  models.Payment{ID: 1},
		Items:    []models.Items{{ID: 1, Name: "item1", Price: 100}},
	}

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	mockDB := postgres.NewMockPostgresDB(ctrl)
	mockDB.EXPECT().InsertOrderToDB(gomock.Any(), &order).Return(nil)

	err := pgDB.InsertOrderToDB(context.Background(), &order)
	assert.NoError(t, err)
}
