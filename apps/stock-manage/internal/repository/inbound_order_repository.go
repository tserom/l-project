package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
)

// InboundOrderRepository provides database access for inbound orders.
type InboundOrderRepository struct {
	db *gorm.DB
}

// NewInboundOrderRepository creates an InboundOrderRepository.
func NewInboundOrderRepository(db *gorm.DB) *InboundOrderRepository {
	return &InboundOrderRepository{db: db}
}

// Create inserts an inbound order and its lines in one transaction.
func (r *InboundOrderRepository) Create(ctx context.Context, order *model.InboundOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lines := order.Lines
		order.Lines = nil
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range lines {
			lines[i].ID = 0
			lines[i].InboundOrderID = order.ID
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

// GetByID returns an inbound order with lines preloaded.
func (r *InboundOrderRepository) GetByID(ctx context.Context, id uint64) (*model.InboundOrder, error) {
	var order model.InboundOrder
	err := r.db.WithContext(ctx).Preload("Lines").First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInboundOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// List returns paginated inbound orders without lines.
func (r *InboundOrderRepository) List(ctx context.Context, page, pageSize int) ([]model.InboundOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&model.InboundOrder{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []model.InboundOrder
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

// Update replaces header fields and lines for a draft inbound order.
func (r *InboundOrderRepository) Update(ctx context.Context, order *model.InboundOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.InboundOrder
		if err := tx.First(&existing, order.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInboundOrderNotFound
			}
			return err
		}
		if existing.Status != model.DocStatusDraft {
			return ErrNotDraft
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"doc_date": order.DocDate,
			"operator": order.Operator,
			"remark":   order.Remark,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("inbound_order_id = ?", order.ID).Delete(&model.InboundOrderLine{}).Error; err != nil {
			return err
		}
		for i := range order.Lines {
			order.Lines[i].ID = 0
			order.Lines[i].InboundOrderID = order.ID
		}
		if len(order.Lines) > 0 {
			if err := tx.Create(&order.Lines).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes a draft inbound order and its lines.
func (r *InboundOrderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.InboundOrder
		if err := tx.First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInboundOrderNotFound
			}
			return err
		}
		if order.Status != model.DocStatusDraft {
			return ErrNotDraft
		}
		if err := tx.Where("inbound_order_id = ?", id).Delete(&model.InboundOrderLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&order).Error
	})
}

// ConfirmStatus sets status to confirmed for an inbound order.
func (r *InboundOrderRepository) ConfirmStatus(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.InboundOrder{}).
		Where("id = ? AND status = ?", id, model.DocStatusDraft).
		Update("status", model.DocStatusConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var order model.InboundOrder
		if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInboundOrderNotFound
			}
			return err
		}
		return ErrNotDraft
	}
	return nil
}
