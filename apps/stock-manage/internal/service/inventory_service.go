package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
)

// InventoryService orchestrates business stock operations.
type InventoryService struct {
	stockCenter *stockcenter.Client
	logRepo     *repository.OperationLogRepository
}

// NewInventoryService creates an InventoryService.
func NewInventoryService(
	stockCenter *stockcenter.Client,
	logRepo *repository.OperationLogRepository,
) *InventoryService {
	return &InventoryService{
		stockCenter: stockCenter,
		logRepo:     logRepo,
	}
}

// QueryInventoryInput is the business query payload.
type QueryInventoryInput struct {
	SKU       string `form:"sku"`
	Warehouse string `form:"warehouse"`
}

// QueryInventory returns inventory for a SKU in a warehouse.
func (s *InventoryService) QueryInventory(ctx context.Context, input QueryInventoryInput) (*stockcenter.Stock, error) {
	if input.SKU == "" || input.Warehouse == "" {
		return nil, errors.New("sku and warehouse are required")
	}
	return s.stockCenter.GetStockBySKU(ctx, input.SKU, input.Warehouse)
}

// ListInventory returns paginated inventory from stock-center.
func (s *InventoryService) ListInventory(ctx context.Context, page, pageSize int) (*stockcenter.StockListData, error) {
	return s.stockCenter.ListStocks(ctx, page, pageSize)
}

// InboundStockInput is the business inbound payload.
type InboundStockInput struct {
	SKU       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int64  `json:"quantity"`
	Operator  string `json:"operator"`
	Remark    string `json:"remark"`
}

// InboundStock creates or increases stock through stock-center and records an audit log.
func (s *InventoryService) InboundStock(ctx context.Context, input InboundStockInput) (*stockcenter.Stock, error) {
	if input.SKU == "" || input.Warehouse == "" {
		return nil, errors.New("sku and warehouse are required")
	}
	if input.Operator == "" {
		return nil, errors.New("operator is required")
	}
	if input.Quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	existing, err := s.stockCenter.GetStockBySKU(ctx, input.SKU, input.Warehouse)
	if errors.Is(err, stockcenter.ErrStockNotFound) {
		stock, createErr := s.stockCenter.CreateStock(ctx, stockcenter.CreateStockInput{
			SKU:       input.SKU,
			Warehouse: input.Warehouse,
			Quantity:  input.Quantity,
		})
		if createErr != nil {
			return nil, createErr
		}
		_ = s.logRepo.Create(ctx, &model.StockOperationLog{
			SKU:       input.SKU,
			Warehouse: input.Warehouse,
			Action:    "inbound_create",
			Operator:  input.Operator,
			Remark:    input.Remark,
		})
		return stock, nil
	}
	if err != nil {
		return nil, err
	}

	stock, err := s.stockCenter.AdjustStock(ctx, existing.ID, stockcenter.AdjustStockInput{
		Quantity: existing.Quantity + input.Quantity,
	})
	if err != nil {
		return nil, err
	}

	_ = s.logRepo.Create(ctx, &model.StockOperationLog{
		SKU:       input.SKU,
		Warehouse: input.Warehouse,
		Action:    "inbound_adjust",
		Operator:  input.Operator,
		Remark:    input.Remark,
	})
	return stock, nil
}
