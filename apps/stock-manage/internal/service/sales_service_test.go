package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSalesOrderService(t *testing.T, center *recordingCenter) (*service.SalesOrderService, *service.ShipmentService, *gorm.DB) {
	t.Helper()

	if center == nil {
		center = &recordingCenter{}
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.DocSequence{},
		&model.SalesOrder{},
		&model.SalesOrderLine{},
		&model.SalesShipment{},
		&model.SalesShipmentLine{},
		&model.StockOperationLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	salesRepo := repository.NewSalesOrderRepository(db)
	shipmentRepo := repository.NewSalesShipmentRepository(db)
	logRepo := repository.NewOperationLogRepository(db)
	salesSvc := service.NewSalesOrderService(db, salesRepo, logRepo)
	shipmentSvc := service.NewShipmentService(db, shipmentRepo, salesRepo, logRepo, center)
	return salesSvc, shipmentSvc, db
}

func sampleSalesLine() service.SalesOrderLineInput {
	return service.SalesOrderLineInput{
		MaterialID: 1,
		BatchID:    2,
		WeightKg:   decimal.NewFromInt(50),
		Quantity:   decimal.Zero,
	}
}

func createConfirmedSalesOrder(t *testing.T, salesSvc *service.SalesOrderService) *model.SalesOrder {
	t.Helper()
	ctx := context.Background()

	order, err := salesSvc.Create(ctx, service.CreateSalesOrderInput{
		CustomerName: "ACME Corp",
		Operator:     "alice",
		Lines:        []service.SalesOrderLineInput{sampleSalesLine()},
	})
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	confirmed, err := salesSvc.Confirm(ctx, order.ID)
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}
	return confirmed
}

func TestSalesOrder_ConfirmDoesNotCallCenter(t *testing.T) {
	center := &recordingCenter{}
	salesSvc, _, _ := setupSalesOrderService(t, center)
	ctx := context.Background()

	order, err := salesSvc.Create(ctx, service.CreateSalesOrderInput{
		CustomerName: "ACME Corp",
		Operator:     "alice",
		Lines:        []service.SalesOrderLineInput{sampleSalesLine()},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	confirmed, err := salesSvc.Confirm(ctx, order.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != model.DocStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", confirmed.Status)
	}
	if center.inboundCalls != 0 || center.outboundCalls != 0 {
		t.Fatalf("expected no center calls, inbound=%d outbound=%d", center.inboundCalls, center.outboundCalls)
	}
}

func TestShipment_ConfirmCallsOutboundWithSaleRefType(t *testing.T) {
	center := &recordingCenter{}
	salesSvc, shipmentSvc, _ := setupSalesOrderService(t, center)
	ctx := context.Background()

	order := createConfirmedSalesOrder(t, salesSvc)

	shipment, err := shipmentSvc.CreateFromSalesOrder(ctx, order.ID, service.CreateShipmentInput{
		Operator:  "bob",
		Warehouse: "WH-A",
	})
	if err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if shipment.Status != model.DocStatusDraft {
		t.Fatalf("expected draft shipment, got %s", shipment.Status)
	}
	if center.outboundCalls != 0 {
		t.Fatalf("expected no outbound before confirm, got %d", center.outboundCalls)
	}

	confirmed, err := shipmentSvc.Confirm(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("confirm shipment: %v", err)
	}
	if confirmed.Status != model.DocStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", confirmed.Status)
	}
	if center.outboundCalls != 1 {
		t.Fatalf("expected 1 outbound call, got %d", center.outboundCalls)
	}
	if center.lastOutbound.RefType != "sale" {
		t.Fatalf("refType: got %q want sale", center.lastOutbound.RefType)
	}
	if center.lastOutbound.RefNo != shipment.DocNo {
		t.Fatalf("refNo: got %q want %q", center.lastOutbound.RefNo, shipment.DocNo)
	}
}

func TestShipment_CreateFromOrderWithPartialLines(t *testing.T) {
	center := &recordingCenter{}
	salesSvc, shipmentSvc, _ := setupSalesOrderService(t, center)
	ctx := context.Background()

	order := createConfirmedSalesOrder(t, salesSvc)

	shipment, err := shipmentSvc.CreateFromSalesOrder(ctx, order.ID, service.CreateShipmentInput{
		Operator: "bob",
		Lines: []service.OrderLineInput{
			{
				MaterialID: 1,
				BatchID:    2,
				Warehouse:  "WH-B",
				WeightKg:   decimal.NewFromInt(25),
				Quantity:   decimal.Zero,
			},
		},
	})
	if err != nil {
		t.Fatalf("create partial shipment: %v", err)
	}
	if len(shipment.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(shipment.Lines))
	}
	if !shipment.Lines[0].WeightKg.Equal(decimal.NewFromInt(25)) {
		t.Fatalf("unexpected weight: %s", shipment.Lines[0].WeightKg)
	}
}

func TestShipment_RequiresConfirmedSalesOrder(t *testing.T) {
	center := &recordingCenter{}
	salesSvc, shipmentSvc, _ := setupSalesOrderService(t, center)
	ctx := context.Background()

	order, err := salesSvc.Create(ctx, service.CreateSalesOrderInput{
		CustomerName: "ACME Corp",
		Operator:     "alice",
		Lines:        []service.SalesOrderLineInput{sampleSalesLine()},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = shipmentSvc.CreateFromSalesOrder(ctx, order.ID, service.CreateShipmentInput{
		Operator:  "bob",
		Warehouse: "WH-A",
	})
	if !errors.Is(err, service.ErrSalesOrderNotConfirmed) {
		t.Fatalf("expected ErrSalesOrderNotConfirmed, got %v", err)
	}
}
