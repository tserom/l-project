package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type recordingCenter struct {
	inboundCalls  int
	outboundCalls int
	inboundErr    error
	outboundErr   error
	lastInbound   stockcenter.StockMovementInput
	lastOutbound  stockcenter.StockMovementInput
}

func (m *recordingCenter) InboundStock(_ context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error) {
	m.inboundCalls++
	m.lastInbound = input
	if m.inboundErr != nil {
		return nil, m.inboundErr
	}
	return &stockcenter.StockBalance{ID: 1}, nil
}

func (m *recordingCenter) OutboundStock(_ context.Context, input stockcenter.StockMovementInput) (*stockcenter.StockBalance, error) {
	m.outboundCalls++
	m.lastOutbound = input
	if m.outboundErr != nil {
		return nil, m.outboundErr
	}
	return &stockcenter.StockBalance{ID: 1}, nil
}

func setupInboundService(t *testing.T, center service.StockMover) (*service.InboundOrderService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.DocSequence{},
		&model.InboundOrder{},
		&model.InboundOrderLine{},
		&model.StockOperationLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewInboundOrderRepository(db)
	logRepo := repository.NewOperationLogRepository(db)
	svc := service.NewInboundOrderService(db, repo, logRepo, center)
	return svc, db
}

func sampleLine() service.OrderLineInput {
	return service.OrderLineInput{
		MaterialID: 1,
		BatchID:    2,
		Warehouse:  "WH-A",
		WeightKg:   decimal.NewFromInt(100),
		Quantity:   decimal.Zero,
	}
}

func TestInboundOrder_CreateRequiresOperator(t *testing.T) {
	svc, _ := setupInboundService(t, &recordingCenter{})
	ctx := context.Background()

	_, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Lines: []service.OrderLineInput{sampleLine()},
	})
	if !errors.Is(err, service.ErrOperatorRequired) {
		t.Fatalf("expected ErrOperatorRequired, got %v", err)
	}
}

func TestInboundOrder_CreateRequiresLines(t *testing.T) {
	svc, _ := setupInboundService(t, &recordingCenter{})
	ctx := context.Background()

	_, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Operator: "alice",
	})
	if !errors.Is(err, service.ErrLinesRequired) {
		t.Fatalf("expected ErrLinesRequired, got %v", err)
	}
}

func TestInboundOrder_ConfirmCallsCenterAndLogs(t *testing.T) {
	center := &recordingCenter{}
	svc, db := setupInboundService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Operator: "alice",
		Lines:    []service.OrderLineInput{sampleLine()},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.Status != model.DocStatusDraft {
		t.Fatalf("expected draft, got %s", order.Status)
	}

	confirmed, err := svc.Confirm(ctx, order.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != model.DocStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", confirmed.Status)
	}
	if center.inboundCalls != 1 {
		t.Fatalf("expected 1 inbound call, got %d", center.inboundCalls)
	}
	if center.lastInbound.RefType != "inbound" {
		t.Fatalf("refType: got %q", center.lastInbound.RefType)
	}
	if center.lastInbound.RefNo != order.DocNo {
		t.Fatalf("refNo: got %q want %q", center.lastInbound.RefNo, order.DocNo)
	}

	var logs []model.StockOperationLog
	if err := db.Find(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].Action != "confirm" || logs[0].DocType != "inbound" {
		t.Fatalf("unexpected log: %+v", logs[0])
	}
}

func TestInboundOrder_CannotConfirmTwice(t *testing.T) {
	center := &recordingCenter{}
	svc, _ := setupInboundService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Operator: "alice",
		Lines:    []service.OrderLineInput{sampleLine()},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := svc.Confirm(ctx, order.ID); err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	if _, err := svc.Confirm(ctx, order.ID); !errors.Is(err, service.ErrAlreadyConfirmed) {
		t.Fatalf("expected ErrAlreadyConfirmed, got %v", err)
	}
	if center.inboundCalls != 1 {
		t.Fatalf("expected only one center call, got %d", center.inboundCalls)
	}
}

func TestInboundOrder_CenterErrorPropagates(t *testing.T) {
	center := &recordingCenter{
		inboundErr: errors.New("stock-center error: insufficient stock"),
	}
	svc, _ := setupInboundService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Operator: "alice",
		Lines:    []service.OrderLineInput{sampleLine()},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = svc.Confirm(ctx, order.ID)
	var centerErr *service.CenterError
	if !errors.As(err, &centerErr) {
		t.Fatalf("expected CenterError, got %v", err)
	}
	if centerErr.Message != "insufficient stock" {
		t.Fatalf("message: got %q", centerErr.Message)
	}

	stillDraft, err := svc.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if stillDraft.Status != model.DocStatusDraft {
		t.Fatalf("expected draft after failed confirm, got %s", stillDraft.Status)
	}
}

func TestInboundOrder_ConfirmViaHTTPCenterClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/stocks/inbound" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data": map[string]interface{}{
				"id":         1,
				"materialId": 1,
				"batchId":    2,
				"warehouse":  "WH-A",
				"weightKg":   "100",
				"quantity":   "0",
			},
		})
	}))
	defer server.Close()

	center := stockcenter.NewClient(server.URL)
	svc, _ := setupInboundService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateInboundOrderInput{
		Operator: "bob",
		Lines:    []service.OrderLineInput{sampleLine()},
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
}

func setupOutboundService(t *testing.T, center service.StockMover) (*service.OutboundOrderService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.DocSequence{},
		&model.OutboundOrder{},
		&model.OutboundOrderLine{},
		&model.StockOperationLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewOutboundOrderRepository(db)
	logRepo := repository.NewOperationLogRepository(db)
	svc := service.NewOutboundOrderService(db, repo, logRepo, center)
	return svc, db
}

func TestOutboundOrder_ConfirmCallsCenter(t *testing.T) {
	center := &recordingCenter{}
	svc, _ := setupOutboundService(t, center)
	ctx := context.Background()

	order, err := svc.Create(ctx, service.CreateOutboundOrderInput{
		Operator: "alice",
		Lines:    []service.OrderLineInput{sampleLine()},
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
	if center.outboundCalls != 1 {
		t.Fatalf("expected 1 outbound call, got %d", center.outboundCalls)
	}
	if center.lastOutbound.RefType != "outbound" {
		t.Fatalf("refType: got %q", center.lastOutbound.RefType)
	}
}
