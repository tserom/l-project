package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/stock-center/pkg/response"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct{}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check returns service health status.
func (h *HealthHandler) Check(c *gin.Context) {
	response.OK(c, gin.H{
		"service": "stock-center",
		"status":  "up",
	})
}
