package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type LedgerRefType string

const (
	RefInbound    LedgerRefType = "inbound"
	RefOutbound   LedgerRefType = "outbound"
	RefProcessing LedgerRefType = "processing"
	RefSale       LedgerRefType = "sale"
	RefAdjust     LedgerRefType = "adjust"
)

// StockLedger is an immutable inventory movement record.
type StockLedger struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID    uint64          `gorm:"not null;index" json:"materialId"`
	BatchID       uint64          `gorm:"not null;index" json:"batchId"`
	Warehouse     string          `gorm:"size:64;not null;index" json:"warehouse"`
	DeltaWeightKg decimal.Decimal `gorm:"type:decimal(20,4);not null" json:"deltaWeightKg"`
	DeltaQuantity decimal.Decimal `gorm:"type:decimal(20,4);not null" json:"deltaQuantity"`
	RefType       LedgerRefType   `gorm:"size:16;not null;index" json:"refType"`
	RefNo         string          `gorm:"size:64;not null;index" json:"refNo"`
	Remark        string          `gorm:"size:256" json:"remark"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (StockLedger) TableName() string { return "stock_ledger" }
