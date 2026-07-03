package repository

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"gorm.io/gorm"
)

var ErrStockNotFound = errors.New("stock not found")

// StockRepository provides direct database access for stock records.
type StockRepository struct {
	db *gorm.DB
}

// NewStockRepository creates a StockRepository.
func NewStockRepository(db *gorm.DB) *StockRepository {
	return &StockRepository{db: db}
}

// List returns paginated stock records.
func (r *StockRepository) List(ctx context.Context, page, pageSize int) ([]model.Stock, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Stock{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var stocks []model.Stock
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&stocks).Error
	return stocks, total, err
}

// GetByID returns a stock record by primary key.
func (r *StockRepository) GetByID(ctx context.Context, id uint64) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.WithContext(ctx).First(&stock, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStockNotFound
	}
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// GetBySKUAndWarehouse returns a stock record by business keys.
func (r *StockRepository) GetBySKUAndWarehouse(ctx context.Context, sku, warehouse string) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.WithContext(ctx).
		Where("sku = ? AND warehouse = ?", sku, warehouse).
		First(&stock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStockNotFound
	}
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// Create inserts a new stock record.
func (r *StockRepository) Create(ctx context.Context, stock *model.Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

// Update saves changes to an existing stock record.
func (r *StockRepository) Update(ctx context.Context, stock *model.Stock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}
