package service

import (
	"context"
	"errors"
	"strings"

	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/pkg/docno"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"gorm.io/gorm"
)

const shipmentDocPrefix = "SS"

// ShipmentService coordinates sales shipment business operations.
type ShipmentService struct {
	db           *gorm.DB
	repo         *repository.SalesShipmentRepository
	salesRepo    *repository.SalesOrderRepository
	logRepo      *repository.OperationLogRepository
	center       StockMover
}

// NewShipmentService creates a ShipmentService.
func NewShipmentService(
	db *gorm.DB,
	repo *repository.SalesShipmentRepository,
	salesRepo *repository.SalesOrderRepository,
	logRepo *repository.OperationLogRepository,
	center StockMover,
) *ShipmentService {
	return &ShipmentService{
		db:        db,
		repo:      repo,
		salesRepo: salesRepo,
		logRepo:   logRepo,
		center:    center,
	}
}

// CreateShipmentInput is the payload for creating a sales shipment.
type CreateShipmentInput struct {
	SalesOrderID uint64           `json:"salesOrderId"`
	DocDate      string           `json:"docDate"`
	Operator     string           `json:"operator"`
	Remark       string           `json:"remark"`
	Warehouse    string           `json:"warehouse"`
	Lines        []OrderLineInput `json:"lines"`
}

// UpdateShipmentInput is the payload for updating a draft sales shipment.
type UpdateShipmentInput struct {
	DocDate  string           `json:"docDate"`
	Operator string           `json:"operator"`
	Remark   string           `json:"remark"`
	Lines    []OrderLineInput `json:"lines"`
}

// List returns paginated sales shipments.
func (s *ShipmentService) List(ctx context.Context, page, pageSize int) ([]model.SalesShipment, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// ListBySalesOrderID returns shipments for a sales order.
func (s *ShipmentService) ListBySalesOrderID(ctx context.Context, salesOrderID uint64) ([]model.SalesShipment, error) {
	if _, err := s.salesRepo.GetByID(ctx, salesOrderID); err != nil {
		return nil, err
	}
	return s.repo.ListBySalesOrderID(ctx, salesOrderID)
}

// GetByID returns a sales shipment with lines.
func (s *ShipmentService) GetByID(ctx context.Context, id uint64) (*model.SalesShipment, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateFromSalesOrder creates a draft shipment from a confirmed sales order.
func (s *ShipmentService) CreateFromSalesOrder(ctx context.Context, salesOrderID uint64, input CreateShipmentInput) (*model.SalesShipment, error) {
	input.SalesOrderID = salesOrderID
	return s.create(ctx, input)
}

// Create creates a draft sales shipment linked to a confirmed sales order.
func (s *ShipmentService) Create(ctx context.Context, input CreateShipmentInput) (*model.SalesShipment, error) {
	return s.create(ctx, input)
}

func (s *ShipmentService) create(ctx context.Context, input CreateShipmentInput) (*model.SalesShipment, error) {
	if strings.TrimSpace(input.Operator) == "" {
		return nil, ErrOperatorRequired
	}
	if input.SalesOrderID == 0 {
		return nil, errors.New("salesOrderId is required")
	}

	order, err := s.salesRepo.GetByID(ctx, input.SalesOrderID)
	if err != nil {
		return nil, err
	}
	if order.Status != model.DocStatusConfirmed {
		return nil, ErrSalesOrderNotConfirmed
	}

	var lines []model.SalesShipmentLine
	if len(input.Lines) > 0 {
		if err := validateShipmentLines(input.Lines); err != nil {
			return nil, err
		}
		lines = toSalesShipmentLines(input.Lines)
	} else {
		if len(order.Lines) == 0 {
			return nil, ErrLinesRequired
		}
		warehouse := strings.TrimSpace(input.Warehouse)
		if warehouse == "" {
			return nil, ErrWarehouseRequired
		}
		lines = copySalesOrderLinesToShipment(order.Lines, warehouse)
	}

	docDate, err := parseDocDate(input.DocDate)
	if err != nil {
		return nil, err
	}

	docNo, err := docno.Generate(ctx, s.db, shipmentDocPrefix)
	if err != nil {
		return nil, err
	}

	shipment := &model.SalesShipment{
		DocNo:        docNo,
		DocDate:      docDate,
		Status:       model.DocStatusDraft,
		SalesOrderID: input.SalesOrderID,
		Operator:     input.Operator,
		Remark:       input.Remark,
		Lines:        lines,
	}
	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, shipment.ID)
}

// Update updates a draft sales shipment.
func (s *ShipmentService) Update(ctx context.Context, id uint64, input UpdateShipmentInput) (*model.SalesShipment, error) {
	if strings.TrimSpace(input.Operator) == "" {
		return nil, ErrOperatorRequired
	}
	if err := validateShipmentLines(input.Lines); err != nil {
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

	shipment := &model.SalesShipment{
		ID:           id,
		DocNo:        existing.DocNo,
		DocDate:      docDate,
		Status:       model.DocStatusDraft,
		SalesOrderID: existing.SalesOrderID,
		Operator:     input.Operator,
		Remark:       input.Remark,
		Lines:        toSalesShipmentLines(input.Lines),
	}
	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a draft sales shipment.
func (s *ShipmentService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

// Confirm posts each line to stock-center and marks the shipment confirmed.
func (s *ShipmentService) Confirm(ctx context.Context, id uint64) (*model.SalesShipment, error) {
	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if shipment.Status == model.DocStatusConfirmed {
		return nil, ErrAlreadyConfirmed
	}
	if shipment.Status != model.DocStatusDraft {
		return nil, errors.New("invalid document status")
	}
	if len(shipment.Lines) == 0 {
		return nil, ErrLinesRequired
	}

	for _, line := range shipment.Lines {
		_, err := s.center.OutboundStock(ctx, stockcenter.StockMovementInput{
			MaterialID: line.MaterialID,
			BatchID:    line.BatchID,
			Warehouse:  line.Warehouse,
			WeightKg:   line.WeightKg.String(),
			Quantity:   line.Quantity.String(),
			RefType:    "sale",
			RefNo:      shipment.DocNo,
			Remark:     shipment.Remark,
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
		DocType:  "sales_shipment",
		DocNo:    shipment.DocNo,
		Action:   "confirm",
		Operator: shipment.Operator,
		Remark:   shipment.Remark,
	})

	return s.repo.GetByID(ctx, id)
}
