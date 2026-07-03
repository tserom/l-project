package service

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/pkg/docno"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"gorm.io/gorm"
)

const processingDocPrefix = "JG"

// ProcessingOrderService coordinates processing order business operations.
type ProcessingOrderService struct {
	db      *gorm.DB
	repo    *repository.ProcessingOrderRepository
	logRepo *repository.OperationLogRepository
	center  StockMover
}

// NewProcessingOrderService creates a ProcessingOrderService.
func NewProcessingOrderService(
	db *gorm.DB,
	repo *repository.ProcessingOrderRepository,
	logRepo *repository.OperationLogRepository,
	center StockMover,
) *ProcessingOrderService {
	return &ProcessingOrderService{
		db:      db,
		repo:    repo,
		logRepo: logRepo,
		center:  center,
	}
}

// ProcessingPickLineInput is a raw material pick line on create/update.
type ProcessingPickLineInput struct {
	MaterialID uint64          `json:"materialId"`
	BatchID    uint64          `json:"batchId"`
	Warehouse  string          `json:"warehouse"`
	WeightKg   decimal.Decimal `json:"weightKg"`
}

// ProcessingFinishLineInput is a finished goods line on create/update.
type ProcessingFinishLineInput struct {
	MaterialID uint64           `json:"materialId"`
	BatchID    uint64           `json:"batchId"`
	Warehouse  string           `json:"warehouse"`
	Quantity   decimal.Decimal  `json:"quantity"`
	WeightKg   *decimal.Decimal `json:"weightKg,omitempty"`
}

// CreateProcessingOrderInput is the payload for creating a processing order.
type CreateProcessingOrderInput struct {
	DocDate     string                      `json:"docDate"`
	Operator    string                      `json:"operator"`
	Remark      string                      `json:"remark"`
	PickLines   []ProcessingPickLineInput   `json:"pickLines"`
	FinishLines []ProcessingFinishLineInput `json:"finishLines"`
}

// UpdateProcessingOrderInput is the payload for updating a draft processing order.
type UpdateProcessingOrderInput struct {
	DocDate     string                      `json:"docDate"`
	Operator    string                      `json:"operator"`
	Remark      string                      `json:"remark"`
	PickLines   []ProcessingPickLineInput   `json:"pickLines"`
	FinishLines []ProcessingFinishLineInput `json:"finishLines"`
}

// List returns paginated processing orders.
func (s *ProcessingOrderService) List(ctx context.Context, page, pageSize int) ([]model.ProcessingOrder, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// GetByID returns a processing order with pick and finish lines.
func (s *ProcessingOrderService) GetByID(ctx context.Context, id uint64) (*model.ProcessingOrder, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a draft processing order with an auto-generated doc number.
func (s *ProcessingOrderService) Create(ctx context.Context, input CreateProcessingOrderInput) (*model.ProcessingOrder, error) {
	if err := validateProcessingOrderInput(input.Operator, input.PickLines, input.FinishLines); err != nil {
		return nil, err
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	docNo, err := docno.Generate(ctx, s.db, processingDocPrefix)
	if err != nil {
		return nil, err
	}

	order := &model.ProcessingOrder{
		DocNo:       docNo,
		DocDate:     docDate,
		Status:      model.DocStatusDraft,
		Operator:    input.Operator,
		Remark:      input.Remark,
		PickLines:   toProcessingPickLines(input.PickLines),
		FinishLines: toProcessingFinishLines(input.FinishLines),
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

// Update updates a draft processing order.
func (s *ProcessingOrderService) Update(ctx context.Context, id uint64, input UpdateProcessingOrderInput) (*model.ProcessingOrder, error) {
	if err := validateProcessingOrderInput(input.Operator, input.PickLines, input.FinishLines); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.Status != model.DocStatusDraft {
		return nil, ErrAlreadyConfirmed
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	order := &model.ProcessingOrder{
		ID:          id,
		DocNo:       existing.DocNo,
		DocDate:     docDate,
		Status:      model.DocStatusDraft,
		Operator:    input.Operator,
		Remark:      input.Remark,
		PickLines:   toProcessingPickLines(input.PickLines),
		FinishLines: toProcessingFinishLines(input.FinishLines),
	}
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a draft processing order.
func (s *ProcessingOrderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// Confirm outbound pick lines, inbound finish lines, calculates loss, and marks confirmed.
func (s *ProcessingOrderService) Confirm(ctx context.Context, id uint64) (*model.ProcessingOrder, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status == model.DocStatusConfirmed {
		return nil, ErrAlreadyConfirmed
	}
	if order.Status != model.DocStatusDraft {
		return nil, errors.New("invalid document status")
	}
	if len(order.PickLines) == 0 || len(order.FinishLines) == 0 {
		return nil, ErrLinesRequired
	}

	totalPick := decimal.Zero
	for _, line := range order.PickLines {
		_, err := s.center.OutboundStock(ctx, stockcenter.StockMovementInput{
			MaterialID: line.MaterialID,
			BatchID:    line.BatchID,
			Warehouse:  line.Warehouse,
			WeightKg:   line.WeightKg.String(),
			Quantity:   decimal.Zero.String(),
			RefType:    "processing",
			RefNo:      order.DocNo,
			Remark:     order.Remark,
		})
		if err != nil {
			return nil, wrapCenterErr(err)
		}
		totalPick = totalPick.Add(line.WeightKg)
	}

	totalFinishWeight := decimal.Zero
	for _, line := range order.FinishLines {
		weightStr := decimal.Zero.String()
		if line.WeightKg != nil {
			weightStr = line.WeightKg.String()
			totalFinishWeight = totalFinishWeight.Add(*line.WeightKg)
		}

		_, err := s.center.InboundStock(ctx, stockcenter.StockMovementInput{
			MaterialID: line.MaterialID,
			BatchID:    line.BatchID,
			Warehouse:  line.Warehouse,
			WeightKg:   weightStr,
			Quantity:   line.Quantity.String(),
			RefType:    "processing",
			RefNo:      order.DocNo,
			Remark:     order.Remark,
		})
		if err != nil {
			return nil, wrapCenterErr(err)
		}
	}

	lossWeightKg := totalPick.Sub(totalFinishWeight)

	if err := s.repo.ConfirmStatus(ctx, id, lossWeightKg); err != nil {
		if errors.Is(err, repository.ErrNotDraft) {
			return nil, ErrAlreadyConfirmed
		}
		return nil, err
	}

	_ = s.logRepo.Create(ctx, &model.StockOperationLog{
		DocType:  "processing",
		DocNo:    order.DocNo,
		Action:   "confirm",
		Operator: order.Operator,
		Remark:   order.Remark,
	})

	return s.repo.GetByID(ctx, id)
}
