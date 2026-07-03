package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// StockBalance holds dual-measure inventory for material + batch + warehouse.
type StockBalance struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID uint64          `gorm:"not null;uniqueIndex:uk_balance;index" json:"materialId"`
	BatchID    uint64          `gorm:"not null;uniqueIndex:uk_balance;index" json:"batchId"`
	Warehouse  string          `gorm:"size:64;not null;uniqueIndex:uk_balance" json:"warehouse"`
	WeightKg   decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"weightKg"`
	Quantity   decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
	OrgID      uint64          `gorm:"not null;default:0;uniqueIndex:uk_balance;index" json:"orgId"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func (StockBalance) TableName() string { return "stock_balance" }
