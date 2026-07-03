package service

import (
	"context"
	"errors"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/pkg/docno"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"gorm.io/gorm"
)

const salesOrderDocPrefix = "SO"

// SalesOrderService coordinates sales order business operations.
type SalesOrderService struct {
	db      *gorm.DB
	repo    *repository.SalesOrderRepository
	logRepo *repository.OperationLogRepository
}

// NewSalesOrderService creates a SalesOrderService.
func NewSalesOrderService(
	db *gorm.DB,
	repo *repository.SalesOrderRepository,
	logRepo *repository.OperationLogRepository,
) *SalesOrderService {
	return &SalesOrderService{
		db:      db,
		repo:    repo,
		logRepo: logRepo,
	}
}

// CreateSalesOrderInput is the payload for creating a sales order.
type CreateSalesOrderInput struct {
	DocDate      string              `json:"docDate"`
	CustomerName string              `json:"customerName"`
	Operator     string              `json:"operator"`
	Remark       string              `json:"remark"`
	Lines        []SalesOrderLineInput `json:"lines"`
}

// UpdateSalesOrderInput is the payload for updating a draft sales order.
type UpdateSalesOrderInput struct {
	DocDate      string              `json:"docDate"`
	CustomerName string              `json:"customerName"`
	Operator     string              `json:"operator"`
	Remark       string              `json:"remark"`
	Lines        []SalesOrderLineInput `json:"lines"`
}

// List returns paginated sales orders.
func (s *SalesOrderService) List(ctx context.Context, page, pageSize int) ([]model.SalesOrder, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// GetByID returns a sales order with lines.
func (s *SalesOrderService) GetByID(ctx context.Context, id uint64) (*model.SalesOrder, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a draft sales order with an auto-generated doc number.
func (s *SalesOrderService) Create(ctx context.Context, input CreateSalesOrderInput) (*model.SalesOrder, error) {
	if err := validateSalesOrderInput(input.CustomerName, input.Operator, input.Lines); err != nil {
		return nil, err
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	docNo, err := docno.Generate(ctx, s.db, salesOrderDocPrefix)
	if err != nil {
		return nil, err
	}

	order := &model.SalesOrder{
		DocNo:        docNo,
		DocDate:      docDate,
		Status:       model.DocStatusDraft,
		CustomerName: input.CustomerName,
		Operator:     input.Operator,
		Remark:       input.Remark,
		Lines:        toSalesOrderLines(input.Lines),
	}
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

// Update updates a draft sales order.
func (s *SalesOrderService) Update(ctx context.Context, id uint64, input UpdateSalesOrderInput) (*model.SalesOrder, error) {
	if err := validateSalesOrderInput(input.CustomerName, input.Operator, input.Lines); err != nil {
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

	order := &model.SalesOrder{
		ID:           id,
		DocNo:        existing.DocNo,
		DocDate:      docDate,
		Status:       model.DocStatusDraft,
		CustomerName: input.CustomerName,
		Operator:     input.Operator,
		Remark:       input.Remark,
		Lines:        toSalesOrderLines(input.Lines),
	}
	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a draft sales order.
func (s *SalesOrderService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// Confirm marks the sales order confirmed without calling stock-center.
func (s *SalesOrderService) Confirm(ctx context.Context, id uint64) (*model.SalesOrder, error) {
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

	if err := s.repo.ConfirmStatus(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotDraft) {
			return nil, ErrAlreadyConfirmed
		}
		return nil, err
	}

	_ = s.logRepo.Create(ctx, &model.StockOperationLog{
		DocType:  "sales_order",
		DocNo:    order.DocNo,
		Action:   "confirm",
		Operator: order.Operator,
		Remark:   order.Remark,
	})

	return s.repo.GetByID(ctx, id)
}
