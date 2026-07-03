package repository

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"gorm.io/gorm"
)

var (
	ErrStockBalanceNotFound  = errors.New("stock balance not found")
	ErrInsufficientWeight    = errors.New("insufficient weight")
	ErrInsufficientQuantity  = errors.New("insufficient quantity")
)

// StockBalanceRepository provides database access for stock balance records.
type StockBalanceRepository struct {
	db *gorm.DB
}

// NewStockBalanceRepository creates a StockBalanceRepository.
func NewStockBalanceRepository(db *gorm.DB) *StockBalanceRepository {
	return &StockBalanceRepository{db: db}
}

// DB returns the underlying database handle for transactions.
func (r *StockBalanceRepository) DB() *gorm.DB {
	return r.db
}

// List returns paginated stock balances with optional material/batch fields for display.
func (r *StockBalanceRepository) List(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.StockBalance, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	base := r.db.WithContext(ctx).Model(&model.StockBalance{})
	q := qp.Apply(base, preds)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var balances []model.StockBalance
	offset := (page - 1) * pageSize
	err := qp.Apply(r.db.WithContext(ctx), preds).
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&balances).Error
	return balances, total, err
}

// GetByKey returns a balance for material + batch + warehouse within an org.
func (r *StockBalanceRepository) GetByKey(ctx context.Context, tx *gorm.DB, materialID, batchID uint64, warehouse string, orgID uint64) (*model.StockBalance, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var balance model.StockBalance
	err := db.WithContext(ctx).
		Where("material_id = ? AND batch_id = ? AND warehouse = ? AND org_id = ?", materialID, batchID, warehouse, orgID).
		First(&balance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStockBalanceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// Add increases weight and quantity on an existing or new balance row.
func (r *StockBalanceRepository) Add(ctx context.Context, tx *gorm.DB, materialID, batchID uint64, warehouse string, orgID uint64, weightKg, quantity decimal.Decimal) (*model.StockBalance, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	balance, err := r.GetByKey(ctx, tx, materialID, batchID, warehouse, orgID)
	if errors.Is(err, ErrStockBalanceNotFound) {
		balance = &model.StockBalance{
			MaterialID: materialID,
			BatchID:    batchID,
			Warehouse:  warehouse,
			WeightKg:   weightKg,
			Quantity:   quantity,
			OrgID:      orgID,
		}
		if err := db.WithContext(ctx).Create(balance).Error; err != nil {
			return nil, err
		}
		return balance, nil
	}
	if err != nil {
		return nil, err
	}

	balance.WeightKg = balance.WeightKg.Add(weightKg)
	balance.Quantity = balance.Quantity.Add(quantity)
	if err := db.WithContext(ctx).Save(balance).Error; err != nil {
		return nil, err
	}
	return balance, nil
}

// Subtract decreases weight and quantity on an existing balance row.
func (r *StockBalanceRepository) Subtract(ctx context.Context, tx *gorm.DB, materialID, batchID uint64, warehouse string, orgID uint64, weightKg, quantity decimal.Decimal) (*model.StockBalance, error) {
	balance, err := r.GetByKey(ctx, tx, materialID, batchID, warehouse, orgID)
	if err != nil {
		return nil, err
	}

	newWeight := balance.WeightKg.Sub(weightKg)
	if newWeight.IsNegative() {
		return nil, ErrInsufficientWeight
	}

	newQuantity := balance.Quantity.Sub(quantity)
	if newQuantity.IsNegative() {
		return nil, ErrInsufficientQuantity
	}

	balance.WeightKg = newWeight
	balance.Quantity = newQuantity

	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.WithContext(ctx).Save(balance).Error; err != nil {
		return nil, err
	}
	return balance, nil
}
