package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"gorm.io/gorm"
)

var ErrMaterialNotFound = errors.New("material not found")

// MaterialRepository provides database access for material records.
type MaterialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository creates a MaterialRepository.
func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// List returns paginated materials matching optional qp predicates.
func (r *MaterialRepository) List(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.Material, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := qp.Apply(r.db.WithContext(ctx).Model(&model.Material{}), preds)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var materials []model.Material
	offset := (page - 1) * pageSize
	err := qp.Apply(r.db.WithContext(ctx), preds).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&materials).Error
	return materials, total, err
}

// GetByID returns a material by primary key.
func (r *MaterialRepository) GetByID(ctx context.Context, id uint64) (*model.Material, error) {
	var material model.Material
	err := r.db.WithContext(ctx).First(&material, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMaterialNotFound
	}
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// ExistsByMaterialCode reports whether a material code is already taken.
func (r *MaterialRepository) ExistsByMaterialCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Material{}).Where("material_code = ?", code).Count(&count).Error
	return count > 0, err
}

// Create inserts a new material record.
func (r *MaterialRepository) Create(ctx context.Context, material *model.Material) error {
	return r.db.WithContext(ctx).Create(material).Error
}

// Update saves changes to an existing material record.
func (r *MaterialRepository) Update(ctx context.Context, material *model.Material) error {
	return r.db.WithContext(ctx).Save(material).Error
}
