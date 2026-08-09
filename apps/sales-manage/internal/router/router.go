package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/sales-manage/internal/handler"
	"github.com/tserom/l-project/apps/sales-manage/internal/repository"
	"github.com/tserom/l-project/apps/sales-manage/internal/service"
	"gorm.io/gorm"
)

// New builds the Gin engine for sales-manage.
func New(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	orderRepo := repository.NewSalesOrderRepository(db)
	invoiceRepo := repository.NewInvoiceDocRepository(db)
	orderSvc := service.NewSalesOrderService(db, orderRepo)
	invoiceSvc := service.NewInvoiceDocService(db, invoiceRepo, orderRepo)

	health := handler.NewHealthHandler()
	orderHandler := handler.NewSalesOrderHandler(orderSvc)
	invoiceHandler := handler.NewInvoiceDocHandler(invoiceSvc)

	r.GET("/health", health.Check)

	api := r.Group("/api/v1")
	sale := api.Group("/sale")
	{
		orders := sale.Group("/orders")
		{
			orders.GET("", orderHandler.List)
			orders.POST("", orderHandler.Create)
			// Static path before :id
			orders.GET("/invoice-candidates", orderHandler.ListInvoiceCandidates)
			orders.GET("/:id", orderHandler.Get)
			orders.PUT("/:id", orderHandler.Update)
			orders.DELETE("/:id", orderHandler.Delete)
		}

		invoices := sale.Group("/invoice-docs")
		{
			invoices.GET("", invoiceHandler.List)
			invoices.POST("", invoiceHandler.Create)
			invoices.GET("/:id", invoiceHandler.Get)
			invoices.POST("/:id/void", invoiceHandler.Void)
		}
	}

	return r
}
