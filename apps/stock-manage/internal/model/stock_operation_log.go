package model

import "time"

// StockOperationLog records business-side stock operations for audit.
type StockOperationLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	DocType   string    `gorm:"size:32;index" json:"docType,omitempty"`
	DocNo     string    `gorm:"size:32;index" json:"docNo,omitempty"`
	SKU       string    `gorm:"size:64;index" json:"sku,omitempty"`
	Warehouse string    `gorm:"size:64" json:"warehouse,omitempty"`
	Action    string    `gorm:"size:32;not null" json:"action"`
	Operator  string    `gorm:"size:64;not null" json:"operator"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
}

// TableName overrides the default GORM table name.
func (StockOperationLog) TableName() string {
	return "stock_operation_log"
}
