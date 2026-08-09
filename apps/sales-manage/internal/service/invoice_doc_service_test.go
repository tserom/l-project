package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tserom/l-project/apps/sales-manage/internal/service"
)

func TestInvoiceCreateConflictRollback(t *testing.T) {
	orderSvc, invoiceSvc, _ := setupOrderSvc(t)
	ctx := context.Background()

	order, err := orderSvc.Create(ctx, sampleInput(true))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	line := order.Lines[0]
	payload := service.CreateInvoiceDocInput{
		Lines: []service.CreateInvoiceDocLineInput{{
			SalesOrderID:     order.ID,
			SalesOrderLineID: line.ID,
			MaterialName:     line.MaterialName,
			Unit:             line.Unit,
			Quantity:         line.Quantity,
			UnitPrice:        line.UnitPrice,
			OrderNo:          order.OrderNo,
			OrderDate:        order.OrderDate,
			CustomerName:     order.CustomerName,
			WarehouseName:    order.WarehouseName,
			DeliveryType:     order.DeliveryType,
		}},
	}
	if _, err := invoiceSvc.Create(ctx, payload); err != nil {
		t.Fatalf("first invoice: %v", err)
	}
	_, err = invoiceSvc.Create(ctx, payload)
	if err == nil || !strings.Contains(err.Error(), "已开票") {
		t.Fatalf("want conflict, got %v", err)
	}

	list, total, err := invoiceSvc.List(ctx, nil, 1, 20)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected one invoice after rollback, total=%d len=%d", total, len(list))
	}
}

func TestInvoiceVoid(t *testing.T) {
	orderSvc, invoiceSvc, _ := setupOrderSvc(t)
	ctx := context.Background()

	order, err := orderSvc.Create(ctx, sampleInput(true))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	line := order.Lines[0]
	doc, err := invoiceSvc.Create(ctx, service.CreateInvoiceDocInput{
		Lines: []service.CreateInvoiceDocLineInput{{
			SalesOrderID:     order.ID,
			SalesOrderLineID: line.ID,
			MaterialName:     line.MaterialName,
			Unit:             line.Unit,
			Quantity:         line.Quantity,
			UnitPrice:        line.UnitPrice,
			OrderNo:          order.OrderNo,
			OrderDate:        order.OrderDate,
			CustomerName:     order.CustomerName,
			WarehouseName:    order.WarehouseName,
			DeliveryType:     order.DeliveryType,
		}},
	})
	if err != nil {
		t.Fatalf("invoice: %v", err)
	}

	voided, err := invoiceSvc.Void(ctx, doc.ID)
	if err != nil {
		t.Fatalf("void: %v", err)
	}
	if voided.Status != "voided" || voided.VoidedAt == nil {
		t.Fatalf("voided state: %+v", voided)
	}

	reloaded, err := orderSvc.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("reload order: %v", err)
	}
	if reloaded.Lines[0].InvoiceDocID != nil {
		t.Fatalf("invoiceDocId should be cleared")
	}

	if _, err := invoiceSvc.Void(ctx, doc.ID); !errors.Is(err, service.ErrInvoiceAlreadyVoid) {
		t.Fatalf("want already void, got %v", err)
	}

	cands, err := orderSvc.ListInvoiceCandidates(ctx, nil)
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	if len(cands) != 1 {
		t.Fatalf("want 1 candidate after void, got %d", len(cands))
	}
}
