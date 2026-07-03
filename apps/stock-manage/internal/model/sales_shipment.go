package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// SalesShipment records actual outbound stock against a sales order.
type SalesShipment struct {
	ID           uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	DocNo        string              `gorm:"size:32;uniqueIndex;not null" json:"docNo"`
	DocDate      time.Time           `gorm:"type:date;not null" json:"docDate"`
	Status       DocStatus           `gorm:"size:16;not null;index" json:"status"`
	SalesOrderID uint64              `gorm:"not null;index" json:"salesOrderId"`
	Operator     string              `gorm:"size:64;not null" json:"operator"`
	Remark       string              `gorm:"size:255" json:"remark"`
	OrgID        uint64              `gorm:"not null;default:0;index" json:"orgId"`
	CreatedBy    uint64              `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	Lines        []SalesShipmentLine `gorm:"foreignKey:SalesShipmentID" json:"lines,omitempty"`
}

func (SalesShipment) TableName() string { return "sales_shipment" }

// SalesShipmentLine is the actual outbound quantity for a shipment.
type SalesShipmentLine struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	SalesShipmentID uint64          `gorm:"not null;index" json:"salesShipmentId"`
	MaterialID      uint64          `gorm:"not null;index" json:"materialId"`
	BatchID         uint64          `gorm:"not null;index" json:"batchId"`
	Warehouse       string          `gorm:"size:64;not null" json:"warehouse"`
	WeightKg        decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"weightKg"`
	Quantity        decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
}

func (SalesShipmentLine) TableName() string { return "sales_shipment_line" }
