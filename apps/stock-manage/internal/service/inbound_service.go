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

const inboundDocPrefix = "IN"

// InboundOrderService coordinates inbound order business operations.
type InboundOrderService struct {
	db      *gorm.DB
	repo    *repository.InboundOrderRepository
	logRepo *repository.OperationLogRepository
	center  StockMover
}

// NewInboundOrderService creates an InboundOrderService.
func NewInboundOrderService(
	db *gorm.DB,
	repo *repository.InboundOrderRepository,
	logRepo *repository.OperationLogRepository,
	center StockMover,
) *InboundOrderService {
	return &InboundOrderService{
		db:      db,
		repo:    repo,
		logRepo: logRepo,
		center:  center,
	}
}

// CreateInboundOrderInput is the payload for creating an inbound order.
type CreateInboundOrderInput struct {
	DocDate  string           `json:"docDate"`
	Operator string           `json:"operator"`
	Remark   string           `json:"remark"`
	Lines    []OrderLineInput `json:"lines"`
}

// UpdateInboundOrderInput is the payload for updating a draft inbound order.
type UpdateInboundOrderInput struct {
	DocDate  string           `json:"docDate"`
	Operator string           `json:"operator"`
	Remark   string           `json:"remark"`
	Lines    []OrderLineInput `json:"lines"`
}

// List returns paginated inbound orders.
func (s *InboundOrderService) List(ctx context.Context, page, pageSize int) ([]model.InboundOrder, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// GetByID returns an inbound order with lines.
func (s *InboundOrderService) GetByID(ctx context.Context, id uint64) (*model.InboundOrder, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a draft inbound order with an auto-generated doc number.
func (s *InboundOrderService) Create(ctx context.Context, input CreateInboundOrderInput) (*model.InboundOrder, error) {
	if err := validateOrderInput(input.Operator, input.Lines); err != nil {
		return nil, err
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	docNo, err := docno.Generate(ctx, s.db, inboundDocPrefix)
	if err != nil {
		return nil, err
	}

	order := &model.InboundOrder{
		DocNo:    docNo,
		DocDate:  docDate,
		Status:   model.DocStatusDraft,
		Operator: input.Operator,
		Remark:   input.Remark,
		Lines:    toInboundLines(input.Lines),
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

// Update updates a draft inbound order.
func (s *InboundOrderService) Update(ctx context.Context, id uint64, input UpdateInboundOrderInput) (*model.InboundOrder, error) {
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

	order := &model.InboundOrder{
		ID:       id,
		DocNo:    existing.DocNo,
		DocDate:  docDate,
		Status:   model.DocStatusDraft,
		Operator: input.Operator,
		Remark:   input.Remark,
		Lines:    toInboundLines(input.Lines),
	}
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a draft inbound order.
func (s *InboundOrderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// Confirm posts each line to stock-center and marks the order confirmed.
func (s *InboundOrderService) Confirm(ctx context.Context, id uint64) (*model.InboundOrder, error) {
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
		_, err := s.center.InboundStock(ctx, stockcenter.StockMovementInput{
			MaterialID: line.MaterialID,
			BatchID:    line.BatchID,
			Warehouse:  line.Warehouse,
			WeightKg:   line.WeightKg.String(),
			Quantity:   line.Quantity.String(),
			RefType:    "inbound",
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
		DocType:  "inbound",
		DocNo:    order.DocNo,
		Action:   "confirm",
		Operator: order.Operator,
		Remark:   order.Remark,
	})

	return s.repo.GetByID(ctx, id)
}
