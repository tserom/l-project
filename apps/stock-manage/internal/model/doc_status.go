package model

// DocStatus is the lifecycle state of a business document.
type DocStatus string

const (
	DocStatusDraft     DocStatus = "draft"
	DocStatusConfirmed DocStatus = "confirmed"
)
