package model

import "time"

type MaterialForm string
type PrimaryUnit string
type MaterialType string
type MaterialStatus string

const (
	FormPlate      MaterialForm = "plate"
	FormPipe       MaterialForm = "pipe"
	FormBar        MaterialForm = "bar"
	FormProfile    MaterialForm = "profile"
	FormPart       MaterialForm = "part"
	UnitKg         PrimaryUnit = "kg"
	UnitPiece      PrimaryUnit = "piece"
	UnitMeter      PrimaryUnit = "meter"
	TypeRaw        MaterialType = "raw"
	TypeFinished   MaterialType = "finished"
	StatusEnabled  MaterialStatus = "enabled"
	StatusDisabled MaterialStatus = "disabled"
)

// Material is the master data record for stainless steel items.
type Material struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	MaterialCode string         `gorm:"size:64;uniqueIndex:uk_material_code;not null" json:"materialCode"`
	Grade        string         `gorm:"size:32;not null;index" json:"grade"`
	Form         MaterialForm   `gorm:"size:16;not null" json:"form"`
	Spec         string         `gorm:"size:128;not null" json:"spec"`
	PrimaryUnit  PrimaryUnit    `gorm:"size:16;not null" json:"primaryUnit"`
	MaterialType MaterialType   `gorm:"size:16;not null" json:"materialType"`
	Status       MaterialStatus `gorm:"size:16;not null;default:enabled" json:"status"`
	OrgID        uint64         `gorm:"not null;default:0;index" json:"orgId"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

func (Material) TableName() string { return "material" }
