package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-manage/internal/config"
	"github.com/tserom/l-project/apps/stock-manage/internal/handler"
	"gorm.io/gorm"
)

// New builds the Gin engine with all routes registered.
func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	_ = cfg
	_ = db

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	healthHandler := handler.NewHealthHandler()
	r.GET("/health", healthHandler.Check)

	return r
}
