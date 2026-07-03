package docno

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Generate allocates the next document number for prefix on the current calendar day.
// Format: {PREFIX}{YYYYMMDD}{4-digit seq}, e.g. IN202607030001. Sequence resets daily per prefix.
func Generate(ctx context.Context, db *gorm.DB, prefix string) (string, error) {
	if prefix == "" {
		return "", errors.New("prefix is required")
	}

	today := dateOnly(time.Now())
	datePart := today.Format("20060102")

	var docNo string
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seq model.DocSequence
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("prefix = ? AND seq_date = ?", prefix, today).
			First(&seq).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			seq = model.DocSequence{
				Prefix:  prefix,
				SeqDate: today,
				LastSeq: 1,
			}
			if err := tx.Create(&seq).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			seq.LastSeq++
			if err := tx.Save(&seq).Error; err != nil {
				return err
			}
		}

		docNo = fmt.Sprintf("%s%s%04d", prefix, datePart, seq.LastSeq)
		return nil
	})
	if err != nil {
		return "", err
	}
	return docNo, nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
