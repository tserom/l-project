package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/sales-manage/internal/model"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"gorm.io/gorm"
)

var ErrSalesOrderNotFound = errors.New("sales order not found")

// SalesOrderRepository persists sales orders.
type SalesOrderRepository struct {
	db *gorm.DB
}

func NewSalesOrderRepository(db *gorm.DB) *SalesOrderRepository {
	return &SalesOrderRepository{db: db}
}

func (r *SalesOrderRepository) Create(ctx context.Context, order *model.SalesOrder) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(order).Error
}

func (r *SalesOrderRepository) GetByID(ctx context.Context, id string) (*model.SalesOrder, error) {
	var order model.SalesOrder
	err := r.db.WithContext(ctx).Preload("Lines").First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSalesOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *SalesOrderRepository) List(
	ctx context.Context,
	preds []qp.Predicate,
	page, pageSize int,
) ([]model.SalesOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 500 {
		pageSize = 500
	}

	q := qp.Apply(r.db.WithContext(ctx).Model(&model.SalesOrder{}), preds)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.SalesOrder
	err := q.Preload("Lines").
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// Replace saves header fields and replaces all lines in a transaction.
func (r *SalesOrderRepository) Replace(ctx context.Context, order *model.SalesOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("sales_order_id = ?", order.ID).Delete(&model.SalesOrderLine{}).Error; err != nil {
			return err
		}
		if err := tx.Omit("Lines").Save(order).Error; err != nil {
			return err
		}
		if len(order.Lines) == 0 {
			return nil
		}
		return tx.Create(&order.Lines).Error
	})
}

func (r *SalesOrderRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Select("Lines").Delete(&model.SalesOrder{ID: id})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrSalesOrderNotFound
	}
	return nil
}

// ListForCandidates returns matching orders with lines (no hard page cap beyond 500).
func (r *SalesOrderRepository) ListForCandidates(
	ctx context.Context,
	preds []qp.Predicate,
) ([]model.SalesOrder, error) {
	var list []model.SalesOrder
	err := qp.Apply(r.db.WithContext(ctx).Model(&model.SalesOrder{}), preds).
		Preload("Lines").
		Order("updated_at DESC").
		Limit(500).
		Find(&list).Error
	return list, err
}
