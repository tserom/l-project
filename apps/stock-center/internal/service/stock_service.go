package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
)

// StockService coordinates stock data access operations.
type StockService struct {
	repo *repository.StockRepository
}

// NewStockService creates a StockService.
func NewStockService(repo *repository.StockRepository) *StockService {
	return &StockService{repo: repo}
}

// ListStocks returns paginated stock records.
func (s *StockService) ListStocks(ctx context.Context, page, pageSize int) ([]model.Stock, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// GetStock returns a stock record by ID.
func (s *StockService) GetStock(ctx context.Context, id uint64) (*model.Stock, error) {
	return s.repo.GetByID(ctx, id)
}

// GetStockBySKU returns a stock record by SKU and warehouse.
func (s *StockService) GetStockBySKU(ctx context.Context, sku, warehouse string) (*model.Stock, error) {
	return s.repo.GetBySKUAndWarehouse(ctx, sku, warehouse)
}

// CreateStockInput is the payload for creating stock.
type CreateStockInput struct {
	SKU       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int64  `json:"quantity"`
}

// CreateStock creates a new stock record.
func (s *StockService) CreateStock(ctx context.Context, input CreateStockInput) (*model.Stock, error) {
	if input.SKU == "" || input.Warehouse == "" {
		return nil, errors.New("sku and warehouse are required")
	}
	if input.Quantity < 0 {
		return nil, errors.New("quantity must be non-negative")
	}

	stock := &model.Stock{
		SKU:       input.SKU,
		Warehouse: input.Warehouse,
		Quantity:  input.Quantity,
	}
	if err := s.repo.Create(ctx, stock); err != nil {
		return nil, err
	}
	return stock, nil
}

// AdjustStockInput is the payload for updating stock quantity.
type AdjustStockInput struct {
	Quantity int64 `json:"quantity"`
}

// AdjustStock updates stock quantity.
func (s *StockService) AdjustStock(ctx context.Context, id uint64, input AdjustStockInput) (*model.Stock, error) {
	if input.Quantity < 0 {
		return nil, errors.New("quantity must be non-negative")
	}

	stock, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stock.Quantity = input.Quantity
	if err := s.repo.Update(ctx, stock); err != nil {
		return nil, err
	}
	return stock, nil
}
