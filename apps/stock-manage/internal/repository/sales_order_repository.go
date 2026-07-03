package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
)

// SalesOrderRepository provides database access for sales orders.
type SalesOrderRepository struct {
	db *gorm.DB
}

// NewSalesOrderRepository creates a SalesOrderRepository.
func NewSalesOrderRepository(db *gorm.DB) *SalesOrderRepository {
	return &SalesOrderRepository{db: db}
}

// Create inserts a sales order and its lines in one transaction.
func (r *SalesOrderRepository) Create(ctx context.Context, order *model.SalesOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lines := order.Lines
		order.Lines = nil
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range lines {
			lines[i].ID = 0
			lines[i].SalesOrderID = order.ID
		}
		if len(lines) > 0 {
			if err := tx.Create(&lines).Error; err != nil {
				return err
			}
		}
		order.Lines = lines
		return nil
	})
}

// GetByID returns a sales order with lines preloaded.
func (r *SalesOrderRepository) GetByID(ctx context.Context, id uint64) (*model.SalesOrder, error) {
	var order model.SalesOrder
	err := r.db.WithContext(ctx).Preload("Lines").First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSalesOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// List returns paginated sales orders without lines.
func (r *SalesOrderRepository) List(ctx context.Context, page, pageSize int) ([]model.SalesOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&model.SalesOrder{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []model.SalesOrder
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

// Update replaces header fields and lines for a draft sales order.
func (r *SalesOrderRepository) Update(ctx context.Context, order *model.SalesOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.SalesOrder
		if err := tx.First(&existing, order.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesOrderNotFound
			}
			return err
		}
		if existing.Status != model.DocStatusDraft {
			return ErrNotDraft
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"doc_date":      order.DocDate,
			"customer_name": order.CustomerName,
			"operator":      order.Operator,
			"remark":        order.Remark,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("sales_order_id = ?", order.ID).Delete(&model.SalesOrderLine{}).Error; err != nil {
			return err
		}
		for i := range order.Lines {
			order.Lines[i].ID = 0
			order.Lines[i].SalesOrderID = order.ID
		}
		if len(order.Lines) > 0 {
			if err := tx.Create(&order.Lines).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes a draft sales order and its lines.
func (r *SalesOrderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.SalesOrder
		if err := tx.First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesOrderNotFound
			}
			return err
		}
		if order.Status != model.DocStatusDraft {
			return ErrNotDraft
		}
		if err := tx.Where("sales_order_id = ?", id).Delete(&model.SalesOrderLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&order).Error
	})
}

// ConfirmStatus sets status to confirmed for a sales order.
func (r *SalesOrderRepository) ConfirmStatus(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.SalesOrder{}).
		Where("id = ? AND status = ?", id, model.DocStatusDraft).
		Update("status", model.DocStatusConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var order model.SalesOrder
		if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesOrderNotFound
			}
			return err
		}
		return ErrNotDraft
	}
	return nil
}
