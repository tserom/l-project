package model

import "time"

// SalesOrder is a sales-front sales order header (not stock-manage inventory SO).
type SalesOrder struct {
	ID                  string           `gorm:"primaryKey;size:36" json:"id"`
	OrderNo             string           `gorm:"size:64;not null;index" json:"orderNo"`
	OrderDate           string           `gorm:"size:10;not null;index" json:"orderDate"`
	CustomerName        string           `gorm:"size:128;not null;index" json:"customerName"`
	WarehouseName       string           `gorm:"size:128;not null" json:"warehouseName"`
	DeliveryType        string           `gorm:"size:64;not null" json:"deliveryType"`
	Remark              string           `gorm:"size:512" json:"remark,omitempty"`
	OutstandingBalance  string           `gorm:"size:64" json:"outstandingBalance,omitempty"`
	TotalQuantity       float64          `gorm:"type:decimal(20,4);not null;default:0" json:"totalQuantity"`
	TotalAmount         float64          `gorm:"type:decimal(20,4);not null;default:0" json:"totalAmount"`
	CreatedAt           time.Time        `json:"createdAt"`
	UpdatedAt           time.Time        `gorm:"index" json:"updatedAt"`
	Lines               []SalesOrderLine `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE" json:"lines"`
}

func (SalesOrder) TableName() string { return "sales_order" }

// SalesOrderLine is a sales order detail line.
type SalesOrderLine struct {
	ID            string  `gorm:"primaryKey;size:36" json:"id"`
	SalesOrderID  string  `gorm:"size:36;not null;index" json:"salesOrderId"`
	MaterialName  string  `gorm:"size:256;not null" json:"materialName"`
	Spec          string  `gorm:"size:256" json:"spec"`
	Unit          string  `gorm:"size:32;not null" json:"unit"`
	Quantity      float64 `gorm:"type:decimal(20,4);not null;default:0" json:"quantity"`
	UnitPrice     float64 `gorm:"type:decimal(20,4);not null;default:0" json:"unitPrice"`
	Amount        float64 `gorm:"type:decimal(20,4);not null;default:0" json:"amount"`
	LineRemark    string  `gorm:"size:512" json:"lineRemark,omitempty"`
	NeedInvoice   bool    `gorm:"not null;default:false" json:"needInvoice"`
	InvoiceDocID  *string `gorm:"size:36;index" json:"invoiceDocId,omitempty"`
}

func (SalesOrderLine) TableName() string { return "sales_order_line" }
