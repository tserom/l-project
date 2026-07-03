package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-center/internal/handler"
	"github.com/tserom/l-project/apps/stock-center/internal/repository"
	"github.com/tserom/l-project/apps/stock-center/internal/service"
	"gorm.io/gorm"
)

// New builds the Gin engine with all routes registered.
func New(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	stockRepo := repository.NewStockRepository(db)
	stockSvc := service.NewStockService(stockRepo)

	healthHandler := handler.NewHealthHandler()
	stockHandler := handler.NewStockHandler(stockSvc)

	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/stocks", stockHandler.List)
		v1.GET("/stocks/by-sku", stockHandler.GetBySKU)
		v1.GET("/stocks/:id", stockHandler.Get)
		v1.POST("/stocks", stockHandler.Create)
		v1.PUT("/stocks/:id/quantity", stockHandler.Adjust)
	}

	return r
}
