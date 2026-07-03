package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-manage/internal/client/stockcenter"
	"github.com/tserom/l-project/apps/stock-manage/internal/config"
	"github.com/tserom/l-project/apps/stock-manage/internal/handler"
	"github.com/tserom/l-project/apps/stock-manage/internal/repository"
	"github.com/tserom/l-project/apps/stock-manage/internal/service"
	"gorm.io/gorm"
)

// New builds the Gin engine with all routes registered.
func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	stockCenterClient := stockcenter.NewClient(cfg.StockCenterBaseURL)
	logRepo := repository.NewOperationLogRepository(db)
	inventorySvc := service.NewInventoryService(stockCenterClient, logRepo)

	healthHandler := handler.NewHealthHandler()
	inventoryHandler := handler.NewInventoryHandler(inventorySvc)

	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/inventory", inventoryHandler.List)
		v1.GET("/inventory/query", inventoryHandler.Query)
		v1.POST("/inventory/inbound", inventoryHandler.Inbound)
	}

	return r
}
