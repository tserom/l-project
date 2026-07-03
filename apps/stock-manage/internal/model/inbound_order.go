package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// InboundOrder records a draft or confirmed inbound stock document.
type InboundOrder struct {
	ID        uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	DocNo     string             `gorm:"size:32;uniqueIndex;not null" json:"docNo"`
	DocDate   time.Time          `gorm:"type:date;not null" json:"docDate"`
	Status    DocStatus          `gorm:"size:16;not null;index" json:"status"`
	Operator  string             `gorm:"size:64;not null" json:"operator"`
	Remark    string             `gorm:"size:255" json:"remark"`
	OrgID     uint64             `gorm:"not null;default:0;index" json:"orgId"`
	CreatedBy uint64             `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Lines     []InboundOrderLine `gorm:"foreignKey:InboundOrderID" json:"lines,omitempty"`
}

func (InboundOrder) TableName() string { return "inbound_order" }

// InboundOrderLine is a single inbound line referencing center material and batch.
type InboundOrderLine struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	InboundOrderID uint64          `gorm:"not null;index" json:"inboundOrderId"`
	MaterialID     uint64          `gorm:"not null;index" json:"materialId"`
	BatchID        uint64          `gorm:"not null;index" json:"batchId"`
	Warehouse      string          `gorm:"size:64;not null" json:"warehouse"`
	WeightKg       decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"weightKg"`
	Quantity       decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
}

func (InboundOrderLine) TableName() string { return "inbound_order_line" }
