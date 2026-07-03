package repository

import (
	"context"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
)

// OperationLogRepository provides database access for business audit logs.
type OperationLogRepository struct {
	db *gorm.DB
}

// NewOperationLogRepository creates an OperationLogRepository.
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create inserts a new operation log.
func (r *OperationLogRepository) Create(ctx context.Context, log *model.StockOperationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
