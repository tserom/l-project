package model

import "time"

// DocSequence tracks the last issued sequence number per prefix and calendar day.
type DocSequence struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Prefix    string    `gorm:"size:8;not null;uniqueIndex:uk_doc_seq" json:"prefix"`
	SeqDate   time.Time `gorm:"type:date;not null;uniqueIndex:uk_doc_seq" json:"seqDate"`
	LastSeq   uint32    `gorm:"not null" json:"lastSeq"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (DocSequence) TableName() string { return "doc_sequence" }
