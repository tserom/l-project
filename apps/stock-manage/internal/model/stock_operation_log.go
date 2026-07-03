package model

import "time"

// StockOperationLog records business-side stock operations for audit.
type StockOperationLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SKU       string    `gorm:"size:64;not null;index" json:"sku"`
	Warehouse string    `gorm:"size:64;not null" json:"warehouse"`
	Action    string    `gorm:"size:32;not null" json:"action"`
	Operator  string    `gorm:"size:64;not null" json:"operator"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName overrides the default GORM table name.
func (StockOperationLog) TableName() string {
	return "stock_operation_log"
}
