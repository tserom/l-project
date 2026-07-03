package service

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/pkg/qp"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"gorm.io/gorm"
)

const defaultWarehouse = "DEFAULT"

// StockMovementInput is the payload for inbound and outbound stock operations.
type StockMovementInput struct {
	MaterialID uint64          `json:"materialId"`
	BatchID    uint64          `json:"batchId"`
	Warehouse  string          `json:"warehouse"`
	WeightKg   decimal.Decimal `json:"weightKg"`
	Quantity   decimal.Decimal `json:"quantity"`
	RefType    string          `json:"refType"`
	RefNo      string          `json:"refNo"`
	Remark     string          `json:"remark"`
}

// StockBalanceService coordinates stock balance and ledger operations.
type StockBalanceService struct {
	balanceRepo *repository.StockBalanceRepository
	ledgerRepo  *repository.StockLedgerRepository
}

// NewStockBalanceService creates a StockBalanceService.
func NewStockBalanceService(balanceRepo *repository.StockBalanceRepository, ledgerRepo *repository.StockLedgerRepository) *StockBalanceService {
	return &StockBalanceService{
		balanceRepo: balanceRepo,
		ledgerRepo:  ledgerRepo,
	}
}

// ListStocks returns paginated stock balances.
func (s *StockBalanceService) ListStocks(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.StockBalance, int64, error) {
	return s.balanceRepo.List(ctx, page, pageSize, preds)
}

// QueryStock returns a balance for the given material, batch, and warehouse.
func (s *StockBalanceService) QueryStock(ctx context.Context, materialID, batchID uint64, warehouse string) (*model.StockBalance, error) {
	if materialID == 0 {
		return nil, errors.New("materialId is required")
	}
	if batchID == 0 {
		return nil, errors.New("batchId is required")
	}
	if warehouse == "" {
		warehouse = defaultWarehouse
	}
	return s.balanceRepo.GetByKey(ctx, nil, materialID, batchID, warehouse, 0)
}

// Inbound increases stock balance and records a ledger entry in one transaction.
func (s *StockBalanceService) Inbound(ctx context.Context, input StockMovementInput) (*model.StockBalance, error) {
	if err := validateMovementInput(input); err != nil {
		return nil, err
	}

	refType, err := resolveRefType(input.RefType, model.RefInbound)
	if err != nil {
		return nil, err
	}
	if input.RefNo == "" {
		return nil, errors.New("refNo is required")
	}

	warehouse := input.Warehouse
	if warehouse == "" {
		warehouse = defaultWarehouse
	}

	var balance *model.StockBalance
	err = s.balanceRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error
		balance, txErr = s.balanceRepo.Add(ctx, tx, input.MaterialID, input.BatchID, warehouse, 0, input.WeightKg, input.Quantity)
		if txErr != nil {
			return txErr
		}

		ledger := &model.StockLedger{
			MaterialID:    input.MaterialID,
			BatchID:       input.BatchID,
			Warehouse:     warehouse,
			DeltaWeightKg: input.WeightKg,
			DeltaQuantity: input.Quantity,
			RefType:       refType,
			RefNo:         input.RefNo,
			Remark:        input.Remark,
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})
	if err != nil {
		return nil, err
	}
	return balance, nil
}

// Outbound decreases stock balance and records a ledger entry in one transaction.
func (s *StockBalanceService) Outbound(ctx context.Context, input StockMovementInput) (*model.StockBalance, error) {
	if err := validateMovementInput(input); err != nil {
		return nil, err
	}

	refType, err := resolveRefType(input.RefType, model.RefOutbound)
	if err != nil {
		return nil, err
	}
	if input.RefNo == "" {
		return nil, errors.New("refNo is required")
	}

	warehouse := input.Warehouse
	if warehouse == "" {
		warehouse = defaultWarehouse
	}

	var balance *model.StockBalance
	err = s.balanceRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var txErr error
		balance, txErr = s.balanceRepo.Subtract(ctx, tx, input.MaterialID, input.BatchID, warehouse, 0, input.WeightKg, input.Quantity)
		if txErr != nil {
			return txErr
		}

		ledger := &model.StockLedger{
			MaterialID:    input.MaterialID,
			BatchID:       input.BatchID,
			Warehouse:     warehouse,
			DeltaWeightKg: input.WeightKg.Neg(),
			DeltaQuantity: input.Quantity.Neg(),
			RefType:       refType,
			RefNo:         input.RefNo,
			Remark:        input.Remark,
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})
	if err != nil {
		return nil, err
	}
	return balance, nil
}

// ListLedger returns paginated ledger entries with optional qp filters.
func (s *StockBalanceService) ListLedger(ctx context.Context, page, pageSize int, preds []qp.Predicate) ([]model.StockLedger, int64, error) {
	return s.ledgerRepo.List(ctx, page, pageSize, preds)
}

func validateMovementInput(input StockMovementInput) error {
	if input.MaterialID == 0 {
		return errors.New("materialId is required")
	}
	if input.BatchID == 0 {
		return errors.New("batchId is required")
	}
	if input.WeightKg.IsNegative() || input.Quantity.IsNegative() {
		return errors.New("weightKg and quantity must not be negative")
	}
	if input.WeightKg.IsZero() && input.Quantity.IsZero() {
		return errors.New("weightKg or quantity must be greater than zero")
	}
	return nil
}

func resolveRefType(value string, fallback model.LedgerRefType) (model.LedgerRefType, error) {
	if value == "" {
		return fallback, nil
	}
	refType := model.LedgerRefType(value)
	switch refType {
	case model.RefInbound, model.RefOutbound, model.RefProcessing, model.RefSale, model.RefAdjust:
		return refType, nil
	default:
		return "", errors.New("invalid refType")
	}
}
