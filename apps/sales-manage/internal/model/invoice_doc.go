package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// InvoiceDocStatus matches sales-front InvoiceDocStatus.
type InvoiceDocStatus string

const (
	InvoiceStatusSaved  InvoiceDocStatus = "saved"
	InvoiceStatusVoided InvoiceDocStatus = "voided"
)

// StringList stores a JSON string array in MySQL.
type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal([]string(s))
	return string(b), err
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("StringList: unsupported Scan type")
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*s = out
	return nil
}

// InvoiceDoc is an invoice document header.
type InvoiceDoc struct {
	ID                 string           `gorm:"primaryKey;size:36" json:"id"`
	InvoiceNo          string           `gorm:"size:64;not null;uniqueIndex" json:"invoiceNo"`
	Status             InvoiceDocStatus `gorm:"size:16;not null;index" json:"status"`
	FilterCustomerName string           `gorm:"size:128" json:"filterCustomerName,omitempty"`
	FilterDateFrom     string           `gorm:"size:10" json:"filterDateFrom,omitempty"`
	FilterDateTo       string           `gorm:"size:10" json:"filterDateTo,omitempty"`
	FilterOrderNos     StringList       `gorm:"type:json" json:"filterOrderNos,omitempty"`
	TotalQuantity      float64          `gorm:"type:decimal(20,4);not null;default:0" json:"totalQuantity"`
	TotalAmount        float64          `gorm:"type:decimal(20,4);not null;default:0" json:"totalAmount"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `gorm:"index" json:"updatedAt"`
	VoidedAt           *time.Time       `json:"voidedAt,omitempty"`
	Lines              []InvoiceDocLine `gorm:"foreignKey:InvoiceDocID;constraint:OnDelete:CASCADE" json:"lines"`
}

func (InvoiceDoc) TableName() string { return "invoice_doc" }

// InvoiceDocLine is an invoice detail with sales-order snapshot fields.
type InvoiceDocLine struct {
	ID               string  `gorm:"primaryKey;size:36" json:"id"`
	InvoiceDocID     string  `gorm:"size:36;not null;index" json:"invoiceDocId"`
	SalesOrderID     string  `gorm:"size:36;not null;index" json:"salesOrderId"`
	SalesOrderLineID string  `gorm:"size:36;not null;index" json:"salesOrderLineId"`
	MaterialName     string  `gorm:"size:256;not null" json:"materialName"`
	Spec             string  `gorm:"size:256" json:"spec"`
	Unit             string  `gorm:"size:32;not null" json:"unit"`
	Quantity         float64 `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
	UnitPrice        float64 `gorm:"type:decimal(20,4);not null;default:0" json:"unitPrice"`
	Amount           float64 `gorm:"type:decimal(20,4);not null;default:0" json:"amount"`
	LineRemark       string  `gorm:"size:512" json:"lineRemark,omitempty"`
	OrderNo          string  `gorm:"size:64;not null" json:"orderNo"`
	OrderDate        string  `gorm:"size:10;not null" json:"orderDate"`
	CustomerName     string  `gorm:"size:128;not null" json:"customerName"`
	WarehouseName    string  `gorm:"size:128;not null" json:"warehouseName"`
	DeliveryType     string  `gorm:"size:64;not null" json:"deliveryType"`
}

func (InvoiceDocLine) TableName() string { return "invoice_doc_line" }
