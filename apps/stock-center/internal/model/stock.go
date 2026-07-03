package model

import "time"

// Stock represents inventory quantity for a SKU in a warehouse.
type Stock struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SKU       string    `gorm:"size:64;uniqueIndex:uk_sku_warehouse;not null" json:"sku"`
	Warehouse string    `gorm:"size:64;uniqueIndex:uk_sku_warehouse;not null" json:"warehouse"`
	Quantity  int64     `gorm:"not null;default:0" json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName overrides the default GORM table name.
func (Stock) TableName() string {
	return "stock"
}
