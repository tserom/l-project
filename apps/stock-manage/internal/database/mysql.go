package database

import (
	"fmt"
	"log"
	"time"

	"github.com/tserom/l-project/apps/stock-manage/internal/config"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMySQL opens a GORM connection and runs auto-migration for business models.
func NewMySQL(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Info
	if cfg.AppEnv == "production" {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(
		&model.StockOperationLog{},
		&model.DocSequence{},
		&model.InboundOrder{},
		&model.InboundOrderLine{},
		&model.OutboundOrder{},
		&model.OutboundOrderLine{},
		&model.SalesOrder{},
		&model.SalesOrderLine{},
		&model.SalesShipment{},
		&model.SalesShipmentLine{},
		&model.ProcessingOrder{},
		&model.ProcessingPickLine{},
		&model.ProcessingFinishLine{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	log.Println("business database connected")
	return db, nil
}
