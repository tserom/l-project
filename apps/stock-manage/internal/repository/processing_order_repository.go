package repository

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
)

// ProcessingOrderRepository provides database access for processing orders.
type ProcessingOrderRepository struct {
	db *gorm.DB
}

// NewProcessingOrderRepository creates a ProcessingOrderRepository.
func NewProcessingOrderRepository(db *gorm.DB) *ProcessingOrderRepository {
	return &ProcessingOrderRepository{db: db}
}

// Create inserts a processing order with pick and finish lines in one transaction.
func (r *ProcessingOrderRepository) Create(ctx context.Context, order *model.ProcessingOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pickLines := order.PickLines
		finishLines := order.FinishLines
		order.PickLines = nil
		order.FinishLines = nil
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range pickLines {
			pickLines[i].ID = 0
			pickLines[i].ProcessingOrderID = order.ID
		}
		if len(pickLines) > 0 {
			if err := tx.Create(&pickLines).Error; err != nil {
				return err
			}
		}
		for i := range finishLines {
			finishLines[i].ID = 0
			finishLines[i].ProcessingOrderID = order.ID
		}
		if len(finishLines) > 0 {
			if err := tx.Create(&finishLines).Error; err != nil {
				return err
			}
		}
		order.PickLines = pickLines
		order.FinishLines = finishLines
		return nil
	})
}

// GetByID returns a processing order with pick and finish lines preloaded.
func (r *ProcessingOrderRepository) GetByID(ctx context.Context, id uint64) (*model.ProcessingOrder, error) {
	var order model.ProcessingOrder
	err := r.db.WithContext(ctx).
		Preload("PickLines").
		Preload("FinishLines").
		First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProcessingOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// List returns paginated processing orders without lines.
func (r *ProcessingOrderRepository) List(ctx context.Context, page, pageSize int) ([]model.ProcessingOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&model.ProcessingOrder{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []model.ProcessingOrder
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

// Update replaces header fields and lines for a draft processing order.
func (r *ProcessingOrderRepository) Update(ctx context.Context, order *model.ProcessingOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ProcessingOrder
		if err := tx.First(&existing, order.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProcessingOrderNotFound
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

		if err := tx.Where("processing_order_id = ?", order.ID).Delete(&model.ProcessingPickLine{}).Error; err != nil {
			return err
		}
		if err := tx.Where("processing_order_id = ?", order.ID).Delete(&model.ProcessingFinishLine{}).Error; err != nil {
			return err
		}

		for i := range order.PickLines {
			order.PickLines[i].ID = 0
			order.PickLines[i].ProcessingOrderID = order.ID
		}
		if len(order.PickLines) > 0 {
			if err := tx.Create(&order.PickLines).Error; err != nil {
				return err
			}
		}
		for i := range order.FinishLines {
			order.FinishLines[i].ID = 0
			order.FinishLines[i].ProcessingOrderID = order.ID
		}
		if len(order.FinishLines) > 0 {
			if err := tx.Create(&order.FinishLines).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes a draft processing order and its lines.
func (r *ProcessingOrderRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.ProcessingOrder
		if err := tx.First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProcessingOrderNotFound
			}
			return err
		}
		if order.Status != model.DocStatusDraft {
			return ErrNotDraft
		}
		if err := tx.Where("processing_order_id = ?", id).Delete(&model.ProcessingPickLine{}).Error; err != nil {
			return err
		}
		if err := tx.Where("processing_order_id = ?", id).Delete(&model.ProcessingFinishLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&order).Error
	})
}

// ConfirmStatus sets status to confirmed and persists lossWeightKg.
func (r *ProcessingOrderRepository) ConfirmStatus(ctx context.Context, id uint64, lossWeightKg decimal.Decimal) error {
	result := r.db.WithContext(ctx).
		Model(&model.ProcessingOrder{}).
		Where("id = ? AND status = ?", id, model.DocStatusDraft).
		Updates(map[string]interface{}{
			"status":         model.DocStatusConfirmed,
			"loss_weight_kg": lossWeightKg,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var order model.ProcessingOrder
		if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProcessingOrderNotFound
			}
			return err
		}
		return ErrNotDraft
	}
	return nil
}
