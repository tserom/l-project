package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/pkg/docno"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"gorm.io/gorm"
)

const outboundDocPrefix = "OUT"

// OutboundOrderService coordinates outbound order business operations.
type OutboundOrderService struct {
	db      *gorm.DB
	repo    *repository.OutboundOrderRepository
	logRepo *repository.OperationLogRepository
	center  StockMover
}

// NewOutboundOrderService creates an OutboundOrderService.
func NewOutboundOrderService(
	db *gorm.DB,
	repo *repository.OutboundOrderRepository,
	logRepo *repository.OperationLogRepository,
	center StockMover,
) *OutboundOrderService {
	return &OutboundOrderService{
		db:      db,
		repo:    repo,
		logRepo: logRepo,
		center:  center,
	}
}

// CreateOutboundOrderInput is the payload for creating an outbound order.
type CreateOutboundOrderInput struct {
	DocDate  string           `json:"docDate"`
	Operator string           `json:"operator"`
	Remark   string           `json:"remark"`
	Lines    []OrderLineInput `json:"lines"`
}

// UpdateOutboundOrderInput is the payload for updating a draft outbound order.
type UpdateOutboundOrderInput struct {
	DocDate  string           `json:"docDate"`
	Operator string           `json:"operator"`
	Remark   string           `json:"remark"`
	Lines    []OrderLineInput `json:"lines"`
}

// List returns paginated outbound orders.
func (s *OutboundOrderService) List(ctx context.Context, page, pageSize int) ([]model.OutboundOrder, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// GetByID returns an outbound order with lines.
func (s *OutboundOrderService) GetByID(ctx context.Context, id uint64) (*model.OutboundOrder, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a draft outbound order with an auto-generated doc number.
func (s *OutboundOrderService) Create(ctx context.Context, input CreateOutboundOrderInput) (*model.OutboundOrder, error) {
	if err := validateOrderInput(input.Operator, input.Lines); err != nil {
		return nil, err
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	docNo, err := docno.Generate(ctx, s.db, outboundDocPrefix)
	if err != nil {
		return nil, err
	}

	order := &model.OutboundOrder{
		DocNo:    docNo,
		DocDate:  docDate,
		Status:   model.DocStatusDraft,
		Operator: input.Operator,
		Remark:   input.Remark,
		Lines:    toOutboundLines(input.Lines),
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

// Update updates a draft outbound order.
func (s *OutboundOrderService) Update(ctx context.Context, id uint64, input UpdateOutboundOrderInput) (*model.OutboundOrder, error) {
	if err := validateOrderInput(input.Operator, input.Lines); err != nil {
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

	order := &model.OutboundOrder{
		ID:       id,
		DocNo:    existing.DocNo,
		DocDate:  docDate,
		Status:   model.DocStatusDraft,
		Operator: input.Operator,
		Remark:   input.Remark,
		Lines:    toOutboundLines(input.Lines),
	}
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a draft outbound order.
func (s *OutboundOrderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// Confirm posts each line to stock-center and marks the order confirmed.
func (s *OutboundOrderService) Confirm(ctx context.Context, id uint64) (*model.OutboundOrder, error) {
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
	if len(order.Lines) == 0 {
		return nil, ErrLinesRequired
	}

	for _, line := range order.Lines {
		_, err := s.center.OutboundStock(ctx, stockcenter.StockMovementInput{
			MaterialID: line.MaterialID,
			BatchID:    line.BatchID,
			Warehouse:  line.Warehouse,
			WeightKg:   line.WeightKg.String(),
			Quantity:   line.Quantity.String(),
			RefType:    "outbound",
			RefNo:      order.DocNo,
			Remark:     order.Remark,
		})
		if err != nil {
			return nil, wrapCenterErr(err)
		}
	}

	if err := s.repo.ConfirmStatus(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotDraft) {
			return nil, ErrAlreadyConfirmed
		}
		return nil, err
	}

	_ = s.logRepo.Create(ctx, &model.StockOperationLog{
		DocType:  "outbound",
		DocNo:    order.DocNo,
		Action:   "confirm",
		Operator: order.Operator,
		Remark:   order.Remark,
	})

	return s.repo.GetByID(ctx, id)
}
