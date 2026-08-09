package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tserom/l-project/apps/sales-manage/internal/model"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"github.com/tserom/l-project/apps/sales-manage/internal/repository"
	"gorm.io/gorm"
)

// SalesOrderService coordinates sales order operations.
type SalesOrderService struct {
	db   *gorm.DB
	repo *repository.SalesOrderRepository
}

func NewSalesOrderService(db *gorm.DB, repo *repository.SalesOrderRepository) *SalesOrderService {
	return &SalesOrderService{db: db, repo: repo}
}

// SalesOrderLineInput is create/update line payload.
type SalesOrderLineInput struct {
	ID           string  `json:"id"`
	MaterialName string  `json:"materialName"`
	Spec         string  `json:"spec"`
	Unit         string  `json:"unit"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unitPrice"`
	LineRemark   string  `json:"lineRemark"`
	NeedInvoice  bool    `json:"needInvoice"`
	InvoiceDocID *string `json:"invoiceDocId"`
}

// SalesOrderInput is create/update header payload.
type SalesOrderInput struct {
	OrderNo            string                `json:"orderNo"`
	OrderDate          string                `json:"orderDate"`
	CustomerName       string                `json:"customerName"`
	WarehouseName      string                `json:"warehouseName"`
	DeliveryType       string                `json:"deliveryType"`
	Remark             string                `json:"remark"`
	OutstandingBalance string                `json:"outstandingBalance"`
	Lines              []SalesOrderLineInput `json:"lines"`
}

func (s *SalesOrderService) List(
	ctx context.Context,
	preds []qp.Predicate,
	page, pageSize int,
) ([]model.SalesOrder, int64, error) {
	return s.repo.List(ctx, preds, page, pageSize)
}

func (s *SalesOrderService) GetByID(ctx context.Context, id string) (*model.SalesOrder, error) {
	order, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrSalesOrderNotFound) {
		return nil, ErrOrderNotFound
	}
	return order, err
}

func (s *SalesOrderService) Create(ctx context.Context, input SalesOrderInput) (*model.SalesOrder, error) {
	if err := validateSalesOrderInput(input); err != nil {
		return nil, err
	}
	now := time.Now()
	order := buildOrder(uuid.NewString(), input, now, now, nil)
	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, order.ID)
}

func (s *SalesOrderService) Update(ctx context.Context, id string, input SalesOrderInput) (*model.SalesOrder, error) {
	if err := validateSalesOrderInput(input); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrSalesOrderNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	order := buildOrder(id, input, existing.CreatedAt, time.Now(), existing.Lines)
	if err := s.repo.Replace(ctx, order); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *SalesOrderService) Delete(ctx context.Context, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrSalesOrderNotFound) {
		return ErrOrderNotFound
	}
	if err != nil {
		return err
	}
	for _, line := range existing.Lines {
		if line.InvoiceDocID != nil && *line.InvoiceDocID != "" {
			return ErrOrderHasInvoice
		}
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrSalesOrderNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	return nil
}

// InvoiceCandidate is a flattened pending invoice line.
type InvoiceCandidate struct {
	ID               string  `json:"id"`
	SalesOrderID     string  `json:"salesOrderId"`
	SalesOrderLineID string  `json:"salesOrderLineId"`
	MaterialName     string  `json:"materialName"`
	Spec             string  `json:"spec"`
	Unit             string  `json:"unit"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unitPrice"`
	Amount           float64 `json:"amount"`
	LineRemark       string  `json:"lineRemark,omitempty"`
	OrderNo          string  `json:"orderNo"`
	OrderDate        string  `json:"orderDate"`
	CustomerName     string  `json:"customerName"`
	WarehouseName    string  `json:"warehouseName"`
	DeliveryType     string  `json:"deliveryType"`
}

func (s *SalesOrderService) ListInvoiceCandidates(
	ctx context.Context,
	preds []qp.Predicate,
) ([]InvoiceCandidate, error) {
	orders, err := s.repo.ListForCandidates(ctx, preds)
	if err != nil {
		return nil, err
	}
	out := make([]InvoiceCandidate, 0)
	for _, order := range orders {
		for _, line := range order.Lines {
			if !line.NeedInvoice {
				continue
			}
			if line.InvoiceDocID != nil && *line.InvoiceDocID != "" {
				continue
			}
			out = append(out, InvoiceCandidate{
				ID:               uuid.NewString(),
				SalesOrderID:     order.ID,
				SalesOrderLineID: line.ID,
				MaterialName:     line.MaterialName,
				Spec:             line.Spec,
				Unit:             line.Unit,
				Quantity:         line.Quantity,
				UnitPrice:        line.UnitPrice,
				Amount:           line.Amount,
				LineRemark:       line.LineRemark,
				OrderNo:          order.OrderNo,
				OrderDate:        order.OrderDate,
				CustomerName:     order.CustomerName,
				WarehouseName:    order.WarehouseName,
				DeliveryType:     order.DeliveryType,
			})
		}
	}
	return out, nil
}

func validateSalesOrderInput(input SalesOrderInput) error {
	if strings.TrimSpace(input.CustomerName) == "" {
		return fmt.Errorf("%w: 客户不能为空", ErrValidation)
	}
	if len(input.Lines) == 0 {
		return fmt.Errorf("%w: 至少需要一行明细", ErrValidation)
	}
	for _, line := range input.Lines {
		if line.Quantity < 0 {
			return fmt.Errorf("%w: 数量不能为负数", ErrValidation)
		}
		if line.UnitPrice < 0 {
			return fmt.Errorf("%w: 单价不能为负数", ErrValidation)
		}
	}
	return nil
}

func buildOrder(
	id string,
	input SalesOrderInput,
	createdAt, updatedAt time.Time,
	existingLines []model.SalesOrderLine,
) *model.SalesOrder {
	prevByID := map[string]model.SalesOrderLine{}
	for _, l := range existingLines {
		prevByID[l.ID] = l
	}

	lines := make([]model.SalesOrderLine, 0, len(input.Lines))
	var totalQty, totalAmt float64
	for _, in := range input.Lines {
		lineID := in.ID
		if lineID == "" {
			lineID = uuid.NewString()
		}
		amount := round4(in.Quantity * in.UnitPrice)
		var invID *string
		if in.InvoiceDocID != nil && *in.InvoiceDocID != "" {
			invID = in.InvoiceDocID
		} else if prev, ok := prevByID[lineID]; ok {
			invID = prev.InvoiceDocID
		}
		lines = append(lines, model.SalesOrderLine{
			ID:           lineID,
			SalesOrderID: id,
			MaterialName: in.MaterialName,
			Spec:         in.Spec,
			Unit:         in.Unit,
			Quantity:     in.Quantity,
			UnitPrice:    in.UnitPrice,
			Amount:       amount,
			LineRemark:   in.LineRemark,
			NeedInvoice:  in.NeedInvoice,
			InvoiceDocID: invID,
		})
		totalQty += in.Quantity
		totalAmt += amount
	}

	return &model.SalesOrder{
		ID:                 id,
		OrderNo:            input.OrderNo,
		OrderDate:          input.OrderDate,
		CustomerName:       strings.TrimSpace(input.CustomerName),
		WarehouseName:      input.WarehouseName,
		DeliveryType:       input.DeliveryType,
		Remark:             input.Remark,
		OutstandingBalance: input.OutstandingBalance,
		TotalQuantity:      round4(totalQty),
		TotalAmount:        round4(totalAmt),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		Lines:              lines,
	}
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
