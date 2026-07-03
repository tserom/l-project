package model

import "time"

// MaterialBatch tracks heat/lot numbers per material within an organization.
type MaterialBatch struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialID uint64    `gorm:"not null;uniqueIndex:uk_material_heat;index" json:"materialId"`
	HeatNo     string    `gorm:"size:64;not null;uniqueIndex:uk_material_heat" json:"heatNo"`
	Remark     string    `gorm:"size:256" json:"remark"`
	OrgID      uint64    `gorm:"not null;default:0;uniqueIndex:uk_material_heat;index" json:"orgId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (MaterialBatch) TableName() string { return "material_batch" }
