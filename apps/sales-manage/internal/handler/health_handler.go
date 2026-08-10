package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tserom/l-project/apps/sales-manage/pkg/response"
)

// HealthHandler serves liveness checks.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

// Check handles GET /health.
func (h *HealthHandler) Check(c *gin.Context) {
	response.OK(c, gin.H{"service": "sales-manage", "status": "up"})
}
