package service_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type orderedCenter struct {
	calls []string
}

func (m *orderedCenter) InboundStock(_ context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error) {
	m.calls = append(m.calls, "inbound:"+input.RefType+":"+input.WeightKg+":"+input.Quantity)
	return &stockcenter.StockBalance{ID: 1}, nil
}

func (m *orderedCenter) OutboundStock(_ context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error) {
	m.calls = append(m.calls, "outbound:"+input.RefType+":"+input.WeightKg)
	return &stockcenter.StockBalance{ID: 1}, nil
}

func setupProcessingService(t *testing.T, center service.StockMover) (*service.ProcessingOrderService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.DocSequence{},
		&model.ProcessingOrder{},
		&model.ProcessingPickLine{},
		&model.ProcessingFinishLine{},
		&model.StockOperationLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewProcessingOrderRepository(db)
	logRepo := repository.NewOperationLogRepository(db)
	svc := service.NewProcessingOrderService(db, repo, logRepo, center)
	return svc, db
}

func TestProcessingOrder_ConfirmOutboundThenInboundWithLoss(t *testing.T) {
	center := &orderedCenter{}
	svc, db := setupProcessingService(t, center)
	ctx := context.Background()

	finishWeight := decimal.NewFromInt(42)
	order, err := svc.Create(ctx, service.CreateProcessingOrderInput{
		Operator: "alice",
		PickLines: []service.ProcessingPickLineInput{
			{
				MaterialID: 1,
				BatchID:    10,
				Warehouse:  "WH-RM",
				WeightKg:   decimal.NewFromInt(50),
			},
		},
		FinishLines: []service.ProcessingFinishLineInput{
			{
				MaterialID: 2,
				BatchID:    20,
				Warehouse:  "WH-FG",
				Quantity:   decimal.NewFromInt(10),
				WeightKg:   &finishWeight,
			},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	confirmed, err := svc.Confirm(ctx, order.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != model.DocStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", confirmed.Status)
	}

	expectedLoss := decimal.NewFromInt(8) // 50 - 42
	if !confirmed.LossWeightKg.Equal(expectedLoss) {
		t.Fatalf("lossWeightKg: got %s want %s", confirmed.LossWeightKg, expectedLoss)
	}

	if len(center.calls) != 2 {
		t.Fatalf("expected 2 center calls, got %d: %v", len(center.calls), center.calls)
	}
	if center.calls[0] != "outbound:processing:50" {
		t.Fatalf("first call: got %q", center.calls[0])
	}
	if center.calls[1] != "inbound:processing:42:10" {
		t.Fatalf("second call: got %q", center.calls[1])
	}

	var logs []model.StockOperationLog
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(logs) != 1 || logs[0].DocType != "processing" || logs[0].Action != "confirm" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestProcessingOrder_ConfirmLossWithoutFinishWeight(t *testing.T) {
	center := &orderedCenter{}
	svc, _ := setupProcessingService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateProcessingOrderInput{
		Operator: "bob",
		PickLines: []service.ProcessingPickLineInput{
			{
				MaterialID: 1,
				BatchID:    10,
				Warehouse:  "WH-RM",
				WeightKg:   decimal.NewFromInt(50),
			},
		},
		FinishLines: []service.ProcessingFinishLineInput{
			{
				MaterialID: 2,
				BatchID:    20,
				Warehouse:  "WH-FG",
				Quantity:   decimal.NewFromInt(10),
			},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	confirmed, err := svc.Confirm(ctx, order.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}

	expectedLoss := decimal.NewFromInt(50)
	if !confirmed.LossWeightKg.Equal(expectedLoss) {
		t.Fatalf("lossWeightKg: got %s want %s", confirmed.LossWeightKg, expectedLoss)
	}
	if center.calls[1] != "inbound:processing:0:10" {
		t.Fatalf("inbound call: got %q", center.calls[1])
	}
}
