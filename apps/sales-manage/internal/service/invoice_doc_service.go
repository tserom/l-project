package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tserom/l-project/apps/sales-manage/internal/model"
	"github.com/tserom/l-project/apps/sales-manage/internal/pkg/qp"
	"github.com/tserom/l-project/apps/sales-manage/internal/repository"
	"gorm.io/gorm"
)

// InvoiceDocService coordinates invoice document operations.
type InvoiceDocService struct {
	db        *gorm.DB
	invoiceRepo *repository.InvoiceDocRepository
	orderRepo   *repository.SalesOrderRepository
}

func NewInvoiceDocService(
	db *gorm.DB,
	invoiceRepo *repository.InvoiceDocRepository,
	orderRepo *repository.SalesOrderRepository,
) *InvoiceDocService {
	return &InvoiceDocService{db: db, invoiceRepo: invoiceRepo, orderRepo: orderRepo}
}

// CreateInvoiceDocInput is the create payload.
type CreateInvoiceDocInput struct {
	FilterCustomerName string                     `json:"filterCustomerName"`
	FilterDateFrom     string                     `json:"filterDateFrom"`
	FilterDateTo       string                     `json:"filterDateTo"`
	FilterOrderNos     []string                   `json:"filterOrderNos"`
	Lines              []CreateInvoiceDocLineInput `json:"lines"`
}

// CreateInvoiceDocLineInput is one invoice line from the client.
type CreateInvoiceDocLineInput struct {
	ID               string  `json:"id"`
	SalesOrderID     string  `json:"salesOrderId"`
	SalesOrderLineID string  `json:"salesOrderLineId"`
	MaterialName     string  `json:"materialName"`
	Spec             string  `json:"spec"`
	Unit             string  `json:"unit"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unitPrice"`
	Amount           float64 `json:"amount"`
	LineRemark       string  `json:"lineRemark"`
	OrderNo          string  `json:"orderNo"`
	OrderDate        string  `json:"orderDate"`
	CustomerName     string  `json:"customerName"`
	WarehouseName    string  `json:"warehouseName"`
	DeliveryType     string  `json:"deliveryType"`
}

func (s *InvoiceDocService) List(
	ctx context.Context,
	preds []qp.Predicate,
	page, pageSize int,
) ([]model.InvoiceDoc, int64, error) {
	return s.invoiceRepo.List(ctx, preds, page, pageSize)
}

func (s *InvoiceDocService) GetByID(ctx context.Context, id string) (*model.InvoiceDoc, error) {
	doc, err := s.invoiceRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrInvoiceDocNotFound) {
		return nil, ErrInvoiceNotFound
	}
	return doc, err
}

func (s *InvoiceDocService) Create(ctx context.Context, input CreateInvoiceDocInput) (*model.InvoiceDoc, error) {
	if len(input.Lines) == 0 {
		return nil, ErrInvoiceLinesEmpty
	}

	docID := uuid.NewString()
	now := time.Now()
	lines := make([]model.InvoiceDocLine, 0, len(input.Lines))
	var totalQty, totalAmt float64

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, in := range input.Lines {
			var order model.SalesOrder
			if err := tx.Preload("Lines").First(&order, "id = ?", in.SalesOrderID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("销售单不存在：%s", in.OrderNo)
				}
				return err
			}
			var src *model.SalesOrderLine
			for i := range order.Lines {
				if order.Lines[i].ID == in.SalesOrderLineID {
					src = &order.Lines[i]
					break
				}
			}
			if src == nil {
				return fmt.Errorf("明细不存在：%s", in.MaterialName)
			}
			if !src.NeedInvoice {
				return fmt.Errorf("明细未勾选需开票：%s / %s", in.OrderNo, in.MaterialName)
			}
			if src.InvoiceDocID != nil && *src.InvoiceDocID != "" {
				return fmt.Errorf("明细已开票，请重新加载：%s / %s", in.OrderNo, in.MaterialName)
			}

			lineID := in.ID
			if lineID == "" {
				lineID = uuid.NewString()
			}
			amount := round4(in.Quantity * in.UnitPrice)
			lines = append(lines, model.InvoiceDocLine{
				ID:               lineID,
				InvoiceDocID:     docID,
				SalesOrderID:     in.SalesOrderID,
				SalesOrderLineID: in.SalesOrderLineID,
				MaterialName:     in.MaterialName,
				Spec:             in.Spec,
				Unit:             in.Unit,
				Quantity:         in.Quantity,
				UnitPrice:        in.UnitPrice,
				Amount:           amount,
				LineRemark:       in.LineRemark,
				OrderNo:          in.OrderNo,
				OrderDate:        in.OrderDate,
				CustomerName:     in.CustomerName,
				WarehouseName:    in.WarehouseName,
				DeliveryType:     in.DeliveryType,
			})
			totalQty += in.Quantity
			totalAmt += amount

			if err := tx.Model(&model.SalesOrderLine{}).
				Where("id = ? AND (invoice_doc_id IS NULL OR invoice_doc_id = '')", in.SalesOrderLineID).
				Updates(map[string]interface{}{
					"invoice_doc_id": docID,
				}).Error; err != nil {
				return err
			}
			// Ensure the conditional update actually attached.
			var check model.SalesOrderLine
			if err := tx.First(&check, "id = ?", in.SalesOrderLineID).Error; err != nil {
				return err
			}
			if check.InvoiceDocID == nil || *check.InvoiceDocID != docID {
				return fmt.Errorf("明细已开票，请重新加载：%s / %s", in.OrderNo, in.MaterialName)
			}
			if err := tx.Model(&model.SalesOrder{}).
				Where("id = ?", in.SalesOrderID).
				Update("updated_at", now).Error; err != nil {
				return err
			}
		}

		doc := &model.InvoiceDoc{
			ID:                 docID,
			InvoiceNo:          "KP" + generateInvoiceNo(now),
			Status:             model.InvoiceStatusSaved,
			FilterCustomerName: strings.TrimSpace(input.FilterCustomerName),
			FilterDateFrom:     input.FilterDateFrom,
			FilterDateTo:       input.FilterDateTo,
			FilterOrderNos:     model.StringList(input.FilterOrderNos),
			TotalQuantity:      round4(totalQty),
			TotalAmount:        round4(totalAmt),
			CreatedAt:          now,
			UpdatedAt:          now,
			Lines:              lines,
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(doc).Error
	})
	if err != nil {
		return nil, err
	}
	return s.invoiceRepo.GetByID(ctx, docID)
}

func (s *InvoiceDocService) Void(ctx context.Context, id string) (*model.InvoiceDoc, error) {
	doc, err := s.invoiceRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrInvoiceDocNotFound) {
		return nil, ErrInvoiceNotFound
	}
	if err != nil {
		return nil, err
	}
	if doc.Status == model.InvoiceStatusVoided {
		return nil, ErrInvoiceAlreadyVoid
	}

	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, line := range doc.Lines {
			var src model.SalesOrderLine
			if err := tx.First(&src, "id = ?", line.SalesOrderLineID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			if src.InvoiceDocID == nil || *src.InvoiceDocID != id {
				continue
			}
			if err := tx.Model(&src).Update("invoice_doc_id", nil).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.SalesOrder{}).
				Where("id = ?", line.SalesOrderID).
				Update("updated_at", now).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.InvoiceDoc{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":     model.InvoiceStatusVoided,
			"updated_at": now,
			"voided_at":  now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.invoiceRepo.GetByID(ctx, id)
}

func generateInvoiceNo(now time.Time) string {
	// YYYYMMDD + unix millis last 4 digits (aligns with front generateOrderNo spirit)
	stamp := now.Format("20060102")
	ms := now.UnixMilli() % 10000
	return fmt.Sprintf("%s%04d", stamp, ms)
}
