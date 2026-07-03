package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// ProcessingOrder links raw material pick lines with finished goods inbound lines.
type ProcessingOrder struct {
	ID           uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	DocNo        string                 `gorm:"size:32;uniqueIndex;not null" json:"docNo"`
	DocDate      time.Time              `gorm:"type:date;not null" json:"docDate"`
	Status       DocStatus              `gorm:"size:16;not null;index" json:"status"`
	LossWeightKg decimal.Decimal        `gorm:"type:decimal(20,4);not null;default:0" json:"lossWeightKg"`
	Operator     string                 `gorm:"size:64;not null" json:"operator"`
	Remark       string                 `gorm:"size:255" json:"remark"`
	OrgID        uint64                 `gorm:"not null;default:0;index" json:"orgId"`
	CreatedBy    uint64                 `gorm:"not null;default:0" json:"createdBy"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
	PickLines    []ProcessingPickLine   `gorm:"foreignKey:ProcessingOrderID" json:"pickLines,omitempty"`
	FinishLines  []ProcessingFinishLine `gorm:"foreignKey:ProcessingOrderID" json:"finishLines,omitempty"`
}

func (ProcessingOrder) TableName() string { return "processing_order" }

// ProcessingPickLine records raw material picked for processing.
type ProcessingPickLine struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcessingOrderID uint64          `gorm:"not null;index" json:"processingOrderId"`
	MaterialID        uint64          `gorm:"not null;index" json:"materialId"`
	BatchID           uint64          `gorm:"not null;index" json:"batchId"`
	Warehouse         string          `gorm:"size:64;not null" json:"warehouse"`
	WeightKg          decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0" json:"weightKg"`
}

func (ProcessingPickLine) TableName() string { return "processing_pick_line" }

// ProcessingFinishLine records finished goods produced by processing.
type ProcessingFinishLine struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ProcessingOrderID uint64           `gorm:"not null;index" json:"processingOrderId"`
	MaterialID        uint64           `gorm:"not null;index" json:"materialId"`
	BatchID           uint64           `gorm:"not null;index" json:"batchId"`
	Warehouse         string           `gorm:"size:64;not null" json:"warehouse"`
	Quantity          decimal.Decimal  `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
	WeightKg          *decimal.Decimal `gorm:"type:decimal(20,4)" json:"weightKg,omitempty"`
}

func (ProcessingFinishLine) TableName() string { return "processing_finish_line" }
