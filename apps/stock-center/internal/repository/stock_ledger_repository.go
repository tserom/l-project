package repository

import (
	"context"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"gorm.io/gorm"
)

// StockLedgerRepository provides database access for stock ledger records.
type StockLedgerRepository struct {
	db *gorm.DB
}

// NewStockLedgerRepository creates a StockLedgerRepository.
func NewStockLedgerRepository(db *gorm.DB) *StockLedgerRepository {
	return &StockLedgerRepository{db: db}
}

// Create inserts a new ledger record.
func (r *StockLedgerRepository) Create(ctx context.Context, tx *gorm.DB, ledger *model.StockLedger) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(ledger).Error
}

// List returns paginated ledger entries matching optional qp predicates.
func (r *StockLedgerRepository) List(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.StockLedger, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := qp.Apply(r.db.WithContext(ctx).Model(&model.StockLedger{}), preds)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entries []model.StockLedger
	offset := (page - 1) * pageSize
	err := qp.Apply(r.db.WithContext(ctx), preds).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	return entries, total, err
}
