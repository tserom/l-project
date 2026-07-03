package docno_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tserom/l-project/apps/stock-manage/internal/model"
	"github.com/tserom/l-project/apps/stock-manage/internal/pkg/docno"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGenerate_DailySequence(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.DocSequence{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	today := time.Now().Format("20060102")

	first, err := docno.Generate(ctx, db, "IN")
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}
	expectedFirst := "IN" + today + "0001"
	if first != expectedFirst {
		t.Fatalf("first doc no: got %q want %q", first, expectedFirst)
	}

	second, err := docno.Generate(ctx, db, "IN")
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}
	expectedSecond := "IN" + today + "0002"
	if second != expectedSecond {
		t.Fatalf("second doc no: got %q want %q", second, expectedSecond)
	}

	out, err := docno.Generate(ctx, db, "OUT")
	if err != nil {
		t.Fatalf("outbound generate: %v", err)
	}
	if !strings.HasPrefix(out, "OUT"+today) {
		t.Fatalf("outbound prefix/date: got %q", out)
	}
	if !strings.HasSuffix(out, "0001") {
		t.Fatalf("outbound seq resets per prefix: got %q", out)
	}
}
