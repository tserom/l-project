package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tserom/l-project/apps/sales-manage/internal/model"
	"github.com/tserom/l-project/apps/sales-manage/internal/repository"
	"github.com/tserom/l-project/apps/sales-manage/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.SalesOrder{},
		&model.SalesOrderLine{},
		&model.InvoiceDoc{},
		&model.InvoiceDocLine{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func setupOrderSvc(t *testing.T) (*service.SalesOrderService, *service.InvoiceDocService, *gorm.DB) {
	t.Helper()
	db := setupDB(t)
	orderRepo := repository.NewSalesOrderRepository(db)
	invoiceRepo := repository.NewInvoiceDocRepository(db)
	return service.NewSalesOrderService(db, orderRepo),
		service.NewInvoiceDocService(db, invoiceRepo, orderRepo),
		db
}

func sampleInput(needInvoice bool) service.SalesOrderInput {
	return service.SalesOrderInput{
		OrderNo:       "SO001",
		OrderDate:     "2026-08-08",
		CustomerName:  "客户A",
		WarehouseName: "中大慧科",
		DeliveryType:  "提货",
		Lines: []service.SalesOrderLineInput{
			{
				MaterialName: "304板",
				Spec:         "1.0",
				Unit:         "kg",
				Quantity:     10,
				UnitPrice:    2.5,
				NeedInvoice:  needInvoice,
			},
		},
	}
}

func TestCreateAndValidate(t *testing.T) {
	svc, _, _ := setupOrderSvc(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, service.SalesOrderInput{
		CustomerName: "",
		Lines:        sampleInput(false).Lines,
	})
	if !errors.Is(err, service.ErrValidation) {
		t.Fatalf("want validation, got %v", err)
	}

	order, err := svc.Create(ctx, sampleInput(false))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.TotalAmount != 25 {
		t.Fatalf("totalAmount=%v", order.TotalAmount)
	}
	if len(order.Lines) != 1 || order.Lines[0].Amount != 25 {
		t.Fatalf("line amount unexpected: %+v", order.Lines)
	}
}

func TestDeleteBlockedWhenInvoiced(t *testing.T) {
	orderSvc, invoiceSvc, _ := setupOrderSvc(t)
	ctx := context.Background()

	order, err := orderSvc.Create(ctx, sampleInput(true))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	line := order.Lines[0]
	_, err = invoiceSvc.Create(ctx, service.CreateInvoiceDocInput{
		Lines: []service.CreateInvoiceDocLineInput{{
			SalesOrderID:     order.ID,
			SalesOrderLineID: line.ID,
			MaterialName:     line.MaterialName,
			Spec:             line.Spec,
			Unit:             line.Unit,
			Quantity:         line.Quantity,
			UnitPrice:        line.UnitPrice,
			Amount:           line.Amount,
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

	if err := orderSvc.Delete(ctx, order.ID); !errors.Is(err, service.ErrOrderHasInvoice) {
		t.Fatalf("want ErrOrderHasInvoice, got %v", err)
	}
}

func TestUpdatePreservesInvoiceDocID(t *testing.T) {
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

	updated, err := orderSvc.Update(ctx, order.ID, service.SalesOrderInput{
		OrderNo:       order.OrderNo,
		OrderDate:     order.OrderDate,
		CustomerName:  "客户B",
		WarehouseName: order.WarehouseName,
		DeliveryType:  order.DeliveryType,
		Lines: []service.SalesOrderLineInput{{
			ID:           line.ID,
			MaterialName: line.MaterialName,
			Unit:         line.Unit,
			Quantity:     11,
			UnitPrice:    2.5,
			NeedInvoice:  true,
			// omit InvoiceDocID
		}},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Lines[0].InvoiceDocID == nil || *updated.Lines[0].InvoiceDocID != doc.ID {
		t.Fatalf("invoiceDocId not preserved: %+v", updated.Lines[0].InvoiceDocID)
	}
}
