package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"gorm.io/gorm"
)

// SalesShipmentRepository provides database access for sales shipments.
type SalesShipmentRepository struct {
	db *gorm.DB
}

// NewSalesShipmentRepository creates a SalesShipmentRepository.
func NewSalesShipmentRepository(db *gorm.DB) *SalesShipmentRepository {
	return &SalesShipmentRepository{db: db}
}

// Create inserts a sales shipment and its lines in one transaction.
func (r *SalesShipmentRepository) Create(ctx context.Context, shipment *model.SalesShipment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lines := shipment.Lines
		shipment.Lines = nil
		if err := tx.Create(shipment).Error; err != nil {
			return err
		}
		for i := range lines {
			lines[i].ID = 0
			lines[i].SalesShipmentID = shipment.ID
		}
		if len(lines) > 0 {
			if err := tx.Create(&lines).Error; err != nil {
				return err
			}
		}
		shipment.Lines = lines
		return nil
	})
}

// GetByID returns a sales shipment with lines preloaded.
func (r *SalesShipmentRepository) GetByID(ctx context.Context, id uint64) (*model.SalesShipment, error) {
	var shipment model.SalesShipment
	err := r.db.WithContext(ctx).Preload("Lines").First(&shipment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSalesShipmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}

// List returns paginated sales shipments without lines.
func (r *SalesShipmentRepository) List(ctx context.Context, page, pageSize int) ([]model.SalesShipment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&model.SalesShipment{})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var shipments []model.SalesShipment
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&shipments).Error
	return shipments, total, err
}

// ListBySalesOrderID returns all shipments for a sales order without lines.
func (r *SalesShipmentRepository) ListBySalesOrderID(ctx context.Context, salesOrderID uint64) ([]model.SalesShipment, error) {
	var shipments []model.SalesShipment
	err := r.db.WithContext(ctx).
		Where("sales_order_id = ?", salesOrderID).
		Order("id DESC").
		Find(&shipments).Error
	return shipments, err
}

// Update replaces header fields and lines for a draft sales shipment.
func (r *SalesShipmentRepository) Update(ctx context.Context, shipment *model.SalesShipment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.SalesShipment
		if err := tx.First(&existing, shipment.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesShipmentNotFound
			}
			return err
		}
		if existing.Status != model.DocStatusDraft {
			return ErrNotDraft
		}

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"doc_date": shipment.DocDate,
			"operator": shipment.Operator,
			"remark":   shipment.Remark,
		}).Error; err != nil {
			return err
		}

		if err := tx.Where("sales_shipment_id = ?", shipment.ID).Delete(&model.SalesShipmentLine{}).Error; err != nil {
			return err
		}
		for i := range shipment.Lines {
			shipment.Lines[i].ID = 0
			shipment.Lines[i].SalesShipmentID = shipment.ID
		}
		if len(shipment.Lines) > 0 {
			if err := tx.Create(&shipment.Lines).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete removes a draft sales shipment and its lines.
func (r *SalesShipmentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var shipment model.SalesShipment
		if err := tx.First(&shipment, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesShipmentNotFound
			}
			return err
		}
		if shipment.Status != model.DocStatusDraft {
			return ErrNotDraft
		}
		if err := tx.Where("sales_shipment_id = ?", id).Delete(&model.SalesShipmentLine{}).Error; err != nil {
			return err
		}
		return tx.Delete(&shipment).Error
	})
}

// ConfirmStatus sets status to confirmed for a sales shipment.
func (r *SalesShipmentRepository) ConfirmStatus(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.SalesShipment{}).
		Where("id = ? AND status = ?", id, model.DocStatusDraft).
		Update("status", model.DocStatusConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var shipment model.SalesShipment
		if err := r.db.WithContext(ctx).First(&shipment, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSalesShipmentNotFound
			}
			return err
		}
		return ErrNotDraft
	}
	return nil
}
