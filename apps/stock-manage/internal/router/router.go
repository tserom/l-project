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

	salesRepo := repository.NewSalesOrderRepository(db)
	salesSvc := service.NewSalesOrderService(db, salesRepo, opLogRepo)

	shipmentRepo := repository.NewSalesShipmentRepository(db)
	shipmentSvc := service.NewShipmentService(db, shipmentRepo, salesRepo, opLogRepo, centerClient)

	salesHandler := handler.NewSalesOrderHandler(salesSvc, shipmentSvc)
	shipmentHandler := handler.NewSalesShipmentHandler(shipmentSvc)

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

		sales := api.Group("/sales-orders")
		{
			sales.GET("", salesHandler.List)
			sales.GET("/:id/shipments", salesHandler.ListShipments)
			sales.POST("/:id/shipments", salesHandler.CreateShipment)
			sales.GET("/:id", salesHandler.Get)
			sales.POST("", salesHandler.Create)
			sales.PUT("/:id", salesHandler.Update)
			sales.DELETE("/:id", salesHandler.Delete)
			sales.POST("/:id/confirm", salesHandler.Confirm)
		}

		shipments := api.Group("/sales-shipments")
		{
			shipments.GET("", shipmentHandler.List)
			shipments.GET("/:id", shipmentHandler.Get)
			shipments.POST("", shipmentHandler.Create)
			shipments.PUT("/:id", shipmentHandler.Update)
			shipments.DELETE("/:id", shipmentHandler.Delete)
			shipments.POST("/:id/confirm", shipmentHandler.Confirm)
		}
	}

	return r
}
