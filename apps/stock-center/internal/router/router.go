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

	healthHandler := handler.NewHealthHandler()

	materialRepo := repository.NewMaterialRepository(db)
	materialSvc := service.NewMaterialService(materialRepo)
	materialHandler := handler.NewMaterialHandler(materialSvc)

	batchRepo := repository.NewBatchRepository(db)
	batchSvc := service.NewBatchService(batchRepo, materialRepo)
	batchHandler := handler.NewBatchHandler(batchSvc)

	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/materials", materialHandler.List)
		v1.GET("/materials/:id", materialHandler.Get)
		v1.POST("/materials", materialHandler.Create)
		v1.PUT("/materials/:id", materialHandler.Update)

		v1.GET("/batches", batchHandler.List)
		v1.GET("/batches/:id", batchHandler.Get)
		v1.POST("/batches", batchHandler.Create)
		v1.PUT("/batches/:id", batchHandler.Update)
	}

	return r
}
