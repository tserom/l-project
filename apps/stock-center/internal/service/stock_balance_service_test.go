package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"github.com/tserom/l-project/apps/stock-center/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStockBalanceService(t *testing.T) *service.StockBalanceService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.StockBalance{}, &model.StockLedger{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	balanceRepo := repository.NewStockBalanceRepository(db)
	ledgerRepo := repository.NewStockLedgerRepository(db)
	return service.NewStockBalanceService(balanceRepo, ledgerRepo)
}

func TestInboundThenOutbound_LeavesRemainingBalance(t *testing.T) {
	svc := setupStockBalanceService(t)
	ctx := context.Background()

	inbound := service.StockMovementInput{
		MaterialID: 1,
		BatchID:    1,
		WeightKg:   decimal.NewFromInt(100),
		Quantity:   decimal.Zero,
		RefNo:      "IN-001",
	}
	if _, err := svc.Inbound(ctx, inbound); err != nil {
		t.Fatalf("inbound: %v", err)
	}

	outbound := service.StockMovementInput{
		MaterialID: 1,
		BatchID:    1,
		WeightKg:   decimal.NewFromInt(30),
		Quantity:   decimal.Zero,
		RefNo:      "OUT-001",
	}
	if _, err := svc.Outbound(ctx, outbound); err != nil {
		t.Fatalf("outbound: %v", err)
	}

	balance, err := svc.QueryStock(ctx, 1, 1, "")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !balance.WeightKg.Equal(decimal.NewFromInt(70)) {
		t.Fatalf("expected 70kg, got %s", balance.WeightKg)
	}

	entries, total, err := svc.ListLedger(ctx, 1, 20, nil)
	if err != nil {
		t.Fatalf("list ledger: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 ledger entries, got %d", total)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 ledger entries in list, got %d", len(entries))
	}
}

func TestOutbound_OverBalanceFails(t *testing.T) {
	svc := setupStockBalanceService(t)
	ctx := context.Background()

	inbound := service.StockMovementInput{
		MaterialID: 1,
		BatchID:    1,
		WeightKg:   decimal.NewFromInt(100),
		Quantity:   decimal.Zero,
		RefNo:      "IN-002",
	}
	if _, err := svc.Inbound(ctx, inbound); err != nil {
		t.Fatalf("inbound: %v", err)
	}

	outbound := service.StockMovementInput{
		MaterialID: 1,
		BatchID:    1,
		WeightKg:   decimal.NewFromInt(150),
		Quantity:   decimal.Zero,
		RefNo:      "OUT-002",
	}
	_, err := svc.Outbound(ctx, outbound)
	if err == nil {
		t.Fatal("expected outbound over balance to fail")
	}
	if !errors.Is(err, repository.ErrInsufficientWeight) {
		t.Fatalf("expected ErrInsufficientWeight, got %v", err)
	}

	balance, err := svc.QueryStock(ctx, 1, 1, "")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !balance.WeightKg.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected balance unchanged at 100kg, got %s", balance.WeightKg)
	}
}
