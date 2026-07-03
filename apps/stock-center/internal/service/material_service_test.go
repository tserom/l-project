package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tserom/l-project/apps/stock-center/internal/model"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"github.com/tserom/l-project/apps/stock-center/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateMaterial_DuplicateCodeRejected(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Material{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.NewMaterialRepository(db)
	svc := service.NewMaterialService(repo)

	input := service.CreateMaterialInput{
		MaterialCode: "TEST-001",
		Grade:        "304",
		Form:         model.FormPlate,
		Spec:         "3mm",
		PrimaryUnit:  model.UnitKg,
		MaterialType: model.TypeRaw,
	}

	if _, err := svc.CreateMaterial(context.Background(), input); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = svc.CreateMaterial(context.Background(), input)
	if err == nil {
		t.Fatal("expected duplicate material code error")
	}
	if !errors.Is(err, service.ErrDuplicateMaterialCode) {
		t.Fatalf("expected ErrDuplicateMaterialCode, got %v", err)
	}
}
