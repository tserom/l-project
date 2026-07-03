package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// SalesOrder records a customer sales order; confirmation does not affect stock.
type SalesOrder struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	DocNo        string           `gorm:"size:32;uniqueIndex;not null" json:"docNo"`
	DocDate      time.Time        `gorm:"type:date;not null" json:"docDate"`
	Status       DocStatus        `gorm:"size:16;not null;index" json:"status"`
	CustomerName string           `gorm:"size:128;not null" json:"customerName"`
	Operator     string           `gorm:"size:64;not null" json:"operator"`
	Remark       string           `gorm:"size:255" json:"remark"`
	OrgID        uint64           `gorm:"not null;default:0;index" json:"orgId"`
	CreatedBy    uint64           `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
	Lines        []SalesOrderLine `gorm:"foreignKey:SalesOrderID" json:"lines,omitempty"`
}

func (SalesOrder) TableName() string { return "sales_order" }

// SalesOrderLine is a sales line with optional unit price.
type SalesOrderLine struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	SalesOrderID uint64           `gorm:"not null;index" json:"salesOrderId"`
	MaterialID   uint64           `gorm:"not null;index" json:"materialId"`
	BatchID      uint64           `gorm:"not null;index" json:"batchId"`
	WeightKg     decimal.Decimal  `gorm:"type:decimal(20,4);not null;default:0" json:"weightKg"`
	Quantity     decimal.Decimal  `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
	UnitPrice    *decimal.Decimal `gorm:"type:decimal(20,4)" json:"unitPrice,omitempty"`
}

func (SalesOrderLine) TableName() string { return "sales_order_line" }
