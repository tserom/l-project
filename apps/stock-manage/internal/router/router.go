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

	healthHandler := handler.NewHealthHandler()
	r.GET("/health", healthHandler.Check)

	centerClient := stockcenter.NewClient(cfg.StockCenterBaseURL)
	opLogRepo := repository.NewOperationLogRepository(db)

	inboundRepo := repository.NewInboundOrderRepository(db)
	inboundSvc := service.NewInboundOrderService(db, inboundRepo, opLogRepo, centerClient)
	inboundHandler := handler.NewInboundOrderHandler(inboundSvc)

	outboundRepo := repository.NewOutboundOrderRepository(db)
	outboundSvc := service.NewOutboundOrderService(db, outboundRepo, opLogRepo, centerClient)
	outboundHandler := handler.NewOutboundOrderHandler(outboundSvc)

	api := r.Group("/api/v1")
	{
		inbound := api.Group("/inbound-orders")
		{
			inbound.GET("", inboundHandler.List)
			inbound.GET("/:id", inboundHandler.Get)
			inbound.POST("", inboundHandler.Create)
			inbound.PUT("/:id", inboundHandler.Update)
			inbound.DELETE("/:id", inboundHandler.Delete)
			inbound.POST("/:id/confirm", inboundHandler.Confirm)
		}

		outbound := api.Group("/outbound-orders")
		{
			outbound.GET("", outboundHandler.List)
			outbound.GET("/:id", outboundHandler.Get)
			outbound.POST("", outboundHandler.Create)
			outbound.PUT("/:id", outboundHandler.Update)
			outbound.DELETE("/:id", outboundHandler.Delete)
			outbound.POST("/:id/confirm", outboundHandler.Confirm)
		}
	}

	return r
}
