package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/sales-manage/internal/model"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"gorm.io/gorm"
)

var ErrInvoiceDocNotFound = errors.New("invoice doc not found")

// InvoiceDocRepository persists invoice documents.
type InvoiceDocRepository struct {
	db *gorm.DB
}

func NewInvoiceDocRepository(db *gorm.DB) *InvoiceDocRepository {
	return &InvoiceDocRepository{db: db}
}

func (r *InvoiceDocRepository) Create(ctx context.Context, doc *model.InvoiceDoc) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(doc).Error
}

func (r *InvoiceDocRepository) GetByID(ctx context.Context, id string) (*model.InvoiceDoc, error) {
	var doc model.InvoiceDoc
	err := r.db.WithContext(ctx).Preload("Lines").First(&doc, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvoiceDocNotFound
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *InvoiceDocRepository) List(
	ctx context.Context,
	preds []qp.Predicate,
	page, pageSize int,
) ([]model.InvoiceDoc, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 500 {
		pageSize = 500
	}

	q := qp.Apply(r.db.WithContext(ctx).Model(&model.InvoiceDoc{}), preds)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.InvoiceDoc
	err := q.Preload("Lines").
		Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *InvoiceDocRepository) Save(ctx context.Context, doc *model.InvoiceDoc) error {
	return r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(doc).Error
}
