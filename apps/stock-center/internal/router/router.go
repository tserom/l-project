package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-center/internal/handler"
	"gorm.io/gorm"
)

// New builds the Gin engine with all routes registered.
func New(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler()

	r.GET("/health", healthHandler.Check)

	return r
}
