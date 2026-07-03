package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"gorm.io/gorm"
)

var ErrBatchNotFound = errors.New("batch not found")

// BatchRepository provides database access for material batch records.
type BatchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a BatchRepository.
func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

// List returns paginated batches matching optional qp predicates.
func (r *BatchRepository) List(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.MaterialBatch, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := qp.Apply(r.db.WithContext(ctx).Model(&model.MaterialBatch{}), preds)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var batches []model.MaterialBatch
	offset := (page - 1) * pageSize
	err := qp.Apply(r.db.WithContext(ctx), preds).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&batches).Error
	return batches, total, err
}

// GetByID returns a batch by primary key.
func (r *BatchRepository) GetByID(ctx context.Context, id uint64) (*model.MaterialBatch, error) {
	var batch model.MaterialBatch
	err := r.db.WithContext(ctx).First(&batch, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBatchNotFound
	}
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// ExistsByMaterialHeatOrg reports whether the business key already exists.
func (r *BatchRepository) ExistsByMaterialHeatOrg(ctx context.Context, materialID uint64, heatNo string, orgID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.MaterialBatch{}).
		Where("material_id = ? AND heat_no = ? AND org_id = ?", materialID, heatNo, orgID).
		Count(&count).Error
	return count > 0, err
}

// Create inserts a new batch record.
func (r *BatchRepository) Create(ctx context.Context, batch *model.MaterialBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

// Update saves changes to an existing batch record.
func (r *BatchRepository) Update(ctx context.Context, batch *model.MaterialBatch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}
